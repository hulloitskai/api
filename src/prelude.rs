pub use std::collections::{HashMap as Map, HashSet as Set};
pub use std::convert::{TryFrom, TryInto};
pub use std::str::FromStr;
pub use std::sync::Arc;

pub use anyhow::{anyhow, bail, Context, Error, Result};
pub use async_trait::async_trait;
pub use bigdecimal::BigDecimal as Decimal;
pub use chrono::{Datelike, Duration, NaiveDate as Date, TimeZone, Utc};
pub use derive_builder::Builder;
pub use futures::{Future, FutureExt, Stream, StreamExt};
pub use lazy_static::lazy_static;
pub use log::{debug, error, info, trace, warn};
pub use regex::Regex;
pub use serde::{Deserialize, Deserializer, Serialize, Serializer};
pub use tokio::prelude::*;
pub use uuid::Uuid;

use chrono::DateTime as ChronoDateTime;
pub type DateTime<Tz = Utc> = ChronoDateTime<Tz>;

use diesel::r2d2::{ConnectionManager, Pool};
use diesel::PgConnection;
pub type DbPool = Pool<ConnectionManager<PgConnection>>;
