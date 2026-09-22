TamaGo - bare metal Go - L4Re user space support
=================================================

tamago | https://github.com/usbarmory/tamago
fork   | https://github.com/mynetz/tamago (branch l4re-native)

Introduction
============

Package `l4re_user` provides support for `GOOS=tamago` Go programs running as
**native L4Re tasks** on the Fiasco microkernel, started by the L4Re loader
(ned/moe) like any other L4Re application.

The package follows the pattern of [go-boot](https://github.com/usbarmory/go-boot)
for UEFI: no L4Re C/C++ libraries are linked and no cgo is used. The L4 system
call is issued from Go assembly and the L4Re services are spoken to with
hand-marshalled IPC:

  * `sys/` — kernel ABI: IPC system call (amd64 `SYSCALL`, arm64 `SVC`),
    UTCB access, message tags, KIP clock, constants.
  * `svc/` — L4Re service bindings: initial environment (`l4re_env_t`),
    memory allocator (`Mem_alloc::alloc`), region map (`Rm::attach`),
    console (`Vcon::write`), parent (`Parent::signal`).
  * `runtime.go`, `cpuinit_$GOARCH.s`, `settls_amd64.s` — the
    `runtime/goos` overlay implementation.

Startup
=======

`cpuinit` (the ELF entry, reached through `runtime/goos.CPUInit`) decodes the
auxiliary vector left on the initial stack by the loader (`AT_L4_ENV`,
`AT_L4_KIP`), switches to a small bootstrap stack and calls `earlyInit`, which
allocates the Go heap as an L4Re dataspace and attaches it to the address
space. The region is published as `runtime/goos.RamStart`/`RamSize`, the
runtime stack is placed at its top and control passes to the Go runtime.

On amd64 the runtime asks the environment to install the TLS base for the
initial thread (`runtime/goos.SetTLS`, see the fork's tamago-go branch); the
L4Re implementation issues the thread system call
`L4_THREAD_AMD64_SET_SEGMENT_BASE_OP` on the calling thread. arm64 keeps `g`
in a register and needs no such step.

Console output goes through `L4::Vcon::write` to the task's `log`
capability, time comes from the kernel's KIP clock stub (`read_ns`), and
returning from `main` exits the task through `L4Re::Parent::signal(0, code)`.

Supported architectures: amd64, arm64.

Compiling
=========

    GOOS=tamago GOARCH=amd64 GOOSPKG=github.com/usbarmory/tamago \
      tamago build -ldflags "-T 0x10010000 -R 0x1000" -o hello-go .

    GOOS=tamago GOARCH=arm64 GOOSPKG=github.com/usbarmory/tamago \
      tamago build -ldflags "-T 0x10010000 -R 0x1000" -o hello-go .

The link address must not collide with the L4Re in-task loader
(`l4re_itas`, 0x70000000 on amd64, 0xc0000000 on arm64).

Limitations
===========

  * Single threaded (no `runtime/goos.Task`); `GOMAXPROCS=1`.
  * `GetRandomData` is seeded from the clock and is not cryptographically
    secure.
  * Fixed heap size (`HeapSize`), allocated at startup.
