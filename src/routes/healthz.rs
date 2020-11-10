use crate::status::{Health, Status};

use warp::reply::{json, Reply};
use warp::{get, head, Filter, Rejection};

pub fn healthz(
) -> impl Filter<Extract = (impl Reply,), Error = Rejection> + Clone {
    let method = head().or(get()).unify();
    method.map(|| {
        let health = Health::new(Status::Pass);
        json(&health)
    })
}
