#![no_std]

use core::panic::PanicInfo;

#[panic_handler]
fn panic(_info: &PanicInfo) -> ! {
    loop {}
}

extern "C" {
    fn _log_from_wasm(msg: *const u8, len: usize);
}

#[no_mangle]
pub extern "C" fn app_main() -> i32 {
    let message = "This is log message from WebAssembly printed in Go host!!".as_bytes();

    unsafe {
        _log_from_wasm(message.as_ptr(), message.len());
    }

    return 0;
}