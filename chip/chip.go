package chip

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/gdamore/tcell"
)

type Chip struct {
	screen     tcell.Screen
	memory     []byte   // 4096 bytes
	display    []byte   // 64x32 bytes (pixels)
	stack      []uint16 // 16 16-bit values
	v          []byte   // 16 8-bit registers Vx (0-F)
	dt         byte     // 8-bit register for delay timer (decremented at 60Hz)
	st         byte     // 8-bit register for sound timer (decremented at 60Hz)
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

func New(screen tcell.Screen) *Chip {
	memory := make([]byte, 4096)
	display := make([]byte, 64*32)
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
		screen:   screen,
		memory:   memory,
		display:  display,
		v:        make([]byte, 16),
		dt:       0,
		st:       0,
		stack:    make([]uint16, 16),
		keys:     make([]byte, 16),
		drawFlag: false,
		pc:       0x200,
		sp:       0,
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

func (c *Chip) Cycle() {
	for {
		err := c.nextOp()
		if err != nil {
			panic(err)
		}

		if c.drawFlag {
			c.draw()
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (c *Chip) nextOp() error {
	pc := c.pc
	opcode := uint16(c.memory[pc])<<8 | uint16(c.memory[pc+1])

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode & 0x0FFF {
		// CLS (00E0)
		case 0x00E0:
			// Clear the display
			c.clearDisplay()
			break
		// RET (00EE)
		case 0x00EE:
			// Return from subroutine
			c.pc = c.stack[c.sp]
			c.sp--
			break
		// SYS addr (0nnn)
		default:
			// Jump to a machine code routine at nnn. It is ignored by modern
			// interpreters
			break
		}
		break
	// JP addr (1nnn)
	case 0x1000:
		// Jump to location nnn
		c.pc = opcode & 0x0FFF
		break
	// CALL addr (2nnn)
	case 0x2000:
		// Call subroutine at nnn
		c.sp++
		c.stack[c.sp] = c.pc
		c.pc = opcode & 0x0FFF
		break
	// SE Vx, byte (3xkk)
	case 0x3000:
		// Skip next instruction if Vx = kk
		x := c.memory[pc] & 0x0F
		if c.v[x] == c.memory[pc+1] {
			c.pc += 2
		}
		break
	// SNE Vx, byte (4xkk)
	case 0x4000:
		// Skip next instruction if Vx != kk
		x := c.memory[pc] & 0x0F
		if c.v[x] != c.memory[pc+1] {
			c.pc += 2
		}
		break
	// SE Vx, Vy (5xy0)
	case 0x5000:
		// Skip next instruction if Vx == Vy
		x := c.memory[pc] & 0x0F
		y := (c.memory[pc+1] & 0xF0) >> 2
		if c.v[x] == c.v[y] {
			c.pc += 2
		}
		break
	// LD Vx, byte (6xkk)
	case 0x6000:
		// Set Vx = kk
		x := c.memory[pc] & 0x0F
		c.v[x] = c.memory[pc+1]
		break
	// ADD Vx, byte (7xkk)
	case 0x7000:
		// Set Vx = Vx + kk
		x := c.memory[pc] & 0x0F
		c.v[x] += c.memory[pc+1]
		break
	case 0x8000:
		x := c.memory[pc] & 0x0F
		y := c.memory[pc+1] & 0xF0 >> 2
		switch opcode & 0x000F {
		// LD Vx, Vy (8xy0)
		case 0x0000:
			// Set Vx = Vy
			c.v[x] = c.v[y]
			break
		// OR Vx, Vy (8xy1)
		case 0x0001:
			// Set Vx = Vx OR Vy
			c.v[x] |= c.v[y]
			break
		// AND Vx, Vy (8xy2)
		case 0x0002:
			// Set Vx = Vx AND Vy
			c.v[x] &= c.v[y]
			break
		// XOR Vx, Vy (8xy3)
		case 0x0003:
			// Set Vx XOR Vy
			c.v[x] ^= c.v[y]
			break
		// ADD Vx, Vy (8xy4)
		case 0x0004:
			// Set Vx = Vx + Vy, set VF = carry
			temp := uint16(c.v[x]) + uint16(c.v[y])
			if temp > 0xFF {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] = uint8(temp)
			break
		// SUB Vx, Vy (8xy5)
		case 0x0005:
			// Set Vx = Vx - Vy, set VF = NOT borrow
			if c.v[x] > c.v[y] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] = c.v[x] - c.v[y]
			break
		// SHR Vx, {, Vy} (8xy6)
		case 0x0006:
			// Set Vx = Vx SHR 1
			if c.v[x]&0b00000001 == 1 {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] >>= 1
			break
		// SUBN Vx, Vy (8xy7)
		case 0x0007:
			// Set Vx = Vy - Vx, set VF = NOT borrow
			if c.v[y] > c.v[x] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] = c.v[y] - c.v[x]
			break
		// SHL Vx, Vy (8xyE)
		case 0x000E:
			// Set Vx = Vx SHL 1
			if c.v[x]&0b10000000 == 1 {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] <<= 1
			break
		default:
			return unrecognizedOpcode(opcode)
		}
		break
	// SNE Vx, Vy (9xy0)
	case 0x9000:
		x := c.memory[pc] & 0x0F
		y := (c.memory[pc+1] & 0xF0) >> 2
		if c.v[x] != c.v[y] {
			c.pc += 2
		}
		break
	// LD I, addr (Annn)
	case 0xA000:
		// Set I = nnn
		c.i = opcode & 0x0FFF
		break
	// JP V0, addr (Bnnn)
	case 0xB000:
		// Jump to location nnn + V0
		c.pc = (opcode & 0x0FFF) + uint16(c.v[0])
		break
	// RND Vx, byte (Cxkk)
	case 0xC000:
		// Set Vx = random byte AND kk
		x := c.memory[pc] & 0x0F
		c.v[x] = byte(rand.Intn(256)) & c.memory[pc+1]
		break
	// DRW Vx, Vy, nibble (Dxyn)
	case 0xD000:
		// Display n-byte sprite starting at memory location I at (Vx, Vy),
		// set VF = collision
		x := c.memory[pc] & 0x0F
		y := (c.memory[pc+1] & 0xF0) >> 2
		n := c.memory[pc+1] & 0x0F
		c.displaySprite(c.v[x], c.v[y], byte(c.i), byte(n))
		break
	case 0xE000:
		switch opcode & 0x00FF {
		// SKP Vx (Ex9E)
		case 0x009E:
			// Skip next instruction if key with the value of Vx is pressed
			x := c.memory[pc] & 0x0F
			if c.isPressed(c.v[x]) {
				c.pc += 2
			}
			break
		// SKNP Vx (ExA1)
		case 0x00A1:
			// Skip next instruction if key with the value of Vx is not pressed
			x := c.memory[pc] & 0x0F
			if !c.isPressed(c.v[x]) {
				c.pc += 2
			}
			break
		default:
			return unrecognizedOpcode(opcode)
		}
	case 0xF000:
		x := c.memory[pc] & 0x0F
		switch opcode & 0x00FF {
		// LD Vx, DT (Fx07)
		case 0x0007:
			// Set Vx = delay timer value
			c.v[x] = c.dt
			break
		// LD Vx, K (Fx0A)
		case 0x000A:
			// Wait for a key press, store the value of the key in Vx
			c.v[x] = c.waitKey()
			break
		// LD DT, Vx (Fx15)
		case 0x0015:
			// Set delay timer = Vx
			c.dt = c.v[x]
			break
		// LD ST, Vx (Fx18)
		case 0x0018:
			// Set sound timer = Vx
			c.st = c.v[x]
			break
		// ADD I, Vx (Fx1E)
		case 0x001E:
			// Set I = I + Vx
			c.i += uint16(c.v[x])
			break
		// LD F, Vx (Fx29)
		case 0x0029:
			// Set I = location of sprite for digit Vx
			c.i = uint16(c.v[x]) * 5 // 5 bytes offset for every digit
			break
		// LD B, Vx (Fx33)
		case 0x0033:
			// Store BCD representation of Vx in memory locations I, I+1, and I+2
			val := c.v[x]
			c.memory[c.i+2] = val % 10
			val /= 10
			c.memory[c.i+1] = val % 10
			val /= 10
			c.memory[c.i] = val % 10
			break
		// LD [I], Vx (Fx55)
		case 0x0055:
			// Store registers V0 through Vx in memory starting at location I
			last := uint16(x)
			for i := uint16(0); i <= last; i++ {
				c.memory[c.i+i] = c.v[i]
			}
			break
		// LD Vx, [I] (Fx65)
		case 0x0065:
			// Read registers V0 through Vx from memory starting at location I
			last := uint16(x)
			for i := uint16(0); i <= last; i++ {
				c.v[i] = c.memory[c.i+i]
			}
			break
		default:
			return unrecognizedOpcode(opcode)
		}
		break
	default:
		return unrecognizedOpcode(opcode)
	}
	return nil
}

func unrecognizedOpcode(opcode uint16) error {
	return fmt.Errorf("unrecognized opcode %o", opcode)
}

func (c *Chip) draw() {
	var width float64 = 64

	st := tcell.StyleDefault
	black := st.Background(tcell.NewHexColor(0))
	gray := st.Background(tcell.NewHexColor(0x444444))

	var glyph rune = '0'

	for i := range c.display {
		x := i % 64
		y := int(math.Floor(float64(i) / width))

		if c.display[i] == 1 {
			c.screen.SetCell(x, y, black, glyph)
		} else {
			c.screen.SetCell(x, y, gray, glyph)
		}
	}
}

func (c *Chip) clearDisplay() {
	for i := range c.display {
		c.display[i] = 0
	}
	c.drawFlag = true
}

// (00,00)-------------(64,00)
// |                         |
// |                         |
// (00,32)-------------(64,32)
func (c *Chip) displaySprite(x byte, y byte, i byte, n byte) {
	c.v[0xF] = 0

	for q := byte(0); q < n; q++ {
		for p := byte(0); p < 8; p++ {
			// 0x80 >> p takes a bit for every loop
			// E.g.
			// memory & 0b10000000 >> 1
			pix := c.memory[i+q] & (0x80 >> p)
			if pix != 0 {
				tx := x + p
				ty := y + q
				if tx >= 64 {
					tx -= 64
				}
				if ty >= 32 {
					ty -= 32
				}

				// Set VF on collision...
				if c.display[ty*64+x] == 1 {
					c.v[0xF] = 1
				}

				c.display[ty*64+x] ^= 1
			}
		}
	}
	c.drawFlag = true
}

func (c *Chip) isPressed(val byte) bool {
	panic("not implemented yet!")
}

func (c *Chip) waitKey() byte {
	panic("not implemented yet!")
}
