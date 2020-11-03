use crate::prelude::*;

use super::Email;
use chrono::{Date as TzDate, LocalResult};

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
pub struct Contact {
    pub first_name: String,
    pub last_name: String,
    pub email: Email,
    pub about: Option<String>,
    pub birthday: Date,
}

impl Contact {
    pub fn name(&self) -> String {
        format!("{} {}", self.first_name, self.last_name)
    }

    pub fn birthday_in_time_zone<Tz>(&self, time_zone: Tz) -> Result<TzDate<Tz>>
    where
        Tz: TimeZone,
    {
        let birthday = time_zone.from_local_date(&self.birthday);
        if let LocalResult::Single(date) = birthday {
            Ok(date)
        } else {
            Err(anyhow!("invalid or ambiguous conversion"))
        }
    }
}
