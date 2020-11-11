use super::context::Context;

use api::routes::graphql::graphql as graphql_route;
use api::routes::graphql::playground as playground_route;
use api::routes::healthz::healthz as healthz_route;
use api::routes::recover;
use api::routes::shortcuts::bargain_day as bargain_day_route;

use api::build::Build;
use api::graph::{Query, Subscription};
use api::grocery::tnt::TntSailor;
use api::models::{Contact, Email};
use api::prelude::*;

use warp::path::{end as warp_root, path as warp_path};
use warp::Filter as WarpFilter;
use warp::{any as warp_any, serve as warp_serve};

use clap::{AppSettings, Clap};
use diesel::r2d2::{ConnectionManager, ManageConnection};
use graphql::{EmptyMutation, Schema};
use std::net::ToSocketAddrs;
use tokio_compat::FutureExt;

#[derive(Debug, Clap)]
#[clap(about = "Serve my personal API")]
#[clap(setting = AppSettings::ColoredHelp)]
#[clap(setting = AppSettings::DeriveDisplayOrder)]
pub struct ServeCli {
    #[clap(
        long,
        about = "Host to serve on",
        default_value = "0.0.0.0",
        env = "API_HOST"
    )]
    pub host: String,

    #[clap(
        long,
        about = "Port to serve on",
        default_value = "8080",
        env = "API_PORT"
    )]
    pub port: u16,

    #[clap(
        long,
        about = "Database URL",
        value_name = "url",
        env = "API_DB_URL",
        hide_env_values = true
    )]
    #[clap(help_heading = Some("DATABASE"))]
    pub db_url: String,

    #[clap(
        long,
        about = "Maximum concurrent database connections",
        value_name = "connections",
        env = "API_DB_MAX_CONNECTIONS"
    )]
    #[clap(help_heading = Some("DATABASE"))]
    pub db_max_connections: Option<u32>,

    #[clap(long, value_name = "first name", env = "API_MY_FIRST_NAME")]
    #[clap(help_heading = Some("SELF"))]
    pub my_first_name: String,

    #[clap(long, value_name = "last name", env = "API_MY_LAST_NAME")]
    #[clap(help_heading = Some("SELF"))]
    pub my_last_name: String,

    #[clap(long, value_name = "email", env = "API_MY_EMAIL")]
    #[clap(help_heading = Some("SELF"))]
    pub my_email: Email,

    #[clap(long, value_name = "description", env = "API_MY_ABOUT")]
    #[clap(help_heading = Some("SELF"))]
    pub my_about: Option<String>,

    #[clap(long, value_name = "birthday", env = "API_MY_BIRTHDAY")]
    #[clap(help_heading = Some("SELF"))]
    pub my_birthday: Date,
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

pub async fn serve(ctx: &Context, cli: ServeCli) -> Result<()> {
    let Context { timestamp, version } = &ctx;
    let build = Build {
        timestamp: timestamp.to_owned().into(),
        version: version.to_owned(),
    };

    let ServeCli {
        db_url,
        db_max_connections,
        ..
    } = &cli;
    let db = connect_db(db_url, db_max_connections.to_owned())
        .context("connect database")?;

    let schema = Schema::build(Query, EmptyMutation, Subscription)
        .data(build)
        .data(db)
        .data(cli.me())
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

    let ServeCli { host, port, .. } = &cli;
    let address = format!("{}:{}", host, port)
        .to_socket_addrs()
        .context("parse address")?
        .as_slice()
        .first()
        .unwrap()
        .to_owned();

    info!("listening on http://{}", &address);
    warp_serve(filter).run(address).compat().await;
    Ok(())
}

fn connect_db(url: &str, max_connections: Option<u32>) -> Result<PgPool> {
    let manager = {
        let manager = ConnectionManager::new(url);
        let mut conn = manager.connect()?;
        manager.is_valid(&mut conn).context("test connection")?;
        manager
    };
    let mut pool = PgPool::builder();
    if let Some(size) = max_connections {
        pool = pool.max_size(size);
    }
    pool.build(manager).context("create connection pool")
}
