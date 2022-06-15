package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

type opCode uint8

const (
	movA   opCode = 0b0011
	movB   opCode = 0b0111
	movA2B opCode = 0b0100
	movB2A opCode = 0b0001
	addA   opCode = 0b0000
	addB   opCode = 0b0101
	jmp    opCode = 0b1111
	jnc    opCode = 0b1110
	inA    opCode = 0b0010
	inB    opCode = 0b0110
	outB   opCode = 0b1001
	outIm  opCode = 0b1011
)

type register struct {
	A  uint8
	B  uint8
	C  bool  // キャリーフラグ
	PC uint8 // プログラムカウンタ
}

type port struct {
	In  uint8
	Out uint8
}

type emulator struct {
	Register *register
	Port     *port
	Rom      []uint8
}

func newEmulator(rom []uint8) (*emulator, error) {
	if len(rom) > 16 {
		return nil, errors.New("Rom size is over. Max size is 16.")
	}
	return &emulator{
		Register: &register{},
		Port:     &port{},
		Rom:      rom,
	}, nil
}

func (e *emulator) String() string {
	r := e.Register
	if r == nil {
		r = &register{}
	}

	p := e.Port
	if p == nil {
		p = &port{}
	}

	str := fmt.Sprintf("| A   |%04b |\n", e.Register.A)
	str += fmt.Sprintf("| B   |%04b |\n", e.Register.B)
	str += fmt.Sprintf("| C   |%v|\n", e.Register.C)
	str += fmt.Sprintf("| PC  |%04b |\n", e.Register.PC)
	str += fmt.Sprintf("| In  |%04b |\n", e.Port.In)
	str += fmt.Sprintf("| Out |%04b |\n", e.Port.Out)
	return str
}

//
// fetch-decode-execute サイクル
//
func (e *emulator) fetch() uint8 {
	if len(e.Rom) <= int(e.Register.PC) {
		return 0
	}
	return e.Rom[int(e.Register.PC)]
}

func (e *emulator) decode(row uint8) (opCode, uint8) {
	op := row >> 4
	im := row & 0x0f
	switch opCode(op) {
	case movA2B, movB2A, inA, inB, outB:
		return opCode(op), 0
	default:
		return opCode(op), im
	}
}

func (e *emulator) exec(ctx context.Context, t *time.Ticker) {
	for {
		select {
		case <-t.C:
			row := e.fetch()
			op, im := e.decode(row)

			switch op {
			case movA:
				e.movA(im)
			case movB:
				e.movB(im)
			case movA2B:
				e.movA2B()
			case movB2A:
				e.movB2A()
			case addA:
				e.addA(im)
			case addB:
				e.addB(im)
			case jmp:
				e.jmp(im)
			case jnc:
				e.jnc(im)
			case inA:
				e.inA()
			case inB:
				e.inB()
			case outB:
				e.outB()
			case outIm:
				e.outIm(im)
			default:
				err := errors.New("OpCode doesn't exist")
				log.Println(err)
				return
			}

			if debugFlag {
				log.Println("")
				fmt.Printf("| op  |%04b |\n", op)
				fmt.Printf("| im  |%04b |\n", im)
				fmt.Println(" -----")
				fmt.Printf("%v", e.String())
			}

			// 終了判定
			if int(e.Register.PC) >= len(e.Rom) {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

//
// 各命令の処理
//
func (e *emulator) movA(im uint8) {
	e.Register.A = im
	e.Register.C = false
	e.Register.PC++
}

func (e *emulator) movB(im uint8) {
	e.Register.B = im
	e.Register.C = false
	e.Register.PC++
}

func (e *emulator) movA2B() {
	e.Register.B = e.Register.A
	e.Register.C = false
	e.Register.PC++
}

func (e *emulator) movB2A() {
	e.Register.A = e.Register.B
	e.Register.C = false
	e.Register.PC++
}

func (e *emulator) addA(im uint8) {
	before := e.Register.A
	after := before + im
	if after > 0x0f {
		e.Register.C = true
	}
	e.Register.A = after & 0x0f
	e.Register.PC++
}

func (e *emulator) addB(im uint8) {
	before := e.Register.B
	after := before + im
	if after > 0x0f {
		e.Register.C = true
	}
	e.Register.B = after & 0x0f
	e.Register.PC++
}

func (e *emulator) jmp(im uint8) {
	e.Register.PC = im
	e.Register.C = false
}

func (e *emulator) jnc(im uint8) {
	if e.Register.C == false {
		e.Register.PC = im
	} else {
		e.Register.PC++
	}
	e.Register.C = false
}

func (e *emulator) inA() {
	e.Register.A = e.Port.In
	e.Register.C = false
	e.Register.PC++
}

func (e *emulator) inB() {
	e.Register.B = e.Port.In
	e.Register.C = false
	e.Register.PC++
}

func (e *emulator) outB() {
	e.Port.Out = e.Register.B
	log.Printf("Port B Out: %04b\n", e.Port.Out)
	e.Register.C = false
	e.Register.PC++
}

func (e *emulator) outIm(im uint8) {
	e.Port.Out = im
	log.Printf("Port Out: %04b\n", e.Port.Out)
	e.Register.C = false
	e.Register.PC++
}

//
// エミュレータ実行
//
var debugFlag bool

func main() {
	flag.BoolVar(&debugFlag, "d", false, "デバッグログを出す")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

	// Romを色々変えて試す
	rom := calcRom
	em, err := newEmulator(rom)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("--- start emulator ---")
	if debugFlag {
		fmt.Printf("%v", em.String())
	}
	em.exec(ctx, t)

	log.Println("--- end ---")
}

//
// Rom
//

// 1+1
var calcRom = []uint8{
	0b00110001, // MOV A 1
	0b00000001, // ADD A 1
	0b01000000, // MOV B A
	0b10010000, // OUT B
}

// Lチカ
var ledRom = []uint8{
	0b10110011, // OUT 0011
	0b10110110, // OUT 0110
	0b10111100, // OUT 1100
	0b10111000, // OUT 1000
	0b10111000, // OUT 1000
	0b10111100, // OUT 1100
	0b10110110, // OUT 0110
	0b10110011, // OUT 0011
	0b10110001, // OUT 0001
	0b11110000, // JMP 0 無限ループ
}

// ラーメンタイマー
var ramenRom = []uint8{
	0b10110111, // OUT 0111 LED3つ点灯
	0b00000001, // ADD A 1
	0b11100001, // JNC 0001 キャリーが発生するまでループ(16回ループ)
	0b00000001, // ADD A 1
	0b11100011, // JNC 0011 ここも16回ループ
	0b10110110, // OUT 0110 LED2つ点灯
	0b00000001, // ADD A 1
	0b11100110, // JNC 0110 ここも16回ループ
	0b00000001, // ADD A 1
	0b11101000, // JNC 1000 ここも16回ループ
	0b10110000, // OUT 0000 LED全て消灯
	0b10110100, // OUT 0100 LED1つ点灯
	0b00000001, // ADD A 1
	0b11101010, // JNC 1010 ここも16回ループ
	0b10111000, // OUT 1000 終了のLED点灯
	0b11111111, // JMP 1111 無限ループ
}
