use crate::common::*;

use chrono::FixedOffset;

pub struct Context {
    pub timestamp: DateTime<FixedOffset>,
    pub version: Option<String>,
}
