package jobobject_test

import (
	"runtime"
	"syscall"
	"testing"
	"unsafe"

	"github.com/depthbomb/win32defs/jobobject"
	"github.com/depthbomb/win32defs/process"
)

func TestNativeJobLimits(t *testing.T) {
	t.Parallel()

	kernel := syscall.NewLazyDLL("kernel32.dll")
	create := kernel.NewProc("CreateJobObjectW")
	set := kernel.NewProc("SetInformationJobObject")
	query := kernel.NewProc("QueryInformationJobObject")
	handle, _, err := create.Call(0, 0)
	if handle == 0 {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := syscall.CloseHandle(syscall.Handle(handle)); err != nil {
			t.Error(err)
		}
	})

	limits := jobobject.NewJOBOBJECT_EXTENDED_LIMIT_INFORMATION()
	basic := limits.GetBasicLimitInformation()
	flags := jobobject.JOB_OBJECT_LIMIT(jobobject.JOB_OBJECT_LIMIT_ACTIVE_PROCESS | jobobject.JOB_OBJECT_LIMIT_PROCESS_MEMORY)
	basic.SetLimitFlags(flags)
	basic.SetActiveProcessLimit(3)
	limits.SetProcessMemoryLimit(64 * 1024 * 1024)
	result, _, err := set.Call(handle, uintptr(jobobject.JobObjectExtendedLimitInformation), uintptr(limits.Pointer()), uintptr(jobobject.JOBOBJECT_EXTENDED_LIMIT_INFORMATIONSize))
	runtime.KeepAlive(limits)
	if result == 0 {
		t.Fatal(err)
	}

	read := jobobject.NewJOBOBJECT_EXTENDED_LIMIT_INFORMATION()
	var written uint32
	result, _, err = query.Call(handle, uintptr(jobobject.JobObjectExtendedLimitInformation), uintptr(read.Pointer()), uintptr(jobobject.JOBOBJECT_EXTENDED_LIMIT_INFORMATIONSize), uintptr(unsafe.Pointer(&written)))
	runtime.KeepAlive(read)
	if result == 0 {
		t.Fatal(err)
	}

	if written != uint32(jobobject.JOBOBJECT_EXTENDED_LIMIT_INFORMATIONSize) || read.GetBasicLimitInformation().GetLimitFlags() != flags || read.GetBasicLimitInformation().GetActiveProcessLimit() != 3 || read.GetProcessMemoryLimit() != limits.GetProcessMemoryLimit() {
		t.Fatalf("extended job limits did not round-trip: bytes=%d flags=%#x processes=%d memory=%d", written, read.GetBasicLimitInformation().GetLimitFlags(), read.GetBasicLimitInformation().GetActiveProcessLimit(), read.GetProcessMemoryLimit())
	}

	standalone := jobobject.NewJOBOBJECT_BASIC_LIMIT_INFORMATION()
	result, _, err = query.Call(handle, uintptr(jobobject.JobObjectBasicLimitInformation), uintptr(standalone.Pointer()), uintptr(jobobject.JOBOBJECT_BASIC_LIMIT_INFORMATIONSize), 0)
	runtime.KeepAlive(standalone)
	if result == 0 || standalone.GetActiveProcessLimit() != 3 {
		t.Fatalf("standalone basic limits: result=%d error=%v processes=%d", result, err, standalone.GetActiveProcessLimit())
	}

	counters := process.NewIO_COUNTERS()
	counters.SetReadOperationCount(1 << 63)
	counters.SetWriteTransferCount(1 << 62)
	limits.SetIoInfo(counters)
	if limits.GetIoInfo().GetReadOperationCount() != 1<<63 || limits.GetIoInfo().GetWriteTransferCount() != 1<<62 {
		t.Fatal("nested I/O counters lost their high bits")
	}
}

func TestNativeProcessIOCounters(t *testing.T) {
	t.Parallel()

	kernel := syscall.NewLazyDLL("kernel32.dll")
	current, _, _ := kernel.NewProc("GetCurrentProcess").Call()
	counters := process.NewIO_COUNTERS()
	for index := range counters.Bytes() {
		counters.Bytes()[index] = 0xff
	}
	result, _, err := kernel.NewProc("GetProcessIoCounters").Call(current, uintptr(counters.Pointer()))
	runtime.KeepAlive(counters)
	if result == 0 {
		t.Fatal(err)
	}

	for _, value := range []uint64{counters.GetReadOperationCount(), counters.GetWriteOperationCount(), counters.GetOtherOperationCount(), counters.GetReadTransferCount(), counters.GetWriteTransferCount(), counters.GetOtherTransferCount()} {
		if value == ^uint64(0) {
			t.Fatal("Windows did not populate an I/O counter")
		}
	}
}
