use crate::prelude::*;

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
pub struct Build {
    pub timestamp: DateTime,
    pub version: Option<String>,
}
