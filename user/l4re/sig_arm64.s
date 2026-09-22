// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// CPU feature detection without toolchain support.
//
// internal/cpu.osInit for GOOS=tamago reads ID_AA64ISAR0_EL1,
// ID_AA64ISAR1_EL1 and ID_AA64PFR0_EL1 with MRS. The registers are
// accessible at EL1 (bare metal) and Linux emulates the access for EL0,
// but Fiasco delivers an exception (ESR EC = 0, "unknown") which l4re_itas
// turns into SIGILL. sigill below is installed as the handler before the
// runtime starts: it decodes the faulting instruction, and for the three
// registers stores 0 (no optional features, generic code paths) into the
// destination register of the interrupted context and skips the
// instruction. The handler is removed again in Hwinit1.
//
// MRS encodings (op0=3, op1=0, CRn=0): ID_AA64PFR0_EL1  CRm=4 op2=0
//                                     ID_AA64ISAR0_EL1 CRm=6 op2=0
//                                     ID_AA64ISAR1_EL1 CRm=6 op2=1
// as instruction words with Rt = 0:  0xd5380400, 0xd5380600, 0xd5380620.

#define L4_SYSF_CALL	0x03

// musl layout (verified against L4Re's headers)
#define UC_REGS		184	// ucontext_t.uc_mcontext.regs[0]
#define UC_PC		440	// ucontext_t.uc_mcontext.pc

// sigReset resets the disposition of the signal in R0 to SIG_DFL with
// L4Re::Itas::sigaction (see svc.Sigaction for the message layout).
// Assembly only: the handler may run in contexts where no Go code may be
// entered. Clobbers R0-R8.
TEXT sigReset(SB),NOSPLIT|NOFRAME,$0
	MRS	TPIDRRO_EL0, R4		// UTCB
	MOVW	$2, R5
	MOVW	R5, (R4)		// opcode sigaction
	MOVW	R0, 4(R4)		// signum
	ADD	$8, R4, R5		// struct sigaction: 19 zero words
	MOVD	$19, R6
zero:
	MOVD	ZR, (R5)
	ADD	$8, R5, R5
	SUB	$1, R6, R6
	CBNZ	R6, zero
	MOVD	·l4reEnv(SB), R2
	MOVD	56(R2), R2		// l4re_env_t.itas
	ORR	$L4_SYSF_CALL, R2, R2
	MOVD	$0x400a0014, R0		// (L4RE_PROTO_ITAS << 16) | 20 words
	MOVD	ZR, R1
	MOVD	ZR, R3
	SVC
	RET

// sigill is the SIGILL handler: void (*)(int signo, siginfo_t *si,
// ucontext_t *uc), C calling convention (R0, R1, R2), entered from itas'
// sigenter trampoline. Only MRS of the three ID registers is emulated;
// anything else resets the disposition to SIG_DFL so that the re-raised
// exception produces itas' diagnostic dump.
TEXT sigill(SB),NOSPLIT|NOFRAME,$0
	MOVD	UC_PC(R2), R3
	MOVWU	(R3), R4		// faulting instruction
	AND	$0x1f, R4, R5		// Rt
	BIC	$0x1f, R4, R4		// opcode without Rt
	MOVW	$0xd5380400, R6
	CMPW	R6, R4
	BEQ	emulate
	MOVW	$0xd5380600, R6
	CMPW	R6, R4
	BEQ	emulate
	MOVW	$0xd5380620, R6
	CMPW	R6, R4
	BEQ	emulate

	// R0 = signo
	B	sigReset(SB)

emulate:
	CMP	$31, R5			// Rt == 31 is XZR: nothing to store
	BEQ	skip
	ADD	R5<<3, R2, R6
	MOVD	ZR, UC_REGS(R6)		// regs[Rt] = 0
skip:
	ADD	$4, R3, R3
	MOVD	R3, UC_PC(R2)		// resume after the MRS
	RET

// sigillHandler returns the address of sigill for Sigaction.
//
// func sigillHandler() uintptr
TEXT ·sigillHandler(SB),NOSPLIT,$0-8
	MOVD	$sigill(SB), R0
	MOVD	R0, ret+0(FP)
	RET
