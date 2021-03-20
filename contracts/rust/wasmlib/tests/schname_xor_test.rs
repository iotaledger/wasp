use wasmlib::ScHname;

macro_rules! test_xor {
    ($test_name:ident, $lhs : expr, $rhs : expr, $expected :expr) => {
        #[test]
        fn $test_name(){
            let lhs : ScHname = ScHname::from_bytes($lhs);
            let rhs : ScHname = ScHname::from_bytes($rhs);

            let actual_schname : ScHname = lhs ^ rhs;
            let actual_bytes : Vec<u8> = actual_schname.to_bytes();
            
            let expected_bytes : Vec<u8> = $expected;
            assert_eq!(expected_bytes, actual_bytes);
        }
    };
}

// First byte
test_xor!(xor_0001_to_0000, &*vec![0, 0, 0, 1], &*vec![0, 0, 0, 0], vec![0, 0, 0, 1]);
test_xor!(xor_0001_to_0001, &*vec![0, 0, 0, 1], &*vec![0, 0, 0, 1], vec![0, 0, 0, 0]);
test_xor!(xor_0000_to_0001, &*vec![0, 0, 0, 0], &*vec![0, 0, 0, 1], vec![0, 0, 0, 1]);

// Second byte
test_xor!(xor_0010_to_0000, &*vec![0, 0, 1, 0], &*vec![0, 0, 0, 0], vec![0, 0, 1, 0]);
test_xor!(xor_0010_to_0010, &*vec![0, 0, 1, 0], &*vec![0, 0, 1, 0], vec![0, 0, 0, 0]);
test_xor!(xor_0000_to_0010, &*vec![0, 0, 0, 0], &*vec![0, 0, 1, 0], vec![0, 0, 1, 0]);

// Third byte
test_xor!(xor_0100_to_0000, &*vec![0, 1, 0, 0], &*vec![0, 0, 0, 0], vec![0, 1, 0, 0]);
test_xor!(xor_0100_to_0100, &*vec![0, 1, 0, 0], &*vec![0, 1, 0, 0], vec![0, 0, 0, 0]);
test_xor!(xor_0000_to_0100, &*vec![0, 0, 0, 0], &*vec![0, 1, 0, 0], vec![0, 1, 0, 0]);

// Fourth byte
test_xor!(xor_1000_to_0000, &*vec![1, 0, 0, 0], &*vec![0, 0, 0, 0], vec![1, 0, 0, 0]);
test_xor!(xor_1000_to_1000, &*vec![1, 0, 0, 0], &*vec![1, 0, 0, 0], vec![0, 0, 0, 0]);
test_xor!(xor_0000_to_1000, &*vec![0, 0, 0, 0], &*vec![1, 0, 0, 0], vec![1, 0, 0, 0]);

// All equal
test_xor!(xor_0000_to_0000, &*vec![0, 0, 0, 0], &*vec![0, 0, 0, 0], vec![0, 0, 0, 0]);
test_xor!(xor_1111_to_1111, &*vec![1, 1, 1, 1], &*vec![1, 1, 1, 1], vec![0, 0, 0, 0]);
