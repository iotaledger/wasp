#![no_std]

use core::panic::PanicInfo;

#[panic_handler]
fn panic(_info: &PanicInfo) -> ! {
    loop {}
}

extern "C" {
    fn SBPublish(msg: *const u8, len: usize);
    // fn SBGetInt64(name: *const u8, strlen: usize) -> (i64, i32);
    fn SBSetInt64(name: *const u8, strlen: usize, value: i64);
}

#[no_mangle]
pub extern "C" fn entry_point1() -> i32 {
    let message = "Value has been set".as_bytes();
    let var_name = "counter".as_bytes();

    unsafe {
        // let val = SBGetInt64(var_name.as_ptr(), var_name.len());
        // let val = if val.1==0 {
        //       val.0
        // } else{
        //     0
        // };
        SBSetInt64(var_name.as_ptr(), var_name.len(), 1);
        SBPublish(message.as_ptr(), message.len());
        // wrong
        //SBPublish(buf, len);
    }
    return 0;
}