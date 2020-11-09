pub mod graphql;
pub mod healthz;
pub mod shortcuts;

use crate::prelude::*;
use warp::reject::Reject;

#[derive(Debug)]
pub struct RouteError(Error);

impl Reject for RouteError {}

impl From<Error> for RouteError {
    fn from(error: Error) -> Self {
        Self(error)
    }
}
