use super::Email;
use crate::prelude::*;

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
pub struct Contact {
    pub name: String,
    pub about: Option<String>,
    pub email: Email,
    pub birthday: Date,
}
