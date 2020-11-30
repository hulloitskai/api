// Workaround for: https://github.com/rust-lang/rust/issues/64450
extern crate async_trait;
extern crate builder;

mod common {
    pub use std::collections::{HashMap as Map, HashSet as Set};
    pub use std::convert::{TryFrom, TryInto};
    pub use std::fmt::{Display, Formatter};
    pub use std::fmt::{Error as FmtError, Result as FmtResult};
    pub use std::str::FromStr;
    pub use std::sync::{Arc, Mutex};

    pub use anyhow::Context as ResultContext;
    pub use anyhow::{bail, format_err, Error, Result};

    pub use futures::{Future, Stream, TryFuture};
    pub use futures_util::{FutureExt, StreamExt};

    pub use async_trait::async_trait;
    pub use bigdecimal::BigDecimal as Decimal;
    pub use builder::Builder;
    pub use chrono::{Datelike, Duration, NaiveDate as Date, TimeZone, Utc};
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
pub mod graphql;
pub mod grocery;
pub mod meta;
pub mod models;
pub mod routes;
pub mod status;
