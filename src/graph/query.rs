use super::prelude::*;
use super::{BuildInfo, Contact};
use crate::models::{BuildInfo as BuildInfoModel, Contact as ContactModel};

pub struct Query;

#[ResolverObject]
impl Query {
    async fn build(&self, ctx: &Context<'_>) -> FieldResult<BuildInfo> {
        let model = ctx.data::<BuildInfoModel>()?;
        let build = BuildInfo::new(model.to_owned());
        Ok(build)
    }

    async fn me(&self, ctx: &Context<'_>) -> FieldResult<Contact> {
        let contact = ctx.data::<ContactModel>()?;
        Ok(Contact::new(contact.to_owned()))
    }
}
