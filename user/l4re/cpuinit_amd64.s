// L4Re userspace support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// AT_NULL / AT_L4_ENV / AT_L4_KIP from
// sources/l4re/pkg/l4re-core/libc/uclibc-ng/contrib/uclibc/include/elf.h
#define AT_NULL    0
#define AT_L4_ENV  0xf1
#define AT_L4_KIP  0xf2

// cpuinit is the ELF entry point. The L4Re loader gives us a SysV-style
// stack: argc at 0(SP), argv pointers, NULL, envp pointers, NULL, then the
// auxv array of (uint64 a_type, uint64 a_un.a_val) pairs terminated with
// AT_NULL. We scan auxv for AT_L4_ENV (env page pointer) and AT_L4_KIP
// (kernel info page pointer), stash them in package globals, set up the Go
// runtime stack inside the RAM region declared in runtime.go, and jump to
// runtime·rt0_amd64_tamago.
//
// Selected via the `linkcpuinit` build tag and the linker flag
// `-E cpuinit`.
TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// AX = walking pointer over the stack-supplied vectors.
	// Start at SP and skip past argc + argv + envp.
	MOVQ	SP, AX

	// argc
	MOVQ	(AX), BX	// BX = argc
	ADDQ	$8, AX		// past argc

	// skip argv: BX+1 entries (including the NULL terminator)
	INCQ	BX
	IMULQ	$8, BX
	ADDQ	BX, AX

	// skip envp: walk until NULL
skip_envp:
	MOVQ	(AX), CX
	ADDQ	$8, AX
	TESTQ	CX, CX
	JNE	skip_envp

	// AX now points at auxv[0]. Walk pairs (type, value) until AT_NULL.
auxv_loop:
	MOVQ	(AX), CX	// type
	MOVQ	8(AX), DX	// value
	TESTQ	CX, CX
	JE	auxv_done

	CMPQ	CX, $AT_L4_ENV
	JNE	check_kip
	MOVQ	DX, ·l4reEnv(SB)
	JMP	auxv_next

check_kip:
	CMPQ	CX, $AT_L4_KIP
	JNE	auxv_next
	MOVQ	DX, ·l4Kip(SB)

auxv_next:
	ADDQ	$16, AX
	JMP	auxv_loop

auxv_done:
	// Resolve runtime/goos.RamStart from the materialised heapPad BSS
	// array (declared in runtime.go). At link time heapPad has a fixed
	// address; we publish that as RamStart so the Go runtime's allocator
	// places arenas there instead of in unmapped memory.
	LEAQ	·heapPad(SB), AX
	MOVQ	AX, runtime∕goos·RamStart(SB)

	// Set up the Go runtime stack at the top of RamStart..+RamSize.
	MOVQ	runtime∕goos·RamStart(SB), SP
	MOVQ	runtime∕goos·RamSize(SB), AX
	MOVQ	runtime∕goos·RamStackOffset(SB), BX
	ADDQ	AX, SP
	SUBQ	BX, SP

	JMP	_rt0_tamago_start(SB)
