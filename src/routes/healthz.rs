use crate::status::{Health, Status};

use std::convert::Infallible;
use warp::reply::{json, Reply};
use warp::{any, Filter};

pub fn healthz(
) -> impl Filter<Extract = (impl Reply,), Error = Infallible> + Clone {
    any().map(|| {
        let health = Health::new(Status::Pass);
        json(&health)
    })
}
