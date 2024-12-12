package service

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// BypassAddress is used as a flag which is recognized by the SecurityValidator during
// checkpoint execution.
var BypassAddress = common.HexToAddress("0x0000000000000000000000000000000000f01274")

// BypassCode is set as the BypassAddress code during state override so that flag is truthy.
var BypassCode = hexutil.MustDecode("0x10")

// AddFortaFirewallStateOverride adds Forta Firewall state override to make transaction simulation
// succeed. Without this state override, the transactions which try to execute a checkpoint will look
// like they revert and cause a confusing experience.
func AddFortaFirewallStateOverride(stateOverride *StateOverride) *StateOverride {
	if stateOverride == nil {
		stateOverride = &StateOverride{}
	}
	(*stateOverride)[BypassAddress] = OverrideAccount{
		Code: (*hexutil.Bytes)(&BypassCode),
	}
	return stateOverride
}
