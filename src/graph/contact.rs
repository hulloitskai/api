use super::prelude::*;
use crate::models::Contact as ContactModel;
use crate::prelude::*;

use chrono_humanize::{
    Accuracy as HumanAccuracy, HumanTime, Tense as HumanTense,
};

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
    birthday: Date,
}

#[ResolverObject]
impl ContactResolvers {
    async fn age(&self) -> String {
        let now = Utc::now().naive_utc().date();
        let age = now - self.birthday;
        HumanTime::from(age)
            .to_text_en(HumanAccuracy::Rough, HumanTense::Present)
    }
}

impl ContactResolvers {
    fn new(ContactModel { birthday, .. }: &ContactModel) -> Self {
        ContactResolvers {
            birthday: birthday.to_owned(),
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
