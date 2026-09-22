TamaGo - bare metal Go - L4Re user space support
=================================================

tamago | https://github.com/usbarmory/tamago
fork   | https://github.com/mynetz/tamago (branch l4re-native)

Works with the unmodified tamago-go toolchain.

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
  * `runtime.go`, `cpuinit_$GOARCH.s`, `sig_$GOARCH.{go,s}` — the
    `runtime/goos` overlay implementation and the startup signal handler.

Startup
=======

`cpuinit` (the ELF entry, reached through `runtime/goos.CPUInit`) decodes the
auxiliary vector left on the initial stack by the loader (`AT_L4_ENV`,
`AT_L4_KIP`), switches to a small bootstrap stack and calls `earlyInit`, which
allocates the Go heap as an L4Re dataspace and attaches it to the address
space. The region is published as `runtime/goos.RamStart`/`RamSize`, the
runtime stack is placed at its top and control passes to the Go runtime.

Two operations the `GOOS=tamago` runtime performs during startup assume it
runs either on bare metal or as a Linux process: on amd64 `runtime.settls`
installs the TLS base with Linux' `arch_prctl(ARCH_SET_FS)` system call, on
arm64 `internal/cpu` reads the privileged `ID_AA64*` registers with `MRS`.
Neither works under Fiasco. Rather than patching the toolchain, `earlyInit`
registers a POSIX signal handler with the in-task server `l4re_itas`
(`svc.Sigaction`), which is the exception handler of every L4Re thread and
delivers signals on the faulting thread: on amd64 the stray `arch_prctl`
is a harmless failed IPC and the following `-8(FS)` access faults, the
SIGSEGV handler (`sig_amd64.s`) then installs FS_BASE with the L4 thread
system call; on arm64 the SIGILL handler (`sig_arm64.s`) emulates the
three `MRS` instructions with 0 (no optional features). The handler is
removed in `Hwinit1`; afterwards faults get the default treatment (itas
diagnostic dump, task termination).

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
