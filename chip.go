package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
)

const (
	memSize  = 0xFFF // fff == 4096
	startOff = 0x200
	stackOff = 0xFA0
)

type Chip struct {
	mem      [memSize]byte //0x000 to 0x1FF (511)
	v        [16]byte
	i        uint16
	pc       uint16
	sp       uint8
	delay    uint8
	sound    uint8
	keyboard [16]byte
	display  [32][63]byte
	stack    [24]uint16 // for now let's not use the emulated memory

	jmp bool // this flag is not part of the spec, used for iteration
}

func NewChip() *Chip {
	var chip Chip
	chip.pc = startOff
	//chip.sp = stackOff
	chip.loadFonts()
	return &chip
}

func (c *Chip) NextInstr() {
	// if the `pc` was manually set
	// don't increment by 2
	if !c.jmp {
		c.pc += 2
	} else {
		// reset value for next iteration
		c.jmp = !c.jmp
	}
}

func (c *Chip) SetJump(val uint16) {
	c.jmp = true
	c.pc = val
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

func (c *Chip) loadFonts() {
	fonts := [80]byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}
	for i, x := range fonts {
		c.mem[i] = x
	}
}

func (c *Chip) Run() error {
	var opcode uint16
	var cntr int
	for ; c.pc < memSize-1; c.NextInstr() {
		// for now lets just deal with 10 instructions
		//if cntr == 10 {
		//break
		//} else {
		//cntr++
		//}
		cntr++

		opcode = uint16(c.mem[c.pc])<<8 | uint16(c.mem[c.pc+1])
		fmt.Printf("pc:%02x, mem:%02x, shift:%02x, opcode: %04x\n",
			c.pc, c.mem[c.pc], uint16(c.mem[c.pc])<<8, opcode)

		switch opcode & 0xF000 {
		case 0x0000:
			switch opcode & 0x00FF {
			case 0x00E0:
				fmt.Println("CLS")
			case 0x00EE:
				c.SetJump(c.stack[c.sp])
				c.sp--
			default:
				return fmt.Errorf("unknown operation: %04x", opcode)
			}
		case 0x1000: // jmp addr
			c.SetJump(opcode & 0x0FFF)
		case 0x2000: // call addr
			c.sp++
			c.stack[c.sp] = c.pc
			c.SetJump(opcode & 0x0FFF)
		case 0x3000: // SE Vx, byte
			x := (opcode & 0x0F00) >> 8
			kk := byte(opcode & 0x00FF)
			if c.v[x] == kk { // @TODO: check if x == 'F' ?
				c.pc += 2
			}
		case 0x4000: // SNE Vx, byte
			x := (opcode & 0x0F00) >> 8
			kk := byte(opcode & 0x00FF)
			if c.v[x] != kk { // @TODO: check if x == 'F' ?
				c.pc += 2
			}
		case 0x5000: // SE Vx, Vy
			x := (opcode & 0x0F00) >> 8
			y := (opcode & 0x00F0) >> 4
			if c.v[x] == c.v[y] { // @TODO: check if x == 'F' ?
				c.pc += 2
			}
		case 0x6000: // LD Vx, byte
			x := (opcode & 0x0F00) >> 8
			kk := byte(opcode & 0x00FF)
			c.v[x] = kk
		case 0x7000: // Add Vx, byte
			x := (opcode & 0x0F00) >> 8
			kk := byte(opcode & 0x00FF)
			c.v[x] += kk
		case 0x8000:
			x := (opcode & 0x0F00) >> 8
			y := (opcode & 0x00F0) >> 4
			switch opcode & 0x000F {
			case 0x0000: // LD Vx, Vy
				c.v[x] = c.v[y]
			case 0x0001:
				c.v[x] |= c.v[y]
			case 0x0002:
				c.v[x] &= c.v[y]
			case 0x0003:
				c.v[x] ^= c.v[y]
			case 0x0004:
				// check if v[x] has enough space for v[y]
				// set the carry flag if there's no space
				if c.v[y] > (0xFF - c.v[x]) {
					c.v[0xF] = 1
				} else {
					c.v[0xF] = 0
				}
				c.v[x] += c.v[y]
			case 0x0005:
				if c.v[y] > c.v[x] {
					c.v[0xF] = 1
				} else {
					c.v[0xF] = 0
				}
				c.v[x] = c.v[x] - c.v[y]
			case 0x0006:
				c.v[0xF] = c.v[x] & 1
				c.v[x] /= 2
			case 0x0007:
				c.v[0xF] = c.v[x] & 1
				c.v[x] = c.v[y] - c.v[x]
			case 0x000E:
				c.v[0xF] = c.v[x] >> 8
				c.v[x] *= 2
			default:
				return fmt.Errorf("unknown operation: %04x", opcode)
			}
		case 0x9000: // SNE Vx, Vy
			x := (opcode & 0x0F00) >> 8
			y := (opcode & 0x00F0) >> 4
			if c.v[x] != c.v[y] { // @TODO: check if x == 'F' ?
				c.pc += 2
			}
		case 0xA000: // LD I, addr
			c.i = opcode & 0x0FFF
		case 0xB000: // JP V0, addr
			c.SetJump((opcode & 0x0FFF) + uint16(c.v[0]))
		case 0xC000: // RND Vx, byte
			x := (opcode & 0x0F00) >> 8
			c.v[x] = uint8(rand.Intn(256)) & uint8((opcode&0x00FF)>>8)
		case 0xD000: // DRW Vx, Vy, nibble
			//fmt.Println("Draw not implemented")
			x := (opcode & 0x0F00) >> 8
			y := (opcode & 0x00F0) >> 4
			n := opcode & 0x000F
			for i := 0; i < n; i++ {
			}
		case 0xE000:
			fmt.Println("Keypad not implemented")
		case 0xF000:
			fmt.Println("Not implemented")
		default:
			return fmt.Errorf("unknown operation: %04x", opcode)
		}
	}
	return nil
}
