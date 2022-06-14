package main

import (
	"flag"
	"fmt"
)

func main() {
	romPath := flag.String("rom", "", "path to a valid CHIP-8 ROM")
	flag.Parse()

	chip := NewChip()
	err := chip.LoadROM(*romPath)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println("chip-8 emulator!", len(chip.mem))
	fmt.Println("testing")
	chip.Run()
}
