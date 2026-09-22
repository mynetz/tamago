// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sys

// Capability selectors (l4sys/include/consts.h).
//
// A capability selector carries the slot index in the bits above CapShift
// and IPC operation flags in the low bits.
const (
	// CapShift is the bit position of the slot index (L4_CAP_SHIFT).
	CapShift = 12

	// CapStride is the distance between two consecutive slots
	// (L4_CAP_OFFSET).
	CapStride uint64 = 1 << CapShift

	// InvalidCap is the invalid selector (L4_INVALID_CAP). When used as
	// IPC destination the kernel resolves it to the calling thread
	// (`self`), which is the idiom for operations on the current thread
	// such as installing the TLS segment base.
	InvalidCap uint64 = 0xFFFFFFFFFFFFF800 // ~0 << (CapShift - 1)

	// InvalidCapBit is the bit marking a selector as invalid
	// (L4_INVALID_CAP_BIT).
	InvalidCapBit uint64 = 1 << (CapShift - 1)
)

// IPC operation flags, OR-ed into the destination selector
// (L4_SYSF_* in l4sys/include/consts.h).
const (
	SysfSend  uint64 = 0x01
	SysfRecv  uint64 = 0x02
	SysfCall  uint64 = SysfSend | SysfRecv
	SysfWait  uint64 = SysfRecv | 0x04 // L4_SYSF_OPEN_WAIT
	SysfReply uint64 = 0x08
)

// Timeouts (l4_timeout_t.raw). Zero means never in both phases.
const (
	TimeoutNever uint64 = 0
)

// Kernel protocol labels (l4sys/include/types.h, L4_PROTO_*).
const (
	ProtoIrq       int64 = -1
	ProtoPageFault int64 = -2
	ProtoSched     int64 = -3
	ProtoTask      int64 = -5
	ProtoThread    int64 = -12
	ProtoLog       int64 = -13
	ProtoFactory   int64 = -15
)

// L4Re service protocol labels (l4re/include/protocols.h).
const (
	ProtoDataspace int64 = 0x4000
	ProtoNamespace int64 = 0x4001
	ProtoParent    int64 = 0x4002
	ProtoGoos      int64 = 0x4003
	ProtoRm        int64 = 0x4005
	ProtoEvent     int64 = 0x4006
	ProtoInhibitor int64 = 0x4007
	ProtoDmaSpace  int64 = 0x4008
	ProtoMmioSpace int64 = 0x4009
	ProtoItas      int64 = 0x400a
	ProtoMemAlloc  int64 = 0x400b
)

// Message tag flag bits (low 16 bits, masked with 0xf000).
const (
	MsgtagTransferFPU uint64 = 0x1000
	MsgtagSchedule    uint64 = 0x2000
	MsgtagPropagate   uint64 = 0x4000
	MsgtagError       uint64 = 0x8000
)

// Variable argument type tags used by factory create messages
// (L4_VARG_TYPE_* in l4sys/include/types.h). A varg occupies two message
// registers: `type | (size << 16)` followed by the value.
const (
	VargTypeNil    uint64 = 0x00
	VargTypeUmword uint64 = 0x01
	VargTypeMword  uint64 = 0x81
	VargTypeString uint64 = 0x02
	VargTypeFpage  uint64 = 0x03
)

// Message item flags (l4sys/include/consts.h).
const (
	// ItemMap identifies a send item as map item (L4_ITEM_MAP).
	ItemMap uint64 = 0x08
	// RcvItemSingleCap makes a receive buffer accept exactly one
	// capability (L4_RCV_ITEM_SINGLE_CAP).
	RcvItemSingleCap uint64 = ItemMap | 0x02
)

// Flexpage encoding (l4sys/include/__l4_fpage.h).
const (
	FpageRightsShift = 0
	FpageTypeShift   = 4
	FpageSizeShift   = 6
	FpageAddrShift   = 12

	FpageTypeObj uint64 = 3

	// Capability rights (L4_CAP_FPAGE_*).
	CapFpageW  uint64 = 0x1
	CapFpageS  uint64 = 0x2
	CapFpageR  uint64 = 0x4
	CapFpageD  uint64 = 0x8
	CapFpageRW uint64 = CapFpageR | CapFpageW
)

// UTCB layout (identical on amd64 and arm64, see
// l4sys/include/ARCH-*/arch/utcb.h).
const (
	UtcbMsgRegsOffset    = 0
	UtcbBufRegsOffset    = 64 * 8
	UtcbThreadRegsOffset = 123 * 8

	// UtcbGenericDataSize is the number of message registers usable for
	// untyped words (L4_UTCB_GENERIC_DATA_SIZE).
	UtcbGenericDataSize = 63
	// UtcbGenericBuffersSize is the number of buffer registers following
	// the buffer descriptor register (L4_UTCB_GENERIC_BUFFERS_SIZE).
	UtcbGenericBuffersSize = 58
)

// L4 error codes (l4sys/include/err.h). They are returned negated in the
// label of a reply tag.
const (
	EOK          = 0
	EPerm        = 1
	ENoEnt       = 2
	EIO          = 5
	ENoMem       = 12
	EFault       = 14
	EBusy        = 16
	EExist       = 17
	ENoDev       = 19
	EInval       = 22
	ERange       = 34
	ENoSys       = 38
	EBadProto    = 39
	EAddrNotAvl  = 99
	ENoReply     = 1000
	EMsgTooShort = 1001
	EMsgTooLong  = 1002
	EMsgMissArg  = 1003
	EIpcLo       = 2000
)

// IPC error codes found in TCR.error when a reply tag carries MsgtagError
// (l4sys/include/ipc.h).
const (
	IpcErrorMask   uint64 = 0x1f
	IpcSndErrMask  uint64 = 0x01
	IpcSeTimeout   uint64 = 0x02
	IpcReTimeout   uint64 = 0x03
	IpcENotExist   uint64 = 0x04
	IpcReCanceled  uint64 = 0x07
	IpcSeCanceled  uint64 = 0x06
	IpcReMsgCut    uint64 = 0x08
	IpcSeMsgCut    uint64 = 0x09
	IpcReMapFailed uint64 = 0x11
	IpcSeMapFailed uint64 = 0x10
)

// Kernel Info Page offsets (l4sys/include/kip.h). The KIP embeds
// kernel-provided code stubs that return the current clock; they follow
// the C calling convention, take no arguments and clobber only scratch
// registers.
const (
	KipOffsReadUs = 0x900
	KipOffsReadNs = 0x980
)
