// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package sys exposes the Fiasco/L4 kernel ABI needed by a `GOOS=tamago` Go
// program running as a native L4Re task. It is the Go counterpart of the C
// `l4sys` library shipped with L4Re and is deliberately thin: it knows about
// message tags, capability selectors, the per-thread UTCB and the IPC
// system call, nothing else.
//
// The package provides:
//
//   - IPC system call trampolines (Send and Call) written in Go assembly
//     following the per-architecture register convention of
//     l4sys/include/ARCH-*/arch/ipc.h.
//   - Accessors for the message registers (MR), buffer registers (BR) and
//     thread control registers (TCR) of the calling thread's UTCB.
//   - Message tag construction and decoding.
//   - The KIP clock accessor used for time keeping.
//   - Constants for protocols, flags, capability selectors and error codes.
//
// All functions are `//go:nosplit`, allocation-free and touch no goroutine
// state, so that they may be called from the L4Re overlay's early
// initialization path (before the Go runtime is started) as well as from
// the runtime's pre-World hooks.
//
// Higher-level L4Re service bindings (memory allocator, region map, console,
// parent) live in the sibling package
// github.com/usbarmory/tamago/user/l4re/svc.
package sys
