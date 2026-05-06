TamaGo - bare metal Go - L4Re userspace support
================================================

tamago | https://github.com/usbarmory/tamago
Fork:    https://github.com/mynetz/tamago (branch l4re-native)

Introduction
============

Package `l4re_user` provides support for `GOOS=tamago` Go programs that run as
**native L4Re tasks** on the Fiasco microkernel. The package follows the
para-virtualisation pattern of [go-boot](https://github.com/usbarmory/go-boot)
which implements the same idea for UEFI runtime services.

The goal is to let `GOOS=tamago` Go programs be loaded by ned/moe and call
into L4Re services (initially: the `log`/Vcon capability for stdout) without
linking against L4Re's libl4re or libc, and without cgo.

Status
======

Iteration 2b (current): minimal Vcon-only build.

  * `cpuinit` saves the L4Re initial environment from auxv (AT_L4_ENV =
    0xf1) and the KIP from AT_L4_KIP (0xf2), then jumps to the Go runtime
    entry.
  * `printk` writes single bytes to the parent-supplied `log` capability via
    a small line buffer and one `L4_VCON_WRITE_OP` IPC per flush.
  * `nanotime` is currently stubbed; calls to `time.Now` are not yet
    expected to be functional.
  * No exit; the program is expected to park in a goroutine forever.

Future iterations are expected to grow KIP-clock support, dataspace-based
heap allocation, IRQ delivery, and Go-side bindings for additional L4Re
service capabilities.

Compiling
=========

The package overrides the default `runtime/goos` symbols `cpuinit`,
`ramStart`, `ramSize`, and `printk`. The application must therefore compile
with the matching tamago build tags:

    -tags linkcpuinit,linkramstart,linkramsize,linkprintk

and the linker entry override:

    -ldflags "-E cpuinit -T <load-addr> -R 0x1000"

The load address must match the L4Re bootstrap configuration for the task.
A typical value for the QEMU q35 setup is `0x10010000`, in line with the
upstream tamago/cloud_hypervisor and microvm boards.
