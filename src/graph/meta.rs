use super::prelude::*;
use crate::models::Meta as MetaModel;

#[derive(ConstantObject)]
struct MetaConstants {
    built: DateTime,
    version: Option<String>,
}

#[derive(CombinedObject)]
pub struct Meta(MetaConstants);

impl Meta {
    pub fn new(MetaModel { built, version }: MetaModel) -> Self {
        let constants = MetaConstants { built, version };
        Meta(constants)
    }
}
