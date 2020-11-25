use super::Email;
use crate::common::*;

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
struct Person {
    id: Uuid,
    created: DateTime,
    updated: DateTime,
    name: String,
    email: Email,
}
