use super::prelude::*;
use super::{Contact, Meta};
use crate::models::{Contact as ContactModel, Meta as MetaModel};

pub struct Query;

#[ResolverObject]
impl Query {
    async fn meta(&self, ctx: &Context<'_>) -> FieldResult<Meta> {
        let meta = ctx.data::<MetaModel>()?;
        Ok(Meta::new(meta.to_owned()))
    }

    async fn me(&self, ctx: &Context<'_>) -> FieldResult<Contact> {
        let contact = ctx.data::<ContactModel>()?;
        Ok(Contact::new(contact.to_owned()))
    }
}
