package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"golang.org/x/term"
)

const (
	mStart = iota
	mEnd
	mBoth
)

type ExitError struct {
	msg string
}

func (e *ExitError) SetError(msg error) {
	e.msg = msg.Error()
}

func (e *ExitError) Error() string {
	return e.msg
}

var (
	initError error
	exitError *ExitError

	stdinFd int
	state   *term.State
	logFile *os.File
)

func init() {
	setupLogging()
	stdinFd = int(os.Stdout.Fd())
	state, initError = term.MakeRaw(stdinFd)
	exitError = new(ExitError)
	fmt.Fprint(os.Stdout, "\x1b[2J")   // clear the screen
	fmt.Fprint(os.Stdout, "\x1b[H")    // CUP - get the cursor UP (top left)
	fmt.Fprint(os.Stdout, "\x1b[?25l") // hide the cursor
}

func setupLogging() {
	// minimal logging
	var logName string
	//if os.Getenv["LOG_FILE"] == "" {
	//logName = "/dev/null"
	//} else {
	//logName = os.Getenv["LOG_FILE"]
	//}
	logName = "debug/chip8.log"
	logFile, err := os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

func shutDown() {
	term.Restore(stdinFd, state)
	fmt.Fprint(os.Stdout, "\x1b[2J")   // clear the screen
	fmt.Fprint(os.Stdout, "\x1b[H")    // CUP - get the cursor UP (top left)
	fmt.Fprint(os.Stdout, "\x1b[?25h") // display the cursor
	if exitError != nil {
		fmt.Println(exitError)
		log.Println(exitError)
	}
	logFile.Close()
}

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		shutDown()
		signal.Stop(signalChan)
	}()

	romPath := flag.String("rom", "", "path to a valid CHIP-8 ROM")
	flag.Parse()

	chip := NewChip()
	err := chip.LoadROM(*romPath)
	if err != nil {
		fmt.Println("error, couldn't start the emulator:", err)
		return
	}
	log.Println("starting the chip-8 emulator")
	completed := make(chan bool)
	go func() {
		chip.Run()
		completed <- true
	}()

	select {
	case <-signalChan:
		exitError = &ExitError{"shutting down (ctr+c pressed)"}
	case <-completed:
		exitError = &ExitError{"rom completed, shutting down"}
	}
}
