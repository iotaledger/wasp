extern crate libc;

use libc::c_int;

use std::convert::From;
use std::error;
use std::ffi::CStr;
use std::fmt;
use std::io;
use std::result;
use std::str;
use nanomsg_sys::*;
use nanomsg_sys::posix_consts::*;

pub type Result<T> = result::Result<T, Error>;

#[derive(Clone, Copy, Eq, PartialEq, Debug)]
pub enum Error {
    Unknown = 0 as isize,
    OperationNotSupported = ENOTSUP as isize,
    ProtocolNotSupported = EPROTONOSUPPORT as isize,
    NoBufferSpace = ENOBUFS as isize,
    NetworkDown = ENETDOWN as isize,
    AddressInUse = EADDRINUSE as isize,
    AddressNotAvailable = EADDRNOTAVAIL as isize,
    ConnectionRefused = ECONNREFUSED as isize,
    OperationNowInProgress = EINPROGRESS as isize,
    NotSocket = ENOTSOCK as isize,
    AddressFamilyNotSupported = EAFNOSUPPORT as isize,
    #[cfg(not(target_os = "openbsd"))]
    WrongProtocol = EPROTO as isize,
    #[cfg(target_os = "openbsd")]
    WrongProtocol = EPROTOTYPE as isize,
    TryAgain = EAGAIN as isize,
    BadFileDescriptor = EBADF as isize,
    InvalidInput = EINVAL as isize,
    TooManyOpenFiles = EMFILE as isize,
    BadAddress = EFAULT as isize,
    PermissionDenied = EACCESS as isize,
    NetworkReset = ENETRESET as isize,
    NetworkUnreachable = ENETUNREACH as isize,
    HostUnreachable = EHOSTUNREACH as isize,
    NotConnected = ENOTCONN as isize,
    MessageTooLong = EMSGSIZE as isize,
    TimedOut = ETIMEDOUT as isize,
    ConnectionAborted = ECONNABORTED as isize,
    ConnectionReset = ECONNRESET as isize,
    ProtocolNotAvailable = ENOPROTOOPT as isize,
    AlreadyConnected = EISCONN as isize,
    SocketTypeNotSupported = ESOCKTNOSUPPORT as isize,
    Terminating = ETERM as isize,
    NameTooLong = ENAMETOOLONG as isize,
    NoDevice = ENODEV as isize,
    FileStateMismatch = EFSM as isize,
    Interrupted = EINTR as isize,
}

impl Error {
    pub fn to_raw(&self) -> c_int {
        *self as c_int
    }

    pub fn from_raw(raw: c_int) -> Error {
        match raw {
            ENOTSUP => Error::OperationNotSupported,
            EPROTONOSUPPORT => Error::ProtocolNotSupported,
            ENOBUFS => Error::NoBufferSpace,
            ENETDOWN => Error::NetworkDown,
            EADDRINUSE => Error::AddressInUse,
            EADDRNOTAVAIL => Error::AddressNotAvailable,
            ECONNREFUSED => Error::ConnectionRefused,
            EINPROGRESS => Error::OperationNowInProgress,
            ENOTSOCK => Error::NotSocket,
            EAFNOSUPPORT => Error::AddressFamilyNotSupported,
            #[cfg(not(target_os = "openbsd"))]
            EPROTO => Error::WrongProtocol,
            #[cfg(target_os = "openbsd")]
            EPROTOTYPE => Error::WrongProtocol,
            EAGAIN => Error::TryAgain,
            EBADF => Error::BadFileDescriptor,
            EINVAL => Error::InvalidInput,
            EMFILE => Error::TooManyOpenFiles,
            EFAULT => Error::BadAddress,
            EACCESS => Error::PermissionDenied,
            ENETRESET => Error::NetworkReset,
            ENETUNREACH => Error::NetworkUnreachable,
            EHOSTUNREACH => Error::HostUnreachable,
            ENOTCONN => Error::NotConnected,
            EMSGSIZE => Error::MessageTooLong,
            ETIMEDOUT => Error::TimedOut,
            ECONNABORTED => Error::ConnectionAborted,
            ECONNRESET => Error::ConnectionReset,
            ENOPROTOOPT => Error::ProtocolNotAvailable,
            EISCONN => Error::AlreadyConnected,
            ESOCKTNOSUPPORT => Error::SocketTypeNotSupported,
            ETERM => Error::Terminating,
            ENAMETOOLONG => Error::NameTooLong,
            ENODEV => Error::NoDevice,
            EFSM => Error::FileStateMismatch,
            EINTR => Error::Interrupted,
            _ => Error::Unknown,
        }
    }
}

impl error::Error for Error {}

