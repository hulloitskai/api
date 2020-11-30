use super::common::*;
use crate::models::Contact as ContactModel;

use chrono_humanize::{
    Accuracy as HumanAccuracy, HumanTime, Tense as HumanTense,
};
use chrono_tz::Tz;

pub struct Contact {
    model: ContactModel,
}

#[Object]
impl Contact {
    async fn first_name(&self) -> &String {
        &self.model.first_name
    }

    async fn last_name(&self) -> &String {
        &self.model.last_name
    }

    async fn name(&self) -> String {
        self.model.name()
    }

    async fn email(&self) -> String {
        self.model.email.to_string()
    }

    async fn about(&self) -> &Option<String> {
        &self.model.about
    }

    async fn age(
        &self,
        #[graphql(desc = "The time zone of the viewer.", default = "UTC")]
        time_zone: String,
    ) -> FieldResult<String> {
        let time_zone = Tz::from_str(&time_zone)
            .map_err(|message| format_err!(message))
            .context("failed to parse time zone")
            .into_field_result()?;
        let birthday = self
            .model
            .birthday_in_time_zone(time_zone)
            .context("failed to represent birthday in the given timezone")
            .into_field_result()?
            .and_hms(0, 0, 0);
        let now: DateTime<Tz> =
            time_zone.from_utc_datetime(&Utc::now().naive_utc());
        let age = now - birthday;
        let age = HumanTime::from(age)
            .to_text_en(HumanAccuracy::Rough, HumanTense::Present);
        Ok(age)
    }
}

impl Contact {
    pub fn new(model: ContactModel) -> Self {
        Contact { model }
    }
}
