use super::host::get_key;

static mut ID_LENGTH: i32 = 0;
static mut ID_LOG: i32 = 0;
static mut ID_TRACE: i32 = 0;

pub fn key_length() -> i32 {
    unsafe {
        if ID_LENGTH == 0 {
            ID_LENGTH = get_key("length");
        }
        ID_LENGTH
    }
}

pub fn key_log() -> i32 {
    unsafe {
        if ID_LOG == 0 {
            ID_LOG = get_key("log");
        }
        ID_LOG
    }
}

pub fn key_trace() -> i32 {
    unsafe {
        if ID_TRACE == 0 {
            ID_TRACE = get_key("trace");
        }
        ID_TRACE
    }
}
