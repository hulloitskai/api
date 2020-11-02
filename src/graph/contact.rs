use super::prelude::*;
use crate::models::Contact as ContactModel;
use crate::prelude::*;

use chrono_humanize::{
    Accuracy as HumanAccuracy, HumanTime, Tense as HumanTense,
};

#[derive(ConstantObject)]
struct ContactConstants {
    name: String,
    about: Option<String>,
    email: String,
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

#[derive(CombinedObject)]
pub struct Contact(ContactConstants, ContactResolvers);

impl Contact {
    pub fn new(
        ContactModel {
            name,
            about,
            email,
            birthday,
        }: ContactModel,
    ) -> Self {
        let constants = ContactConstants {
            name,
            about,
            email: email.to_string(),
        };
        let resolvers = ContactResolvers { birthday };
        Contact(constants, resolvers)
    }
}
