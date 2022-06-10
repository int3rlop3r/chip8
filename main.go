package main

import "fmt"

type Chip struct {
	//0x000 to 0x1FF
	mem      [4096]byte
	v        [16]byte
	i        uint16
	delay    uint8
	sound    uint8
	keyboard [16]byte
}

func main() {
	var chip Chip
	_ = chip.v
	fmt.Println("chip-8 emulator!")
}
