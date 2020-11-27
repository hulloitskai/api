use crate::common::*;

pub mod serve;
pub use serve::*;

#[derive(Debug, Clap)]
pub enum Command {
    Serve(ServeCli),
}
