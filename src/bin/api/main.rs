use api::env::load as load_env;
use api::prelude::*;

use logger::init as init_logger;
use sentry::init as init_sentry;

use clap::{AppSettings, Clap};
use std::env::set_var as set_env_var;
use tokio::main as tokio;

mod context;
use context::*;
mod serve;
use serve::*;

#[derive(Debug, Clap)]
#[clap(name = "api", version = env!("BUILD_VERSION"))]
#[clap(about = "My personal API server")]
#[clap(setting = AppSettings::ColoredHelp)]
#[clap(setting = AppSettings::DeriveDisplayOrder)]
pub struct Cli {
    #[clap(
        long,
        about = "Sentry DSN for error reporting",
        value_name = "dsn",
        env = "API_SENTRY_DSN",
        hide_env_values = true
    )]
    pub sentry_dsn: Option<String>,

    #[clap(
        long,
        about = "Log level and directives",
        value_name = "level",
        env = "API_LOG",
        default_value = "warn,api=info",
        hide_default_value = true
    )]
    pub log: String,

    #[clap(subcommand)]
    pub cmd: Command,
}

#[derive(Debug, Clap)]
pub enum Command {
    Serve(ServeCli),
}

#[tokio]
async fn main() -> Result<()> {
    load_env().context("load environment variables")?;

    // Parse command line and initialize Sentry.
    let cli = Cli::parse();
    let _guard = cli
        .sentry_dsn
        .as_ref()
        .map(|dsn| init_sentry(dsn.as_str()))
        .or_else(|| {
            warn!("Missing Sentry DSN; Sentry is disabled");
            None
        });

    // Read build info.
    let timestamp = DateTime::parse_from_rfc3339(env!("BUILD_TIMESTAMP"))
        .context("parse build timestamp")?;
    let version = match env!("BUILD_VERSION") {
        "" => None,
        version => Some(version.to_owned()),
    };
    let context = Context { timestamp, version };

    // Configure logger.
    set_env_var("RUST_LOG", &cli.log);
    init_logger();
    if let Some(version) = &context.version {
        debug!("starting up (version: {})", version);
    } else {
        debug!("starting up");
    };

    // Run subcommand.
    use Command::*;
    let cmd = match cli.cmd {
        Serve(cli) => serve(&context, cli),
    };
    cmd.await
}
