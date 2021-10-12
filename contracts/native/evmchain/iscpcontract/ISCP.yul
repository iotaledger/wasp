object "ISCPYul" {
  code {
    // Protection against sending Ether
    require(iszero(callvalue()))

    switch selector()

    case 0xe6c75c6b /* "triggerEvent(string)" */ {
      // triggerEvent("asd") ->
      //   00 e6c75c6b                                                         first 4 bytes of keccak("triggerEvent(string)")
      //   04 0000000000000000000000000000000000000000000000000000000000000020 location of data part
      //   24 0000000000000000000000000000000000000000000000000000000000000003 len("asd")
      //   44 6173640000000000000000000000000000000000000000000000000000000000 "asd"
      let size := calldataload(0x24)
      let offset := msize()
      calldatacopy(offset, 0x44, size)
      verbatim_2i_0o(hex"c0", offset, size)
    }

    case 0x47ce07cc /* "entropy()" */ {
      let e := verbatim_0i_1o(hex"c1")
      mstore(0, e)
      return(0, 0x20)
    }

    default {
      revert(0, 0)
    }

    function selector() -> s {
        s := div(calldataload(0), 0x100000000000000000000000000000000000000000000000000000000)
    }

    function require(condition) {
        if iszero(condition) { revert(0, 0) }
    }
  }
}
