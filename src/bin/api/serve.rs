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

use tokio::runtime::Runtime;
use tokio_compat::FutureExt;

use clap::Clap;
use diesel::r2d2::{ConnectionManager, ManageConnection};
use graphql::{EmptyMutation, Schema};
use std::net::ToSocketAddrs;

#[derive(Debug, Clap)]
#[clap(about = "Serve my personal API")]
pub struct ServeCli {
    #[clap(
        long,
        env = "API_HOST",
        about = "Host to serve on",
        value_name = "HOST",
        default_value = "0.0.0.0"
    )]
    pub host: String,

    #[clap(
        long,
        env = "API_PORT",
        about = "Port to serve on",
        value_name = "PORT",
        default_value = "8080"
    )]
    pub port: u16,

    #[clap(
        long,
        env = "API_DB_URL",
        about = "Database URL",
        value_name = "URL",
        hide_env_values = true
    )]
    #[clap(help_heading = Some("DATABASE"))]
    pub db_url: String,

    #[clap(
        long,
        env = "API_DB_MAX_CONNECTIONS",
        about = "Maximum concurrent database connections",
        value_name = "N"
    )]
    #[clap(help_heading = Some("DATABASE"))]
    pub db_max_connections: Option<u32>,

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

    let runtime = Runtime::new().context("initialize runtime")?;
    runtime.block_on(async {
        info!("listening on http://{}", &address);
        warp_serve(filter).run(address).compat().await;
        Ok(())
    })
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
