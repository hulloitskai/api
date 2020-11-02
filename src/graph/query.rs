use super::prelude::*;
use super::{Contact, Meta};
use crate::prelude::*;

use crate::models::Contact as ContactModel;

#[derive(ConstantObject)]
struct QueryConstants {
    me: Contact,
    meta: Meta,
}

impl QueryConstants {}

#[derive(CombinedObject)]
pub struct Query(QueryConstants);

impl Query {
    pub fn new(
        built: DateTime,
        version: Option<String>,
        contact: &ContactModel,
    ) -> Self {
        let constants = QueryConstants {
            me: Contact::new(contact),
            meta: Meta::new(built, version),
        };
        Query(constants)
    }
}
