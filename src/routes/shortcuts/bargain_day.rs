use crate::grocery::{Product, Sailor};
use crate::routes::common::*;

use warp::filters::query::query;
use warp::reject::{custom, Rejection};
use warp::reply::Reply;
use warp::{get, head, Filter};

use std::sync::Arc;

#[derive(Debug, Clone, Serialize, Deserialize)]
struct BargainDayParams {
    postcode: String,
}

pub fn bargain_day<S>(
    sailor: Arc<S>,
) -> impl Filter<Extract = (impl Reply,), Error = Rejection> + Clone
where
    S: Sailor + Send + Sync,
{
    let method = head().or(get()).unify();
    method
        .map(move || sailor.clone())
        .and(query::<BargainDayParams>())
        .and_then(|sailor: Arc<S>, params: BargainDayParams| async move {
            let products: Vec<Product> = sailor
                .get_sale_products(params.postcode)
                .await
                .context("failed to get on-sale products")
                .map_err(|error| custom(RouteError::from(error)))?
                .collect();
            let message = match products[..].split_last() {
                Some((last, init)) => format!(
                    "{}, and {}",
                    init.iter()
                        .map(describe_product)
                        .collect::<Vec<String>>()
                        .join(", "),
                    describe_product(last)
                ),
                None => match products.first() {
                    Some(product) => describe_product(product),
                    None => {
                        let error = format_err!("no products found");
                        return Err(custom(RouteError::from(error)));
                    }
                },
            };
            let message = format!("Today's sale items include: {}", &message);
            Ok(message)
        })
}

fn describe_product(product: &Product) -> String {
    format!("{} (${:.2})", product.name, product.price())
}
