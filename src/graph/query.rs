use super::common::*;
use super::{BuildInfo, Contact};

use crate::meta::BuildInfo as BuildInfoModel;
use crate::models::Contact as ContactModel;

pub struct Query;

#[Object]
impl Query {
    async fn build(&self, ctx: &Context<'_>) -> FieldResult<BuildInfo> {
        let model = ctx.data::<BuildInfoModel>()?;
        let build = BuildInfo::new(model.to_owned());
        Ok(build)
    }

    async fn me(&self, ctx: &Context<'_>) -> FieldResult<Contact> {
        let model = ctx.data::<ContactModel>()?;
        let contact = Contact::new(model.to_owned());
        Ok(contact)
    }
}
