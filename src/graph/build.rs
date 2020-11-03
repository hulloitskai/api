use super::prelude::*;
use crate::models::BuildInfo as BuildInfoModel;

#[derive(ConstantObject)]
pub struct BuildInfo {
    timestamp: DateTime,
    version: Option<String>,
}

impl BuildInfo {
    pub fn new(model: BuildInfoModel) -> Self {
        let BuildInfoModel { timestamp, version } = model;
        BuildInfo { timestamp, version }
    }
}
