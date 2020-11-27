use super::common::*;

use warp::header::optional as header;
use warp::path::{full as full_path, FullPath};
use warp::reply::{html, Reply};
use warp::{any, Filter, Rejection};

use graphql::http::{
    playground_source, GraphQLPlaygroundConfig as PlaygroundConfig,
};
use graphql::{ObjectType, SubscriptionType};
use graphql::{Request as GraphQLRequest, Schema};

use graphql_warp::{
    graphql as graphql_filter,
    graphql_subscription as graphql_subscription_filter,
    Response as GraphQLResponse,
};

use std::convert::Infallible;
use tokio::runtime::Runtime;

pub fn graphql<Q, M, S>(
    schema: Schema<Q, M, S>,
    runtime: Arc<Runtime>,
) -> impl Filter<Extract = (impl Reply,), Error = Rejection> + Clone
where
    Q: ObjectType + Send + Sync + 'static,
    M: ObjectType + Send + Sync + 'static,
    S: SubscriptionType + Send + Sync + 'static,
{
    let graphql = graphql_filter(schema.clone())
        .map(move |(schema, request)| (schema, request, runtime.clone()))
        .and_then(
            |(schema, request, runtime): (
                Schema<Q, M, S>,
                GraphQLRequest,
                Arc<Runtime>,
            )| async move {
                let response = runtime
                    .spawn(async move { schema.execute(request).await })
                    .await
                    .unwrap();
                Ok::<_, Infallible>(GraphQLResponse::from(response))
            },
        );
    let subscription = graphql_subscription_filter(schema);
    subscription.or(graphql)
}

pub fn playground(
) -> impl Filter<Extract = (impl Reply,), Error = Rejection> + Clone {
    any()
        .and(full_path())
        .and(header::<String>("X-Forwarded-Prefix"))
        .map(|path: FullPath, prefix: Option<String>| {
            let prefix = prefix.unwrap_or_else(String::new);
            let endpoint = format!("{}{}graphql", &prefix, path.as_str());
            let source = playground_source(
                PlaygroundConfig::new(&endpoint)
                    .subscription_endpoint(&endpoint),
            );
            html(source)
        })
}
