mod common {
    pub use crate::ctx::*;

    pub use anyhow::{format_err, Context as AnyhowContext, Result};
    pub use chrono::{Datelike, Duration, NaiveDate as Date, TimeZone, Utc};
    pub use clap::Clap;
    pub use log::{debug, error, info, warn};

    pub type DateTime<Tz = Utc> = chrono::DateTime<Tz>;
}

mod cmd;
mod ctx;

use common::*;
use logger::init as init_logger;
use sentry::init as init_sentry;

use cmd::*;
use std::env::set_var as set_env_var;

use api::env::load as load_env;
use clap::{AppSettings, Clap};

#[derive(Debug, Clap)]
#[clap(name = "api", version = env!("BUILD_VERSION"))]
#[clap(about = "My personal API server")]
#[clap(global_setting = AppSettings::ColoredHelp)]
#[clap(global_setting = AppSettings::DeriveDisplayOrder)]
pub struct Cli {
    #[clap(
        long,
        env = "API_SENTRY_DSN",
        about = "Sentry DSN for error reporting",
        value_name = "DSN",
        hide_env_values = true,
        global = true
    )]
    pub sentry_dsn: Option<String>,

    #[clap(
        long,
        env = "API_LOG",
        about = "Log level and directives",
        value_name = "LEVEL",
        default_value = "warn,api=info",
        hide_default_value = true,
        global = true
    )]
    pub log: String,

    #[clap(subcommand)]
    pub cmd: Command,
}

fn main() -> Result<()> {
    load_env().context("failed to load environment variables")?;

    // Parse command line and initialize Sentry.
    let cli = Cli::parse();
    let _guard = cli
        .sentry_dsn
        .as_ref()
        .map(|dsn| init_sentry(dsn.as_str()))
        .or_else(|| {
            warn!("missing Sentry DSN; Sentry is disabled");
            None
        });

    // Read build info.
    let timestamp = DateTime::parse_from_rfc3339(env!("BUILD_TIMESTAMP"))
        .context("parse build timestamp")?;
    let version = match env!("BUILD_VERSION") {
        "" => None,
        version => Some(version.to_owned()),
    };
    let ctx = Context { timestamp, version };

    // Configure logger.
    set_env_var("RUST_LOG", &cli.log);
    init_logger();
    if let Some(version) = &ctx.version {
        debug!("starting up (version: {})", version);
    } else {
        debug!("starting up");
    };

    // Run subcommand.
    use Command::*;
    match cli.cmd {
        Serve(cli) => serve(&ctx, cli),
    }
}
