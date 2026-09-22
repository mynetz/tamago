// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// ELF auxiliary vector tags used by the L4Re loader
// (l4util/include/elf.h).
#define AT_NULL    0
#define AT_L4_ENV  0xf1
#define AT_L4_KIP  0xf2

// cpuinit is the program entry (runtime/goos.CPUInit jumps here), see
// cpuinit_amd64.s for the description of the initial stack layout.
//
// The thread starts at EL0; there is no privilege level handling.
TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	MOVD	RSP, R0

	// skip argc and argv[0..argc]
	MOVD	(R0), R1
	ADD	$1, R1, R1
	LSL	$3, R1, R1
	ADD	$8, R0, R0
	ADD	R1, R0, R0

	// skip envp until NULL
envp:
	MOVD	(R0), R2
	ADD	$8, R0, R0
	CBNZ	R2, envp

	// R0 = &auxv[0]
auxv:
	MOVD	(R0), R2
	MOVD	8(R0), R3
	ADD	$16, R0, R0
	CBZ	R2, done
	CMP	$AT_L4_ENV, R2
	BNE	kip
	MOVD	R3, ·l4reEnv(SB)
	B	auxv
kip:
	CMP	$AT_L4_KIP, R2
	BNE	auxv
	MOVD	R3, ·l4Kip(SB)
	B	auxv

done:
	// bootstrap stack (16-byte aligned top of bootStack)
	MOVD	$·bootStack(SB), R1
	ADD	$(16*1024-16), R1, R1
	AND	$~15, R1, R1
	MOVD	R1, RSP

	CALL	·earlyInit(SB)

	// runtime/goos.RamStart/RamSize describe the heap region
	MOVD	·heapStart(SB), R0
	MOVD	·heapSize(SB), R1
	MOVD	R0, runtime∕goos·RamStart(SB)
	MOVD	R1, runtime∕goos·RamSize(SB)

	// runtime stack at the top of the region
	ADD	R1, R0, R0
	MOVD	runtime∕goos·RamStackOffset(SB), R2
	SUB	R2, R0, R0
	AND	$~15, R0, R0
	MOVD	R0, RSP

	B	_rt0_tamago_start(SB)