impl From<io::Error> for Error {
    fn from(err: io::Error) -> Error {
        match err.kind() {
            io::ErrorKind::PermissionDenied => Error::PermissionDenied,
            io::ErrorKind::ConnectionRefused => Error::ConnectionRefused,
            io::ErrorKind::ConnectionReset => Error::ConnectionReset,
            io::ErrorKind::ConnectionAborted => Error::ConnectionAborted,
            io::ErrorKind::NotConnected => Error::NotConnected,
            io::ErrorKind::AddrInUse => Error::AddressInUse,
            io::ErrorKind::AddrNotAvailable => Error::AddressNotAvailable,
            io::ErrorKind::AlreadyExists => Error::AlreadyConnected,
            io::ErrorKind::WouldBlock => Error::TryAgain,
            io::ErrorKind::InvalidInput => Error::InvalidInput,
            io::ErrorKind::TimedOut => Error::TimedOut,
            io::ErrorKind::Interrupted => Error::Interrupted,
            _ => Error::Unknown,
        }
    }
}

impl From<Error> for io::Error {
    fn from(err: Error) -> io::Error {
        let as_std_error: &dyn error::Error = &err;
        let description = as_std_error.to_string();
        match err {
            Error::PermissionDenied => io::Error::new(io::ErrorKind::PermissionDenied, description),
            Error::ConnectionRefused => {
                io::Error::new(io::ErrorKind::ConnectionRefused, description)
            }
            Error::ConnectionReset => io::Error::new(io::ErrorKind::ConnectionReset, description),
            Error::ConnectionAborted => {
                io::Error::new(io::ErrorKind::ConnectionAborted, description)
            }
            Error::NotConnected => io::Error::new(io::ErrorKind::NotConnected, description),
            Error::AddressInUse => io::Error::new(io::ErrorKind::AddrInUse, description),
            Error::AddressNotAvailable => {
                io::Error::new(io::ErrorKind::AddrNotAvailable, description)
            }
            Error::AlreadyConnected => io::Error::new(io::ErrorKind::AlreadyExists, description),
            Error::TryAgain => io::Error::new(io::ErrorKind::WouldBlock, description),
            Error::InvalidInput => io::Error::new(io::ErrorKind::InvalidInput, description),
            Error::TimedOut => io::Error::new(io::ErrorKind::TimedOut, description),
            Error::Interrupted => io::Error::new(io::ErrorKind::Interrupted, description),
            _ => io::Error::new(io::ErrorKind::Other, description),
        }
    }
}

impl fmt::Display for Error {
    fn fmt(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        let description = unsafe {
            let nn_errno = *self as c_int;
            let c_ptr: *const libc::c_char = nn_strerror(nn_errno);
            let c_str = CStr::from_ptr(c_ptr);
            let bytes = c_str.to_bytes();
            str::from_utf8(bytes).unwrap_or("Error")
        };
        write!(formatter, "{}", description)
    }
}

pub fn last_nano_error() -> Error {
    let nn_errno = unsafe { nn_errno() };

    Error::from_raw(nn_errno)
}

#[cfg(test)]
#[allow(unused_must_use)]
mod tests {
    use super::Error;
    use libc;
    use nanomsg_sys::*;
    use std::convert::From;
    use std::io;

    fn assert_convert_error_code_to_error(error_code: libc::c_int, expected_error: Error) {
        let converted_error = Error::from_raw(error_code);
        assert_eq!(expected_error, converted_error)
    }

    #[test]
    fn can_convert_error_code_to_error() {
        assert_convert_error_code_to_error(ENOTSUP, Error::OperationNotSupported);
        assert_convert_error_code_to_error(
            EPROTONOSUPPORT,
            Error::ProtocolNotSupported,
        );
        assert_convert_error_code_to_error(EADDRINUSE, Error::AddressInUse);
        assert_convert_error_code_to_error(EHOSTUNREACH, Error::HostUnreachable);
    }

    fn check_error_kind_match(nano_err: Error, io_err_kind: io::ErrorKind) {
        let io_err: io::Error = From::from(nano_err);

        assert_eq!(io_err_kind, io_err.kind())
    }

    #[test]
    fn nano_err_can_be_converted_to_io_err() {
        check_error_kind_match(Error::TimedOut, io::ErrorKind::TimedOut);
        check_error_kind_match(Error::PermissionDenied, io::ErrorKind::PermissionDenied);
        check_error_kind_match(Error::ConnectionRefused, io::ErrorKind::ConnectionRefused);
        check_error_kind_match(Error::OperationNotSupported, io::ErrorKind::Other);
        check_error_kind_match(Error::NotConnected, io::ErrorKind::NotConnected);
        check_error_kind_match(Error::Interrupted, io::ErrorKind::Interrupted);
    }

    #[test]
    fn nano_err_can_be_converted_from_io_err() {
        let io_err = io::Error::new(io::ErrorKind::TimedOut, "Timed out");
        let nano_err: Error = From::from(io_err);

        assert_eq!(Error::TimedOut, nano_err)
    }
}
