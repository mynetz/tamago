// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linksettls

#include "textflag.h"

// setTLS implements runtime/goos.SetTLS for bare metal execution.
//
// The runtime only invokes the hook at CPL != 0; at CPL 0 it writes
// IA32_FS_BASE itself with WRMSR. This body exists so that bare metal
// programs link, and installs the base defensively should it ever be
// reached at CPL 0 (DI = base, already biased for -8(FS)).
//
// func setTLS(base uintptr)
TEXT setTLS(SB),NOSPLIT|NOFRAME,$0
	MOVQ	DI, AX
	MOVQ	$0xc0000100, CX	// IA32_MSR_FS_BASE
	MOVQ	$0, DX
	WRMSR
	RET
