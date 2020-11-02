use api::prelude::*;

use std::convert::Infallible;
use std::env::VarError as EnvVarError;
use std::net::SocketAddr;

use graphql::http::{
    playground_source as playground,
    GraphQLPlaygroundConfig as PlaygroundConfig,
};
use graphql::{EmptyMutation, EmptySubscription, Request, Schema};
use graphql_warp::{graphql as graphql_filter, Response};

use warp::path::end as warp_root;
use warp::reply::{html as warp_html, json as warp_json};
use warp::Filter as WarpFilter;
use warp::{path as warp_path, serve as warp_serve};

use diesel::r2d2::{ConnectionManager, ManageConnection};
use logger::try_init as init_logger;
use sentry::init as init_sentry;

use api::env::{load as load_env, var as env_var};
use api::graph::{Context as ResolverContext, DbPool, Query};
use api::models::{Contact, Email};
use api::status::{Health, Status};

type ApiSchema = Schema<Query, EmptyMutation, EmptySubscription>;

#[tokio::main]
async fn main() -> Result<()> {
    load_env().context("load environment variables")?;
    init_logger().context("init logger")?;

    let timestamp = DateTime::parse_from_rfc3339(env!(r"BUILD_TIMESTAMP"))
        .context("parse build timestamp")?;
    let version = match env!(r"BUILD_VERSION") {
        "" => None,
        version => Some(version),
    };
    if let Some(version) = version {
        info!("Starting up (version: {})", version);
    } else {
        info!("Starting up");
    }

    let _guard = match env_var("SENTRY_DSN") {
        Ok(dsn) => {
            let guard = init_sentry(dsn);
            Some(guard)
        }
        Err(_) => {
            info!("Missing Sentry DSN; Sentry is disabled");
            None
        }
    };

    let db = connect_db().context("connect database")?;
    let context = ResolverContext::new(db);
    let contact = contact_from_env().context("get contact from env")?;

    let query =
        Query::new(timestamp.into(), version.map(ToOwned::to_owned), &contact);
    let schema = Schema::build(query, EmptyMutation, EmptySubscription)
        .data(context)
        .finish();

    let graphql = warp_path("graphql").and(graphql_filter(schema)).and_then(
        |(schema, request): (ApiSchema, Request)| async move {
            let response = schema.execute(request).await;
            Ok::<_, Infallible>(Response::from(response))
        },
    );
    let healthz = warp_path("healthz").map(|| {
        let health = Health::new(Status::Pass);
        warp_json(&health)
    });
    let index = warp_root().map(|| {
        let html = playground(PlaygroundConfig::new("graphql"));
        warp_html(html)
    });
    let filter = index.or(healthz).or(graphql);

    let server_port = env_var("PORT").context("get port")?;
    let server_addr: SocketAddr =
        format!("0.0.0.0:{}", &server_port).parse()?;

    info!("Listening on http://{}", server_addr);
    warp_serve(filter).run(server_addr).await;
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
    DbPool::new(manager).context("create connection pool")
}

fn contact_from_env() -> Result<Contact> {
    let first_name = env_var("CONTACT_FIRST_NAME").context("first name")?;
    let last_name = env_var("CONTACT_LAST_NAME").context("last name")?;

    let about = match env_var("CONTACT_ABOUT") {
        Ok(about) => Ok(Some(about)),
        Err(error) => match error {
            EnvVarError::NotPresent => Ok(None),
            error => Err(error),
        },
    }
    .context("about")?;

    let email = env_var("CONTACT_EMAIL").context("email")?;
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
