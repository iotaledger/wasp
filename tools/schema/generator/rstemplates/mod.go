package rstemplates

var modRs = map[string]string{
	// *******************************
	"mod.rs": `
#![allow(unused_imports)]

pub use consts::*;
pub use contract::*;
pub use params::*;
pub use results::*;

pub mod consts;
pub mod contract;
pub mod params;
pub mod results;
`,
}
