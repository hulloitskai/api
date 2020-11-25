use crate::common::*;

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
pub struct BuildInfo {
    pub timestamp: DateTime,
    pub version: Option<String>,
}
