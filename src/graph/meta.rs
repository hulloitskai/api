use super::common::*;
use crate::meta::BuildInfo as BuildInfoModel;

#[derive(SimpleObject)]
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
