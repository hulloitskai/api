use api::env::load as load_env;
use api::prelude::*;

use clap::{AppSettings, Clap};
use sentry::init as init_sentry;
use tokio::main as tokio;

mod context;
use context::*;
mod serve;
use serve::*;

#[derive(Debug, Clap)]
#[clap(name = "api", version = env!("BUILD_VERSION"), author)]
#[clap(about = "My personal API server")]
#[clap(setting = AppSettings::ColoredHelp)]
pub struct Cli {
    #[clap(
        long,
        value_name = "dsn",
        env = "API_SENTRY_DSN",
        hide_env_values = true
    )]
    pub sentry_dsn: Option<String>,

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

    let cli = Cli::parse();
    let _guard = cli
        .sentry_dsn
        .as_ref()
        .map(|dsn| init_sentry(dsn.as_str()))
        .or_else(|| {
            warn!("Missing Sentry DSN; Sentry is disabled");
            None
        });

    let timestamp = DateTime::parse_from_rfc3339(env!("BUILD_TIMESTAMP"))
        .context("parse build timestamp")?;
    let version = match env!("BUILD_VERSION") {
        "" => None,
        version => Some(version.to_owned()),
    };

    use Command::*;
    let ctx = Context { timestamp, version };
    let cmd = match cli.cmd {
        Serve(cli) => serve(&ctx, cli),
    };
    cmd.await
}
