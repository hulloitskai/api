mod config;
use config::{Clap, Config};

use warp::path::{end as warp_root, path as warp_path};
use warp::Filter as WarpFilter;
use warp::{any as warp_any, serve as warp_serve};

use tokio::main as tokio;
use tokio_compat::FutureExt;

use diesel::r2d2::{ConnectionManager, ManageConnection};
use graphql::{EmptyMutation, Schema};
use logger::try_init as init_logger;
use sentry::init as init_sentry;
use std::net::ToSocketAddrs;

use api::env::load as load_env;
use api::graph::{Query, Subscription};
use api::grocery::tnt::TntSailor;
use api::models::BuildInfo;
use api::prelude::*;

use api::routes::graphql::graphql as graphql_route;
use api::routes::graphql::playground as playground_route;
use api::routes::healthz::healthz as healthz_route;
use api::routes::recover;
use api::routes::shortcuts::bargain_day as bargain_day_route;

#[tokio]
async fn main() -> Result<()> {
    load_env().context("load environment variables")?;

    let config = Config::parse();
    let _guard = config
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

    init_logger().context("init logger")?;
    if let Some(version) = &version {
        info!("Starting up (version: {})", version);
    } else {
        info!("Starting up");
    };

    let meta = BuildInfo {
        timestamp: timestamp.into(),
        version,
    };
    let me = config.me();
    let db = connect_db(&config).context("connect database")?;
    let schema = Schema::build(Query, EmptyMutation, Subscription)
        .data(meta)
        .data(db)
        .data(me)
        .finish();

    let sailor = TntSailor::new();
    let sailor = Arc::new(sailor);

    let shortcuts_bargain_day =
        warp_path("bargain-day").and(bargain_day_route(sailor));
    let shortcuts = warp_path("shortcuts").and(shortcuts_bargain_day);
    let playground = warp_any().and(playground_route());
    let graphql = warp_path("graphql").and(graphql_route(&schema));
    let healthz = warp_path("healthz").and(healthz_route());
    let filter = warp_root()
        .and(playground)
        .or(healthz)
        .or(graphql)
        .or(shortcuts)
        .recover(recover);

    let Config { host, port, .. } = &config;
    let address = format!("{}:{}", host, port)
        .to_socket_addrs()
        .context("parse address")?
        .as_slice()
        .first()
        .unwrap()
        .to_owned();
    info!("Listening on http://{}", &address);
    warp_serve(filter).run(address).compat().await;
    Ok(())
}

fn connect_db(
    Config {
        db_url,
        db_max_connections,
        ..
    }: &Config,
) -> Result<DbPool> {
    let manager = {
        let manager = ConnectionManager::new(db_url);
        let mut conn = manager.connect()?;
        manager.is_valid(&mut conn).context("test connection")?;
        manager
    };
    let mut pool = DbPool::builder();
    if let Some(size) = db_max_connections {
        pool = pool.max_size(*size);
    }
    pool.build(manager).context("create connection pool")
}
