use super::prelude::*;
use crate::prelude::*;

use crate::models::Contact as ContactModel;
use chrono_humanize::{
    Accuracy as HumanAccuracy, HumanTime, Tense as HumanTense,
};
use chrono_tz::Tz;

#[derive(ConstantObject)]
struct ContactConstants {
    first_name: String,
    last_name: String,
    name: String,
    email: String,
    about: Option<String>,
}

impl ContactConstants {
    fn new(model: &ContactModel) -> Self {
        let ContactModel {
            first_name,
            last_name,
            email,
            about,
            ..
        } = model;
        ContactConstants {
            first_name: first_name.to_owned(),
            last_name: last_name.to_owned(),
            name: model.name(),
            email: email.to_string(),
            about: about.to_owned(),
        }
    }
}

pub struct ContactResolvers {
    model: ContactModel,
}

#[ResolverObject]
impl ContactResolvers {
    async fn age(
        &self,
        #[graphql(desc = "The time zone of the viewer.", default = "UTC")]
        time_zone: String,
    ) -> FieldResult<String> {
        let time_zone = Tz::from_str(&time_zone)
            .map_err(|message| anyhow!(message))
            .context("parse time zone")
            .map_err(|error| format!("{:#}", error))?;
        let birthday = self
            .model
            .birthday_in_time_zone(time_zone)
            .map(|date| date.and_hms(0, 0, 0))
            .context("get birthday in timezone")
            .map_err(|error| format!("{:#}", error))?;
        let now: DateTime<Tz> =
            time_zone.from_utc_datetime(&Utc::now().naive_utc());
        let age = now - birthday;
        let age = HumanTime::from(age)
            .to_text_en(HumanAccuracy::Rough, HumanTense::Present);
        Ok(age)
    }
}

impl ContactResolvers {
    fn new(model: &ContactModel) -> Self {
        ContactResolvers {
            model: model.to_owned(),
        }
    }
}

#[derive(CombinedObject)]
pub struct Contact(ContactConstants, ContactResolvers);

impl Contact {
    pub fn new(model: &ContactModel) -> Self {
        let constants = ContactConstants::new(model);
        let resolvers = ContactResolvers::new(model);
        Contact(constants, resolvers)
    }
}
