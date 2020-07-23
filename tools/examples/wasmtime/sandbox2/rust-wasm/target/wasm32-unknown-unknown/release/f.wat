(module
  (type (;0;) (func (param i32 i32 i64)))
  (type (;1;) (func (param i32 i32)))
  (type (;2;) (func (result i32)))
  (import "env" "SBSetInt64" (func $SBSetInt64 (type 0)))
  (import "env" "SBPublish" (func $SBPublish (type 1)))
  (func $entry_point1 (type 2) (result i32)
    i32.const 1048602
    i32.const 7
    i64.const 1
    call $SBSetInt64
    i32.const 1048576
    i32.const 26
    call $SBPublish
    i32.const 0)
  (table (;0;) 1 1 funcref)
  (memory (;0;) 17)
  (global (;0;) (mut i32) (i32.const 1048576))
  (global (;1;) i32 (i32.const 1048609))
  (global (;2;) i32 (i32.const 1048609))
  (export "memory" (memory 0))
  (export "entry_point1" (func $entry_point1))
  (export "__data_end" (global 1))
  (export "__heap_base" (global 2))
  (data (;0;) (i32.const 1048576) "Value has been incrementedcounter"))
