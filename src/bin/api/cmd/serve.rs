use crate::common::{info as __info, *};

use api::routes::graphql::graphql as graphql_route;
use api::routes::graphql::playground as playground_route;
use api::routes::healthz::healthz as healthz_route;
use api::routes::recover;
use api::routes::shortcuts::bargain_day as bargain_day_route;

use api::db::PgPool;
use api::graph::{Query, Subscription};
use api::graphql::extensions::Logging as LoggingExtension;
use api::grocery::tnt::TntSailor;
use api::meta::BuildInfo;
use api::models::{Contact, Email};

use warp::path::{end as warp_root, path as warp_path};
use warp::Filter as WarpFilter;
use warp::{any as warp_any, serve as warp_serve};

use tokio::runtime::Runtime;
use tokio_compat::FutureExt;

use graphql::extensions::ApolloTracing as TracingExtension;
use graphql::{EmptyMutation, Schema};

use std::net::ToSocketAddrs;
use std::sync::Arc;

use clap::Clap;
use diesel::r2d2::{ConnectionManager, ManageConnection};

macro_rules! info {
    ($($arg:tt)+) => (
        __info!(target: "api::serve", $($arg)+);
    )
}

#[derive(Debug, Clap)]
#[clap(about = "Serve my personal API")]
pub struct ServeCli {
    #[clap(
        long,
        env = "API_TRACE",
        about = "Enable Apollo Tracing",
        takes_value = false
    )]
    #[clap(help_heading = Some("SERVER"))]
    pub trace: bool,

    #[clap(
        long,
        env = "API_HOST",
        about = "Host to serve on",
        value_name = "HOST",
        default_value = "0.0.0.0"
    )]
    #[clap(help_heading = Some("SERVER"))]
    pub host: String,

    #[clap(
        long,
        env = "API_PORT",
        about = "Port to serve on",
        value_name = "PORT",
        default_value = "8080"
    )]
    #[clap(help_heading = Some("SERVER"))]
    pub port: u16,

    #[clap(long, env = "API_MY_FIRST_NAME", value_name = "first name")]
    #[clap(help_heading = Some("SELF"))]
    pub my_first_name: String,

    #[clap(long, env = "API_MY_LAST_NAME", value_name = "last name")]
    #[clap(help_heading = Some("SELF"))]
    pub my_last_name: String,

    #[clap(long, env = "API_MY_EMAIL", value_name = "email")]
    #[clap(help_heading = Some("SELF"))]
    pub my_email: Email,

    #[clap(long, env = "API_MY_ABOUT", value_name = "description")]
    #[clap(help_heading = Some("SELF"))]
    pub my_about: Option<String>,

    #[clap(long, env = "API_MY_BIRTHDAY", value_name = "birthday")]
    #[clap(help_heading = Some("SELF"))]
    pub my_birthday: Date,

    #[clap(
        long,
        env = "API_DATABASE_URL",
        about = "Database URL",
        value_name = "URL",
        hide_env_values = true
    )]
    #[clap(help_heading = Some("DATABASE"))]
    pub database_url: String,

    #[clap(
        long,
        env = "API_DATABASE_MAX_CONNECTIONS",
        about = "Maximum number of concurrent database connections",
        value_name = "N"
    )]
    #[clap(help_heading = Some("DATABASE"))]
    pub database_max_connections: Option<u32>,
}

impl ServeCli {
    pub fn me(&self) -> Contact {
        let Self {
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

pub fn serve(ctx: &Context, cli: ServeCli) -> Result<()> {
    let db = {
        let ServeCli {
            database_url: url,
            database_max_connections: max_connections,
            ..
        } = &cli;
        init_db(url, max_connections.to_owned())
            .context("failed to initialize database")?
    };

    let build = {
        let Context { timestamp, version } = &ctx;
        BuildInfo {
            timestamp: timestamp.to_owned().into(),
            version: version.to_owned(),
        }
    };

    let query = Query;
    let mutation = EmptyMutation;
    let subscription = Subscription;

    let mut schema = Schema::build(query, mutation, subscription)
        .extension(LoggingExtension)
        .data(build)
        .data(db)
        .data(cli.me());
    if cli.trace {
        info!("using Apollo Tracing extension");
        schema = schema.extension(TracingExtension);
    }
    let schema = schema.finish();

    let runtime = Runtime::new().context("failed to initialize runtime")?;
    let runtime = Arc::new(runtime);

    let sailor = TntSailor::new();
    let sailor = Arc::new(sailor);

    let shortcuts_bargain_day =
        warp_path("bargain-day").and(bargain_day_route(sailor));
    let shortcuts = warp_path("shortcuts").and(shortcuts_bargain_day);

    let playground = warp_any().and(playground_route());
    let graphql =
        warp_path("graphql").and(graphql_route(schema, runtime.clone()));
    let healthz = warp_path("healthz").and(healthz_route());
    let filter = warp_root()
        .and(playground)
        .or(healthz)
        .or(graphql)
        .or(shortcuts)
        .recover(recover);

    let ServeCli { host, port, .. } = &cli;
    let address = format!("{}:{}", host, port)
        .to_socket_addrs()
        .context("invalid server address")?
        .as_slice()
        .first()
        .unwrap()
        .to_owned();

    runtime.block_on(async {
        info!("listening on http://{}", &address);
        warp_serve(filter).run(address).compat().await;
        Ok(())
    })
}

fn init_db(url: &str, max_connections: Option<u32>) -> Result<PgPool> {
    let manager = {
        let manager = ConnectionManager::new(url);
        let mut conn = manager.connect()?;
        manager.is_valid(&mut conn).context("invalid connection")?;
        manager
    };
    let mut pool = PgPool::builder();
    if let Some(size) = max_connections {
        pool = pool.max_size(size);
    }
    pool.build(manager)
        .context("failed to create connection pool")
}
