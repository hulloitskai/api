use crate::prelude::*;

use warp::reject::{custom, Rejection};
use warp::reply::Reply;
use warp::{any, Filter};

use std::sync::Arc;

use crate::grocery::{Product, Sailor};
use crate::routes::RouteError;

pub fn bargain_day<S>(
    sailor: Arc<S>,
) -> impl Filter<Extract = (impl Reply,), Error = Rejection> + Clone
where
    S: Sailor + Send + Sync,
{
    any()
        .map(move || sailor.clone())
        .and_then(|sailor: Arc<S>| async move {
            let products: Vec<Product> = sailor
                .get_sale_products()
                .await
                .context("get on-sale products")
                .map_err(|error| custom(RouteError::from(error)))?;
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
                        let error = anyhow!("no products found");
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
