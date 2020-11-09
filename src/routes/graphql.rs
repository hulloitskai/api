use std::convert::Infallible;

use warp::header::optional as header;
use warp::path::{full as full_path, FullPath};
use warp::reply::{html, Reply};
use warp::{any, Filter, Rejection};

use graphql::http::{
    playground_source, GraphQLPlaygroundConfig as PlaygroundConfig,
};
use graphql::{
    ObjectType, Request as GraphQLRequest, Schema, SubscriptionType,
};
use graphql_warp::{
    graphql as graphql_filter,
    graphql_subscription as graphql_subscription_filter,
    Response as GraphQLResponse,
};

pub fn graphql<Query, Mutation, Subscription>(
    schema: &Schema<Query, Mutation, Subscription>,
) -> impl Filter<Extract = (impl Reply,), Error = Rejection> + Clone
where
    Query: ObjectType + Send + Sync + 'static,
    Mutation: ObjectType + Send + Sync + 'static,
    Subscription: SubscriptionType + Send + Sync + 'static,
{
    let subscription = graphql_subscription_filter(schema.clone());
    subscription.or(graphql_filter(schema.clone()).and_then(
        |(schema, request): (
            Schema<Query, Mutation, Subscription>,
            GraphQLRequest,
        )| async move {
            let response = schema.execute(request).await;
            Ok::<_, Infallible>(GraphQLResponse::from(response))
        },
    ))
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
