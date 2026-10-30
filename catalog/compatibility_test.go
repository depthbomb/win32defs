package catalog_test

import (
	"testing"

	"github.com/depthbomb/win32defs/hresult"
	"github.com/depthbomb/win32defs/ntstatus"
	"github.com/depthbomb/win32defs/power"
)

func TestRetainedConstants(t *testing.T) {
	t.Parallel()

	checkDesktopConstant(t, "hresult", "WU_E_STANDBY_ACTIVITY_NOT_ALLOWED", hresult.WU_E_STANDBY_ACTIVITY_NOT_ALLOWED, -2145124249, hresult.Parse, hresult.Names)
	checkDesktopConstant(t, "ntstatus", "STATUS_ADMINLESS_ACCESS_DENIED", ntstatus.STATUS_ADMINLESS_ACCESS_DENIED, 0xC000A204, ntstatus.Parse, ntstatus.Names)
	checkDesktopConstant(t, "power", "MonitorRequestReasonAusterityBatteryDrain", power.MonitorRequestReasonAusterityBatteryDrain, 55, power.Parse, power.Names)
}
