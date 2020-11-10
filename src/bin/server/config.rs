use api::models::{Contact, Email};
use api::prelude::*;

pub use clap::{AppSettings, ArgSettings, Clap};

#[derive(Debug, Clap)]
#[clap(name = "api-server", version = env!("BUILD_VERSION"), author)]
#[clap(about = "My personal API server")]
#[clap(setting = AppSettings::ColoredHelp)]
pub struct Config {
    #[clap(long, default_value = "0.0.0.0", env = "API_HOST")]
    pub host: String,
    #[clap(long, default_value = "8080", env = "API_PORT")]
    pub port: u16,

    #[clap(
        long,
        value_name = "url",
        env = "API_DB_URL",
        hide_env_values = true,
        help_heading = Some("DATABASE")
    )]
    pub db_url: String,
    #[clap(
        long,
        value_name = "connections",
        env = "API_DB_MAX_CONNECTIONS",
        help_heading = Some("DATABASE")
    )]
    pub db_max_connections: Option<u32>,

    #[clap(
        long,
        value_name = "dsn",
        env = "API_SENTRY_DSN",
        hide_env_values = true,
        help_heading = Some("SENTRY")
    )]
    pub sentry_dsn: Option<String>,

    #[clap(
        long,
        value_name = "first name",
        env = "API_MY_FIRST_NAME",
        help_heading = Some("SELF")
    )]
    pub my_first_name: String,
    #[clap(
        long,
        value_name = "last name",
        env = "API_MY_LAST_NAME",
        help_heading = Some("SELF")
    )]
    pub my_last_name: String,
    #[clap(
        long,
        value_name = "email",
        env = "API_MY_EMAIL",
        help_heading = Some("SELF")
    )]
    pub my_email: Email,
    #[clap(
        long,
        value_name = "description",
        env = "API_MY_ABOUT",
        help_heading = Some("SELF")
    )]
    pub my_about: Option<String>,
    #[clap(
        long,
        value_name = "birthday",
        env = "API_MY_BIRTHDAY",
        help_heading = Some("SELF")
    )]
    pub my_birthday: Date,
}

impl Config {
    pub fn me(&self) -> Contact {
        let Config {
            my_first_name,
            my_last_name,
            my_about,
            my_email,
            my_birthday,
            ..
        } = self;
        Contact {
            first_name: my_first_name.to_owned(),
            last_name: my_last_name.to_owned(),
            email: my_email.to_owned(),
            about: my_about.to_owned(),
            birthday: my_birthday.to_owned(),
        }
    }
}
