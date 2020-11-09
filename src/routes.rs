pub mod graphql;
pub mod healthz;
pub mod shortcuts;

use crate::prelude::*;

use http::StatusCode;
use warp::reject::{Reject, Rejection};
use warp::reply::{with_status, Reply, Response};

#[derive(Debug, Clone)]
pub struct RouteError {
    message: String,
    status: StatusCode,
}

impl Reject for RouteError {}

impl From<Error> for RouteError {
    fn from(error: Error) -> Self {
        let message = format!("Error: {:#}\n", error);
        RouteError {
            message,
            status: StatusCode::INTERNAL_SERVER_ERROR,
        }
    }
}

impl Reply for RouteError {
    fn into_response(self) -> Response {
        with_status(self.message, self.status).into_response()
    }
}

pub async fn recover(rejection: Rejection) -> Result<impl Reply, Rejection> {
    if let Some(error) = rejection.find::<RouteError>() {
        Ok(error.to_owned())
    } else {
        Err(rejection)
    }
}
