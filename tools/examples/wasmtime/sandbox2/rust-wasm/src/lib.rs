#![no_std]

use core::panic::PanicInfo;

#[panic_handler]
fn panic(_info: &PanicInfo) -> ! {
    loop {}
}

extern "C" {
    fn _publish(msg: *const u8, len: usize);
}

#[no_mangle]
pub extern "C" fn entry_point1() -> i32 {
    let message = "This is log message from WebAssembly pubished in Go host!!".as_bytes();

    unsafe {
        _publish(message.as_ptr(), message.len());
    }

    return 0;
}