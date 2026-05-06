// L4Re userspace support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package l4re_user provides support for using `GOOS=tamago` to build Go
// programs that run as native L4Re tasks on the Fiasco microkernel.
//
// This package is only meant to be used with `GOOS=tamago` as supported by
// the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
//
// The package mirrors the structure of `tamago/user/linux` and the
// para-virtualisation approach used by `go-boot` for UEFI runtime services.
// Instead of issuing Linux SYSCALLs (user/linux) or calling UEFI service
// function pointers (go-boot), it issues L4 IPC syscalls directly via Go
// assembly to capabilities discovered from the L4Re initial environment.
//
// This iteration only wires up `printk`. `ramStart`/`ramSize` are set to
// fixed values matching the L4Re q35 layout used by the integration repo's
// hello-go-cfg scenario.
package l4re_user

import (
	"runtime/goos"
	_ "unsafe"
)

// Initial environment pointers, populated by cpuinit (see ipc_amd64.s).
//
// The L4Re loader passes these to a freshly-started task via the ELF
// auxiliary vector entries:
//
//	AT_L4_ENV  = 0xf1  -> pointer to l4re_env_t
//	AT_L4_KIP  = 0xf2  -> pointer to the Kernel Info Page
//
// See sources/l4re/pkg/l4re-core/libc/uclibc-ng/contrib/uclibc/include/elf.h
// and sources/l4re/pkg/l4re-core/libc/uclibc-ng/contrib/uclibc/libc/misc/elf/dl-support.c
// in the integration repo for the upstream consumer of this convention.
var (
	l4reEnv uint64 // l4re_env_t *
	l4Kip   uint64 // l4_kip_t *
)

// RAM region. We must keep this entirely within the pages L4Re actually
// maps for our task. Since iteration 2b doesn't yet allocate dataspaces
// dynamically, we *materialise* the RAM region as a fixed BSS array
// (heapPad below). The Go runtime then allocates inside that array and
// L4Re backs the whole mapping with real pages because they are part of
// the ELF's PT_LOAD RW segment.
//
// ramStart is set in cpuinit to the address of heapPad[0]; ramSize is its
// length. ramStackOffset places the stack near the top of that region.

//go:linkname ramStart runtime/goos.RamStart
var ramStart uint64 = 0

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = uint64(len(heapPad))

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint64 = 0x1000

// heapPad reserves a fixed RW region inside the ELF for Go's runtime heap
// and stack. 8 MiB is plenty for hello world. Increase deliberately if a
// future app exhausts it.
var heapPad [8 * 1024 * 1024]byte

// Hwinit0 runs before the Go runtime is available. We seed Bloc so the heap
// allocator places its arenas inside the RAM region we declared.
//
//go:linkname hwinit0 runtime/goos.Hwinit0
func hwinit0() {
	goos.Bloc = uintptr(ramStart)
}

// Hwinit1 runs after the Go runtime is up. Nothing to do for the minimal
// build; reserved for future RNG / clock / IRQ initialisation.
//
//go:linkname hwinit1 runtime/goos.Hwinit1
func hwinit1() {
}

// nanotime returns the system time in nanoseconds.
//
// Until KIP-clock support is wired up we return a monotonically increasing
// counter so that Go runtime housekeeping (ticks, GC pacing) sees forward
// progress without panic. It is *not* a wall clock.
var nanoCounter int64

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	nanoCounter += 1000
	return nanoCounter
}

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {}

//go:linkname getRandomData runtime/goos.GetRandomData
func getRandomData(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// printk writes a single byte to the L4Re log capability via Vcon. Tamago's
// runtime calls printk byte-by-byte; we accumulate into a small line buffer
// and flush on newline or when the buffer is full.
const printBufLen = 256

var (
	printBuf [printBufLen]byte
	printLen int
)

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	if l4reEnv == 0 {
		// L4Re env was not captured at startup; nothing we can do.
		return
	}
	printBuf[printLen] = c
	printLen++
	if c == '\n' || printLen == printBufLen {
		flushLog()
	}
}

func flushLog() {
	if printLen == 0 {
		return
	}
	// Offset of l4re_env_t.log: parent (0), rm (8), mem_alloc (16), log (24).
	// See sources/l4re/pkg/l4re-core/l4re/include/env.h.
	logCap := *(*uint64)(unsafePtr(l4reEnv + 0x18))
	vconWrite(logCap, printBuf[:printLen])
	printLen = 0
}

// flushLogIfPending is exported as a side-effect-free helper for tests.
func flushLogIfPending() { flushLog() }
