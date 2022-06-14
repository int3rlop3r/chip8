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
	pc       uint16
	sp       uint16
	stack    [16]uint16
	delay    uint8
	sound    uint8
	keyboard [16]byte
	display  [32][63]byte
}

func NewChip() *Chip {
	var chip Chip
	chip.pc = startOff
	return &chip
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

func (c *Chip) Run() error {
	var opcode uint16
	//for ; c.pc < memSize-1; c.pc += 2 {
	for i := 0; i < 10; i++ { // for now lets just disassemble 10 opcodes
		opcode = uint16(c.mem[c.pc])<<8 | uint16(c.mem[c.pc+1])
		fmt.Printf("pc:%02x, mem:%02x, shift:%02x, opcode: %04x\n",
			c.pc, c.mem[c.pc], uint16(c.mem[c.pc])<<8, opcode)
		//c.pc += 2
		switch opcode & 0xF000 {
		case 0x0000:
			switch opcode & 0x00FF {
			case 0x00E0:
				fmt.Println("CLS")
			case 0x00EE:
				fmt.Println("RET")
			default:
				return fmt.Errorf("unknown operation: %04x", opcode)
			}
			c.pc += 2
		case 0x1000: // jmp addr
			c.pc = opcode & 0x0FFF
		case 0x2000: // call addr
			c.stack[c.sp] = c.pc
			c.sp++
			c.pc = opcode & 0x0FFF
		case 0x3000: // SE Vx, byte
			x := opcode & 0x0F00
			kk := byte(opcode & 0x00FF)
			if c.v[x] == kk { // @TODO: check if x == 'F' ?
				c.pc += 2
			}
			c.pc += 2
		case 0x4000: // SNE Vx, byte
			x := opcode & 0x0F00
			kk := byte(opcode & 0x00FF)
			if c.v[x] != kk { // @TODO: check if x == 'F' ?
				c.pc += 2
			}
			c.pc += 2
		case 0x5000: // SE Vx, Vy
			x := opcode & 0x0F00
			y := opcode & 0x00F0
			if c.v[x] == c.v[y] { // @TODO: check if x == 'F' ?
				c.pc += 2
			}
			c.pc += 2
		case 0x6000: // LD Vx, byte
			x := opcode & 0x0F00
			kk := byte(opcode & 0x00FF)
			c.v[x] = kk
			c.pc += 2
		case 0x7000: // Add Vx, byte
			x := opcode & 0x0F00
			kk := byte(opcode & 0x00FF)
			c.v[x] += kk
			c.pc += 2
		case 0x8000: // LD Vx, Vy
			// @WIP
			x := opcode & 0x0F00
			y := opcode & 0x00F0
			if c.v[x] == c.v[y] { // @TODO: check if x == 'F' ?
				c.pc += 2
			}
			c.pc += 2
		case 0xA000:
			c.i = opcode & 0x0FFF
		default:
			return fmt.Errorf("unknown operation: %04x", opcode)
		}
	}
}
