// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract RevertTest {
    uint32 public count = 0;

    function incrementThenRevert() public {
        count = count + 1;
        revert(); // should revert the count increment
    }

    function selfCallRevert() public {
        bool success;
        try this.incrementThenRevert() {
            success = true;
        } catch {
            success = false;
        }
        require(!success);
        require(count == 0);
    }
}
