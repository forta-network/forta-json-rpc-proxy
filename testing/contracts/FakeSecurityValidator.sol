// SPDX-License-Identifier: No license
pragma solidity ^0.8.25;

/// @notice A fake implementation of the validator for testing.
contract FakeSecurityValidator {
    address constant BYPASS_FLAG = 0x0000000000000000000000000000000000f01274;

    error AttestationNotFound();

    bool public enabled;

    function enable() public {
        enabled = true;
    }

    function executeCheckpoint(bytes32) public {
        /// Throw the same error as the real validator.
        if (!enabled && BYPASS_FLAG.code.length == 0) revert AttestationNotFound();
        /// Consume the "enabled" state.
        enabled = false;
    }
}
