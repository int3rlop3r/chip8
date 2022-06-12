package main

import (
	"fmt"
	"io"
	"os"
)

const (
	memSize  = 0xFFF // fff == 4096
	startOff = 0x200
)

type Chip struct {
	//0x000 to 0x1FF
	mem      [memSize]byte
	v        [16]byte
	i        uint16
	delay    uint8
	sound    uint8
	keyboard [16]byte
	display  [32][63]byte
}

func (c *Chip) LoadROM(path string) error {
	//fmt.Println("the rom path is:", path)
	var err error
	var rom *os.File
	rom, err = os.Open(path)
	if err != nil {
		return err
	}

	b := make([]byte, 1)
	for i := 0; i < memSize; i++ {
		cur := i + startOff
		_, err = rom.Read(b)
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("error while loading rom: %w", err)
		}
		c.mem[cur] = b[0]
	}

	if err == nil {
		return fmt.Errorf("ROM too large to fit in RAM")
	}

	return nil
}
