use super::prelude::*;
use super::{BuildInfo, Contact};

use crate::build::Build;
use crate::models::Contact as ContactModel;

pub struct Query;

#[ResolverObject]
impl Query {
    async fn build(&self, ctx: &Context<'_>) -> FieldResult<BuildInfo> {
        let model = ctx.data::<Build>()?;
        let build = BuildInfo::new(model.to_owned());
        Ok(build)
    }

    async fn me(&self, ctx: &Context<'_>) -> FieldResult<Contact> {
        let contact = ctx.data::<ContactModel>()?;
        Ok(Contact::new(contact.to_owned()))
    }
}
