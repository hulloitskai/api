use super::Email;
use crate::prelude::*;

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
pub struct Contact {
    pub first_name: String,
    pub last_name: String,
    pub email: Email,
    pub about: Option<String>,
    pub birthday: Date,
}

impl Contact {
    pub fn name(&self) -> String {
        format!("{} {}", self.first_name, self.last_name)
    }
}
