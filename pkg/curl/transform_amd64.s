// Code generated by command: go run transform_amd64_asm.go -out ../transform_amd64.s -stubs ../transform_amd64.go -pkg curl. DO NOT EDIT.

// +build amd64,gc,!purego

#include "textflag.h"

// func transform(lto *[729]uint, hto *[729]uint, lfrom *[729]uint, hfrom *[729]uint)
TEXT ·transform(SB), NOSPLIT, $0-32
	MOVQ lto+0(FP), AX
	MOVQ hto+8(FP), CX
	MOVQ lfrom+16(FP), DX
	MOVQ hfrom+24(FP), BX
	MOVQ $0x00000051, BP

RoundLoop:
	// a = from[0]
	MOVQ (DX), SI
	MOVQ (BX), DI

	// b = from[364]
	MOVQ 2912(DX), R8
	MOVQ 2912(BX), R9

	// a = sBox(a, b)
	XORQ R8, DI
	ANDQ SI, DI
	XORQ R9, SI
	ORQ  DI, SI
	NOTQ DI

	// to[0] = a
	MOVQ DI, (AX)
	MOVQ SI, (CX)
	MOVQ $0x0000016c, R10
	MOVQ $0x00000001, R11

StateLoop:
	// a = from[364+t]
	MOVQ 2912(DX)(R10*8), DI
	MOVQ 2912(BX)(R10*8), SI

	// b = sBox(b, a)
	XORQ DI, R9
	ANDQ R8, R9
	XORQ SI, R8
	ORQ  R9, R8
	NOTQ R9

	// to[0+i] = b
	MOVQ R9, (AX)(R11*8)
	MOVQ R8, (CX)(R11*8)

	// b = from[-1+t]
	MOVQ -8(DX)(R10*8), R9
	MOVQ -8(BX)(R10*8), R8

	// a = sBox(a, b)
	XORQ R9, SI
	ANDQ DI, SI
	XORQ R8, DI
	ORQ  SI, DI
	NOTQ SI

	// to[1+i] = a
	MOVQ SI, 8(AX)(R11*8)
	MOVQ DI, 8(CX)(R11*8)

	// a = from[363+t]
	MOVQ 2904(DX)(R10*8), SI
	MOVQ 2904(BX)(R10*8), DI

	// b = sBox(b, a)
	XORQ SI, R8
	ANDQ R9, R8
	XORQ DI, R9
	ORQ  R8, R9
	NOTQ R8

	// to[2+i] = b
	MOVQ R8, 16(AX)(R11*8)
	MOVQ R9, 16(CX)(R11*8)

	// b = from[-2+t]
	MOVQ -16(DX)(R10*8), R8
	MOVQ -16(BX)(R10*8), R9

	// a = sBox(a, b)
	XORQ R8, DI
	ANDQ SI, DI
	XORQ R9, SI
	ORQ  DI, SI
	NOTQ DI

	// to[3+i] = a
	MOVQ DI, 24(AX)(R11*8)
	MOVQ SI, 24(CX)(R11*8)
	SUBQ $0x00000002, R10
	ADDQ $0x00000004, R11
	CMPQ R11, $0x000002d9
	JL   StateLoop

	// swap buffers
	XCHGQ DX, AX
	XCHGQ BX, CX
	DECQ  BP
	JNZ   RoundLoop
	RET