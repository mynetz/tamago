// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package svc provides Go bindings for the L4Re user-level services that a
// freshly started task reaches through its initial environment: the memory
// allocator (Mem_alloc), the region map (Rm), the console (Vcon/Log) and
// the parent. It is the Go counterpart of the C `l4re_c` library and of the
// L4Re::* C++ client stubs; message layouts follow the L4::Ipc RPC
// marshalling rules of l4sys/include/cxx/ipc_iface.
//
// All entry points perform synchronous IPC on the calling thread's UTCB
// and never allocate. Functions marked `//go:nosplit` are used by the
// overlay before the Go runtime is started and must stay free of stack
// growth, write barriers and goroutine state.
package svc

import (
	"unsafe"

	"github.com/usbarmory/tamago/user/l4re/sys"
)

// Cap is a capability selector (slot index shifted by sys.CapShift).
type Cap uint64

// IsValid reports whether the selector denotes a capability slot.
//
//go:nosplit
func (c Cap) IsValid() bool { return uint64(c)&sys.InvalidCapBit == 0 }

// Env mirrors the C `l4re_env_t` structure (l4re/include/env.h) that the
// L4Re loader hands to a new task through the AT_L4_ENV auxiliary vector
// entry. Field order and sizes must match the C definition exactly.
type Env struct {
	Parent            Cap     // parent object (L4Re::Parent)
	Rm                Cap     // region map (L4Re::Rm)
	MemAlloc          Cap     // memory allocator (L4Re::Mem_alloc)
	Log               Cap     // console (L4::Vcon)
	MainThread        Cap     // first user thread
	Factory           Cap     // kernel object factory
	Scheduler         Cap     // scheduler
	Itas              Cap     // ITAS services
	DbgEvents         Cap     // debug events service
	FirstFreeCap      uint64  // first free slot index (not shifted)
	FirstFreeReplyCap uint64  // first free reply capability index
	UtcbArea          uint64  // l4_fpage_t of the UTCB area
	FirstFreeUtcb     uint64  // first UTCB available to the application
	Caps              uintptr // initial objects array (l4re_env_cap_entry_t)
}

var (
	env     *Env
	kip     uintptr
	nextCap uint64
)

// Init publishes the initial environment and the Kernel Info Page
// addresses decoded from the auxiliary vector. It must be called once
// before any other function of this package, typically from the
// overlay's CPU initialization path.
//
//go:nosplit
func Init(envAddr, kipAddr uintptr) {
	if envAddr == 0 {
		return
	}
	env = (*Env)(unsafe.Pointer(envAddr))
	kip = kipAddr
	nextCap = env.FirstFreeCap << sys.CapShift
}

// GetEnv returns the initial environment, or nil before Init.
//
//go:nosplit
func GetEnv() *Env { return env }

// KIP returns the Kernel Info Page address, or 0 before Init.
//
//go:nosplit
func KIP() uintptr { return kip }

// AllocCap reserves the next free capability slot. Slots are handed out
// from a bump pointer seeded by l4re_env_t.first_free_cap; the kernel
// populates a slot when an IPC maps a capability into it.
//
//go:nosplit
func AllocCap() Cap {
	c := Cap(nextCap)
	nextCap += sys.CapStride
	return c
}
