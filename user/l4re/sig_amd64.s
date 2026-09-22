// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// Thread local storage bring-up without toolchain support.
//
// The Go runtime keeps g at -8(FS). At CPL 3, runtime.settls asks the
// kernel to install FS_BASE with Linux' arch_prctl(ARCH_SET_FS, base):
//
//	RAX = 158, RDI = 0x1002, RSI = base, SYSCALL
//
// Under Fiasco the SYSCALL instruction is the L4 IPC entry and this request
// is executed as an IPC with the register contents as operands. RDX still
// holds CPUID.1:EDX from the feature probe in rt0_amd64_tamago; its bit 3
// (PSE, architectural on x86-64) is the L4_obj_ref::Ipc_reply flag, so the
// kernel looks up the (empty) implicit reply capability, fails with
// L4_error::Not_existent and returns immediately with the error bit set in
// RAX and all other registers, RSI in particular, untouched
// (fiasco/src/kern/thread_object.cpp Obj_cap::deref, sys_ipc_wrapper).
// The runtime's error check accepts that and the very next instruction, a
// store through -8(FS) with FS_BASE still 0, faults at 0xfffffffffffffff8.
//
// l4re_itas turns that fault into SIGSEGV. sigsegv below is installed as
// the handler before the runtime starts: it recognises the fault address,
// takes the base from the interrupted RSI, installs it with the L4 thread
// system call and returns; itas resumes the faulting instruction, which
// now succeeds. The handler is removed again in Hwinit1.

#define THREAD_SET_SEG_BASE_OP	0x12			// L4_THREAD_AMD64_SET_SEGMENT_BASE_OP, FS
#define MSGTAG_THREAD_WORDS2	0xFFFFFFFFFFF40002	// (L4_PROTO_THREAD << 16) | 2
#define L4_INVALID_CAP		0xFFFFFFFFFFFFF800	// resolves to the calling thread
#define L4_SYSF_CALL		0x03

#define TLS_FAULT_ADDR		0xfffffffffffffff8	// -8(FS) with FS_BASE == 0

// musl layout (verified against L4Re's headers)
#define SI_ADDR			16	// siginfo_t.si_addr
#define UC_RSI			112	// ucontext_t.uc_mcontext.gregs[REG_RSI]

// setFSBase installs FS_BASE = DI for the calling thread
// (L4::Thread::amd64_set_segment_base_op on L4_INVALID_CAP = self).
// Clobbers AX, BX, CX, DX, SI, R8, R11.
TEXT setFSBase(SB),NOSPLIT|NOFRAME,$0
	BYTE	$0x65; MOVQ (0), BX	// BX = *(gs:0) = UTCB
	MOVQ	$THREAD_SET_SEG_BASE_OP, (BX)
	MOVQ	DI, 8(BX)
	MOVQ	$(L4_INVALID_CAP | L4_SYSF_CALL), DX
	MOVQ	$MSGTAG_THREAD_WORDS2, AX
	XORQ	SI, SI
	XORQ	R8, R8
	SYSCALL
	RET

// sigReset resets the disposition of the signal in DI to SIG_DFL with
// L4Re::Itas::sigaction (see svc.Sigaction for the message layout). It is
// written in assembly because the handler may run before the runtime's TLS
// exists, when no Go code (not even nosplit) can be entered. Clobbers AX,
// BX, CX, DX, SI, R8, R11.
TEXT sigReset(SB),NOSPLIT|NOFRAME,$0
	BYTE	$0x65; MOVQ (0), BX	// BX = UTCB
	MOVL	$2, (BX)		// opcode sigaction
	MOVL	DI, 4(BX)		// signum
	LEAQ	8(BX), SI		// struct sigaction: 19 zero words (SIG_DFL)
	MOVQ	$19, CX
	XORQ	AX, AX
zero:
	MOVQ	AX, (SI)
	ADDQ	$8, SI
	DECQ	CX
	JNE	zero
	MOVQ	·l4reEnv(SB), DX
	MOVQ	56(DX), DX		// l4re_env_t.itas
	ORQ	$L4_SYSF_CALL, DX
	MOVQ	$0x400a0014, AX		// (L4RE_PROTO_ITAS << 16) | 20 words
	XORQ	SI, SI
	XORQ	R8, R8
	SYSCALL
	RET

// sigsegv is the SIGSEGV handler: void (*)(int signo, siginfo_t *si,
// ucontext_t *uc), C calling convention (RDI, RSI, RDX), entered from
// itas' sigenter trampoline. Only the TLS bootstrap fault is handled; any
// other SIGSEGV resets the disposition to SIG_DFL so that the re-raised
// fault produces itas' diagnostic dump.
TEXT sigsegv(SB),NOSPLIT|NOFRAME,$0
	MOVQ	SI_ADDR(SI), AX
	MOVQ	$TLS_FAULT_ADDR, CX
	CMPQ	AX, CX
	JNE	other

	MOVQ	UC_RSI(DX), DI		// base from the interrupted arch_prctl
	CALL	setFSBase(SB)
	RET

other:
	CALL	sigReset(SB)		// DI = signo
	RET

// sigsegvHandler returns the address of sigsegv for Sigaction.
//
// func sigsegvHandler() uintptr
TEXT ·sigsegvHandler(SB),NOSPLIT,$0-8
	LEAQ	sigsegv(SB), AX
	MOVQ	AX, ret+0(FP)
	RET
