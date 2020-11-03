use crate::prelude::*;

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
pub struct Meta {
    pub built: DateTime,
    pub version: Option<String>,
}
