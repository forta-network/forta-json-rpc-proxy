// SPDX-License-Identifier: No license
pragma solidity ^0.8.25;

interface ISecurityValidator {
    function executeCheckpoint(bytes32) external;
}

contract ProtectedContract {
    event SafelyExecuted();

    ISecurityValidator public validator;

    constructor(ISecurityValidator _validator) {
        validator = _validator;
    }

    function protectedFunc() public safeExecution {
        emit SafelyExecuted();
    }

    function nonProtectedFunc() public {}

    modifier safeExecution() {
        validator.executeCheckpoint(bytes32(0));
        _;
    }
}
