package chip

import (
	"time"
)

type Chip struct {
	memory     []byte   // 4096 bytes
	stack      []uint16 // 16 16-bit values
	registers  []byte   // 16 8-bit registers Vx (0-F)
	delays     []byte   // 2 8-bit registers for delay timers (decremented at 60Hz)
	sounds     []byte   // 2 8-bit registers for sound timers (decremented at 60Hz)
	keys       []byte
	i          uint16 // 16-bit register, only lowest 12 bit are used
	pc         uint16 // 16-bit program counter
	sp         byte   // 8-bit stack pointer
	currentKey byte   // currently pressed key
	drawFlag   bool   // should display be drawn
}

var keyMap = map[rune]byte{
	'1': 1,  // 1
	'2': 2,  // 2
	'3': 3,  // 3
	'4': 4,  // 4
	'Q': 5,  // Q
	'W': 6,  // W
	'E': 7,  // E
	'R': 8,  // R
	'A': 9,  // A
	'S': 10, // S
	'D': 11, // D
	'F': 12, // F
	'Z': 13, // Z
	'X': 14, // X
	'C': 15, // C
	'V': 16, // V
}

func New() *Chip {
	memory := make([]byte, 4096)
	fontSet := []byte{
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
	for i, b := range fontSet {
		memory[i] = b
	}

	return &Chip{
		memory:    memory,
		registers: make([]byte, 16),
		stack:     make([]uint16, 16),
		keys:      make([]byte, 16),
		drawFlag:  false,
		pc:        0x200,
		sp:        0,
	}
}

func (c *Chip) LoadRom(rom []byte) {
	for i, b := range rom {
		c.memory[0x200+i] = b
	}
}

func (c *Chip) KeyDown(char rune) {
	idx := keyMap[char]
	c.keys[idx] = 1
	c.currentKey = idx
}

func (c *Chip) KeyUp(char rune) {
	idx := keyMap[char]
	c.keys[idx] = 0
	c.currentKey = 0
}

func (c *Chip) cycle() {
	for {
		c.nextOp()

		if c.drawFlag {
			c.draw()
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (c *Chip) nextOp() {
	pc := c.pc
	opcode := uint16(c.memory[pc])<<8 | uint16(c.memory[pc+1])

	switch opcode & 0xF000 {
	// SYS addr
	case 0x0000:
		switch opcode & 0x0FFF {
		// CLS
		case 0x00E0:
			// clear display here
			break
		// RET
		case 0x0EE:
			// return from subroutine
			break
		}
	}
}

func (c *Chip) draw() {

}
