// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package l4re_user provides support for using `GOOS=tamago` in L4Re user
// space, i.e. for running Go programs as native tasks on the Fiasco
// microkernel, started by the L4Re loader (ned/moe) like any other L4Re
// application.
//
// The package implements the `runtime/goos` overlay contract without
// linking L4Re's C/C++ libraries and without cgo: the L4 system call is
// issued from Go assembly and the L4Re services (memory allocator, region
// map, console, parent) are spoken to with hand-marshalled IPC, following
// the pattern go-boot uses for UEFI. The kernel ABI lives in package
// user/l4re/sys, the service bindings in user/l4re/svc.
//
// Startup sequence (cpuinit_$GOARCH.s):
//
//  1. Decode the auxiliary vector left on the initial stack by the L4Re
//     loader to find the initial environment (AT_L4_ENV) and the Kernel
//     Info Page (AT_L4_KIP).
//  2. On a small bootstrap stack, allocate the Go heap as an L4Re
//     dataspace and attach it to the address space (earlyInit).
//  3. Publish the region as runtime/goos.RamStart/RamSize, place the
//     runtime stack at its top and enter the Go runtime.
//
// This package is only meant to be used with `GOOS=tamago` as supported by
// the TamaGo framework for bare metal Go, see https://github.com/usbarmory/tamago.
package l4re_user

import (
	"runtime/goos"
	_ "unsafe"

	"github.com/usbarmory/tamago/user/l4re/svc"
	"github.com/usbarmory/tamago/user/l4re/sys"
)

// HeapSize is the size of the dataspace allocated for the Go heap and
// runtime stack. It is a compile time constant because the region must
// exist before the runtime starts.
const HeapSize = 64 * 1024 * 1024

// Initial environment pointers, decoded from the auxiliary vector by
// cpuinit before any Go code runs.
var (
	l4reEnv uintptr // l4re_env_t *, AT_L4_ENV (0xf1)
	l4Kip   uintptr // Kernel Info Page, AT_L4_KIP (0xf2)
)

// Runtime memory region, established by earlyInit. cpuinit publishes it as
// runtime/goos.RamStart and RamSize and places the runtime stack
// RamStackOffset bytes below its end.
var (
	heapStart uintptr
	heapSize  uintptr
)

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint64 = 0x1000

// bootStack is the stack earlyInit runs on. Pre-runtime Go code is
// nosplit and shallow, the size is generous.
var bootStack [16 * 1024]byte

// earlyInit runs on the bootstrap stack before the Go runtime exists: no
// goroutine, no TLS, no stack growth. It and everything it calls must be
// nosplit and free of allocation.
//
// It allocates the heap dataspace from the memory allocator and attaches
// it to the address space, publishing the region in heapStart/heapSize.
// On failure it reports on the console and terminates the task.
//
//go:nosplit
func earlyInit() {
	svc.Init(l4reEnv, l4Kip)

	if svc.GetEnv() == nil {
		// No environment: we were not started by an L4Re loader.
		for {
			sys.Sleep(sys.TimeoutNever)
		}
	}

	ds, err := svc.AllocDataspace(HeapSize, 0, 0)
	if err != 0 {
		earlyFatal("l4re: heap dataspace allocation failed: ", err)
	}

	start, err := svc.Attach(0, HeapSize, svc.RmRW|svc.RmSearchAddr, ds)
	if err != 0 {
		earlyFatal("l4re: heap attach failed: ", err)
	}

	heapStart = start
	heapSize = HeapSize
}

//go:nosplit
func earlyFatal(msg string, err sys.Errno) {
	svc.LogString(msg)
	svc.LogString(err.Error())
	svc.LogString("\n")
	svc.Exit(1)
}

// hwinit0 runs before the World is started. The heap region was attached
// by earlyInit; seed the allocator base.
//
//go:linkname hwinit0 runtime/goos.Hwinit0
func hwinit0() {
	goos.Bloc = heapStart
}

// hwinit1 runs after the World is started.
//
//go:linkname hwinit1 runtime/goos.Hwinit1
func hwinit1() {
	goos.Exit = exit
}

// exit terminates the task through the parent (ned/moe record the exit
// code), see svc.Exit.
func exit(code int32) {
	flushLog()
	svc.Exit(code)
}

// nanotime returns the kernel clock in nanoseconds, read through the
// kernel-provided KIP stub.
//
//go:linkname nanotime runtime/goos.Nanotime
//go:nosplit
func nanotime() int64 {
	if l4Kip == 0 {
		return 0
	}
	return sys.KipClockNs(l4Kip)
}

// L4Re offers no entropy source to plain tasks; seed a xorshift generator
// from the clock. This is not cryptographically secure and is documented
// as such: applications needing entropy must supply their own source.
var rngState uint64

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	rngState = uint64(nanotime()) | 1
}

//go:linkname getRandomData runtime/goos.GetRandomData
func getRandomData(b []byte) {
	for i := range b {
		rngState ^= rngState << 13
		rngState ^= rngState >> 7
		rngState ^= rngState << 17
		b[i] = byte(rngState)
	}
}

// printk writes a byte to the L4Re console. The runtime emits output
// byte-wise; bytes are collected in a line buffer and flushed on newline
// or when full, one L4::Vcon::write per flush.
const printBufLen = svc.VconMaxWrite

var (
	printBuf [printBufLen]byte
	printLen int
)

//go:linkname printk runtime/goos.Printk
//go:nosplit
func printk(c byte) {
	printBuf[printLen] = c
	printLen++
	if c == '\n' || printLen == printBufLen {
		flushLog()
	}
}

//go:nosplit
func flushLog() {
	if printLen == 0 {
		return
	}
	svc.LogWrite(printBuf[:printLen])
	printLen = 0
}
