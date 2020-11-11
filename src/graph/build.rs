use super::prelude::*;
use crate::build::Build;

#[derive(ConstantObject)]
pub struct BuildInfo {
    timestamp: DateTime,
    version: Option<String>,
}

impl BuildInfo {
    pub fn new(build: Build) -> Self {
        let Build { timestamp, version } = build;
        BuildInfo { timestamp, version }
    }
}
