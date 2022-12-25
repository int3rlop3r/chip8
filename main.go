package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var logFile *os.File

func init() {
	// minimal logging
	var logName string
	if os.Getenv["LOG_FILE"] == "" {
		logName = "/dev/null"
	} else {
		logName = os.Getenv["LOG_FILE"]
	}
	logName = "debug/chip8.log"
	logFile, err := os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

func main() {
	defer logFile.Close()
	romPath := flag.String("rom", "", "path to a valid CHIP-8 ROM")
	flag.Parse()

	chip := NewChip()
	err := chip.LoadROM(*romPath)
	if err != nil {
		fmt.Println("error, couldn't start the emulator:", err)
		return
	}
	log.Println("starting the chip-8 emulator")
	chip.Run()
}
