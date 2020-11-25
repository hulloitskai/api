use crate::common::*;

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Status {
    Pass,
    Warn,
    Fail,
}

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
pub struct Health {
    status: Status,
}

impl Health {
    pub fn new(status: Status) -> Self {
        Health { status }
    }
}
