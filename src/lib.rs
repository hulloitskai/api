// Workaround for: https://github.com/rust-lang/rust/issues/64450
extern crate async_trait;
extern crate builder;

mod common {
    pub use std::collections::{HashMap as Map, HashSet as Set};
    pub use std::convert::{TryFrom, TryInto};
    pub use std::str::FromStr;
    pub use std::sync::Arc;

    pub use anyhow::Context as ResultContext;
    pub use anyhow::{bail, format_err, Error, Result};

    pub use async_trait::async_trait;
    pub use bigdecimal::BigDecimal as Decimal;
    pub use builder::Builder;
    pub use chrono::{Datelike, Duration, NaiveDate as Date, TimeZone, Utc};
    pub use futures::{Future, FutureExt, Stream, StreamExt};
    pub use lazy_static::lazy_static;
    pub use log::{debug, error, info, trace, warn};
    pub use regex::Regex;
    pub use serde::{Deserialize, Deserializer, Serialize, Serializer};
    pub use tokio::prelude::*;
    pub use uuid::Uuid;

    pub type DateTime<Tz = Utc> = chrono::DateTime<Tz>;
}

pub mod db;
pub mod env;
pub mod graph;
pub mod grocery;
pub mod meta;
pub mod models;
pub mod routes;
pub mod status;
