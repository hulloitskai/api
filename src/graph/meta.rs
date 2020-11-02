use super::prelude::*;
use crate::prelude::*;

#[derive(ConstantObject)]
struct MetaConstants {
    built: DateTime,
    version: Option<String>,
}

#[derive(CombinedObject)]
pub struct Meta(MetaConstants);

impl Meta {
    pub fn new(built: DateTime, version: Option<String>) -> Self {
        let constants = MetaConstants { built, version };
        Meta(constants)
    }
}
