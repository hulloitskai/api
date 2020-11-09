use api::prelude::*;

use std::env::VarError as EnvVarError;
use std::net::ToSocketAddrs;

use graphql::{EmptyMutation, Schema};

use warp::path::{end as warp_root, path as warp_path};
use warp::Filter as WarpFilter;
use warp::{any as warp_any, serve as warp_serve};

use tokio::main as tokio;
use tokio_compat::FutureExt;

use diesel::r2d2::{ConnectionManager, ManageConnection};
use logger::try_init as init_logger;
use sentry::init as init_sentry;

use api::env::{load as load_env, var as env_var, var_or as env_var_or};
use api::graph::{Query, Subscription};
use api::grocery::tnt::TntSailor;
use api::models::{BuildInfo, Contact, Email};

use api::routes::graphql::graphql as graphql_route;
use api::routes::graphql::playground as playground_route;
use api::routes::healthz::healthz as healthz_route;
use api::routes::shortcuts::bargain_day as bargain_day_route;

#[tokio]
async fn main() -> Result<()> {
    load_env().context("load environment variables")?;
    init_logger().context("init logger")?;

    let timestamp = DateTime::parse_from_rfc3339(env!(r"BUILD_TIMESTAMP"))
        .context("parse build timestamp")?;
    let version = match env!(r"BUILD_VERSION") {
        "" => None,
        version => Some(version.to_owned()),
    };
    if let Some(version) = &version {
        info!("Starting up (version: {})", version);
    } else {
        info!("Starting up");
    };

    let _guard = match env_var("SENTRY_DSN") {
        Ok(dsn) => {
            let guard = init_sentry(dsn);
            Some(guard)
        }
        Err(_) => {
            warn!("Missing Sentry DSN; Sentry is disabled");
            None
        }
    };

    let meta = BuildInfo {
        timestamp: timestamp.into(),
        version,
    };
    let me = contact_from_env().context("get contact from env")?;
    let db = connect_db().context("connect database")?;
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

    let server_host = env_var_or("HOST", "0.0.0.0").context("get host")?;
    let server_port = env_var_or("PORT", "8080").context("get port")?;
    let server_addr = format!("{}:{}", &server_host, &server_port)
        .to_socket_addrs()
        .context("parse address")?
        .as_slice()
        .first()
        .unwrap()
        .to_owned();

    info!("Listening on http://{}", &server_addr);
    warp_serve(
        warp_root()
            .and(playground)
            .or(healthz)
            .or(graphql)
            .or(shortcuts),
    )
    .run(server_addr)
    .compat()
    .await;
    Ok(())
}

fn connect_db() -> Result<DbPool> {
    let url = env_var("POSTGRES_URL").context("get url")?;
    let manager = {
        let manager = ConnectionManager::new(&url);
        let mut conn = manager.connect()?;
        manager.is_valid(&mut conn).context("test connection")?;
        manager
    };

    let mut pool = DbPool::builder();
    if let Ok(size) = env_var("POSTGRES_MAX_CONNECTIONS") {
        let size: u32 = size.parse().context("parse max connections")?;
        info!("Limiting to {} Postgres connections", size);
        pool = pool.max_size(size);
    }
    pool.build(manager).context("create connection pool")
}

fn contact_from_env() -> Result<Contact> {
    let first_name = env_var("CONTACT_FIRST_NAME").context("get first name")?;
    let last_name = env_var("CONTACT_LAST_NAME").context("get last name")?;

    let about = match env_var("CONTACT_ABOUT") {
        Ok(about) => Ok(Some(about)),
        Err(error) => match error {
            EnvVarError::NotPresent => Ok(None),
            error => Err(error),
        },
    }
    .context("get about text")?;

    let email = env_var("CONTACT_EMAIL").context("get email")?;
    let email = Email::new(email).context("parse email")?;

    let birthday = env_var("CONTACT_BIRTHDAY").context("birthday")?;
    let birthday = Date::from_str(&birthday).context("parse birthday")?;

    Ok(Contact {
        first_name,
        last_name,
        about,
        email,
        birthday,
    })
}
