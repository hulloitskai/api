use crate::prelude::*;

use std::env::{
    set_var as set_env_var, var as env_var, VarError as EnvVarError,
};
use std::io::ErrorKind as IoErrorKind;

use dotenv::dotenv;

pub const NAMESPACE: &str = "API";

pub fn key(name: &str) -> String {
    format!("{}_{}", NAMESPACE, name.to_uppercase())
}

pub fn var(name: &str) -> Result<String, EnvVarError> {
    let key = key(name);
    env_var(&key)
}

pub fn var_or(
    name: &str,
    default: impl Into<String>,
) -> Result<String, EnvVarError> {
    let key = key(name);
    env_var(&key).or_else(|error| match error {
        EnvVarError::NotPresent => Ok(default.into()),
        error => Err(error),
    })
}

pub fn load() -> Result<()> {
    if let Err(dotenv::Error::Io(error)) = dotenv() {
        if error.kind() != IoErrorKind::NotFound {
            return Err(error).context("load .env");
        }
    }

    // Configure backtraces.
    if None == var("BACKTRACE").ok() {
        set_env_var("RUST_BACKTRACE", "1")
    }

    // Configure logging.
    set_env_var(
        "RUST_LOG",
        match var("LOG").ok() {
            Some(s) => s,
            None => "warn,api=info".to_owned(),
        },
    );

    Ok(())
}
