pub use log::{debug, error, info, trace, warn};
pub use std::collections::{HashMap as Map, HashSet as Set};
pub use std::convert::{TryFrom, TryInto};
pub use std::str::FromStr;
pub use std::sync::Arc;

pub use anyhow::{anyhow, bail, Context, Error, Result};
pub use chrono::{Datelike, Duration, NaiveDate as Date, TimeZone, Utc};
pub use futures::{Stream, StreamExt};
pub use regex::Regex;
pub use serde::{Deserialize, Deserializer, Serialize, Serializer};
pub use tokio::prelude::*;
pub use uuid::Uuid;

pub use lazy_static::lazy_static;

use chrono::DateTime as ChronoDateTime;
pub type DateTime<Tz = Utc> = ChronoDateTime<Tz>;

use diesel::r2d2::{ConnectionManager, Pool};
use diesel::PgConnection;
pub type DbPool = Pool<ConnectionManager<PgConnection>>;
