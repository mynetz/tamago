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

// cpuinit is the program entry (runtime/goos.CPUInit jumps here).
//
// The L4Re loader (l4re_itas) starts the main thread with a SysV style
// initial stack: argc, argv[], NULL, envp[], NULL, then auxv pairs
// (a_type, a_val) terminated by AT_NULL. We only need AT_L4_ENV and
// AT_L4_KIP.
//
// Once the pointers are saved we switch to the bootstrap stack, let Go
// code (earlyInit, nosplit) allocate and attach the heap dataspace, then
// publish the region to the runtime, place the runtime stack at its top
// and enter the Go runtime.
TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	MOVQ	SP, AX

	// skip argc and argv[0..argc] (argc+1 entries incl. NULL)
	MOVQ	(AX), BX
	INCQ	BX
	SHLQ	$3, BX
	ADDQ	$8, AX
	ADDQ	BX, AX

	// skip envp until NULL
envp:
	MOVQ	(AX), CX
	ADDQ	$8, AX
	TESTQ	CX, CX
	JNE	envp

	// AX = &auxv[0]
auxv:
	MOVQ	(AX), CX
	MOVQ	8(AX), DX
	ADDQ	$16, AX
	TESTQ	CX, CX
	JE	done
	CMPQ	CX, $AT_L4_ENV
	JNE	kip
	MOVQ	DX, ·l4reEnv(SB)
	JMP	auxv
kip:
	CMPQ	CX, $AT_L4_KIP
	JNE	auxv
	MOVQ	DX, ·l4Kip(SB)
	JMP	auxv

done:
	// bootstrap stack (16-byte aligned top of bootStack)
	LEAQ	·bootStack+(16*1024-16)(SB), SP
	ANDQ	$~15, SP

	// Calling Go from assembly goes through a compiler generated ABI0
	// wrapper that loads g from TLS (-8(FS)). Install a provisional TLS
	// base pointing at bootTLS so that g reads as &bootG (a zeroed
	// stand-in that nosplit code never dereferences beyond the stack
	// guard). The runtime installs the real base for m0 later on.
	LEAQ	·bootG(SB), AX
	MOVQ	AX, ·bootTLS(SB)
	LEAQ	·bootTLS+8(SB), DI
	CALL	setFSBase(SB)

	CALL	·earlyInit(SB)

	// Drop the provisional TLS base again. The runtime installs its own
	// through the SIGSEGV handler registered by earlyInit, which relies
	// on the first -8(FS) access faulting (see sig_amd64.s).
	XORQ	DI, DI
	CALL	setFSBase(SB)

	// runtime/goos.RamStart/RamSize describe the heap region
	MOVQ	·heapStart(SB), AX
	MOVQ	·heapSize(SB), BX
	MOVQ	AX, runtime∕goos·RamStart(SB)
	MOVQ	BX, runtime∕goos·RamSize(SB)

	// runtime stack at the top of the region
	ADDQ	BX, AX
	SUBQ	runtime∕goos·RamStackOffset(SB), AX
	ANDQ	$~15, AX
	MOVQ	AX, SP

	JMP	_rt0_tamago_start(SB)
