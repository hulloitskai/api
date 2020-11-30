use api::env::load as load_env;
use api::env::var as env_var;

use api::grocery::tnt::TntSailor;
use api::grocery::Product;
use api::grocery::Sailor;

use anyhow::{format_err, Context as ResultContext, Error, Result};
use logger::init as init_logger;

use std::env::{args, VarError as EnvVarError};
use tokio::runtime::Runtime;

fn main() -> Result<()> {
    load_env().context("failed to load environment variables")?;
    init_logger();

    let page_size = match env_var("TNT_PAGE_SIZE") {
        Ok(s) => Ok(Some(s)),
        Err(EnvVarError::NotPresent) => Ok(None),
        Err(error) => Err(error),
    }
    .context("failed to get page size")?;
    let page_size: Result<Option<u32>> =
        page_size.map_or(Ok(None as Option<u32>), |s| {
            let size: u32 = s.parse()?;
            Ok(Some(size))
        });
    let page_size = page_size.context("failed to parse page size")?;

    let mut sailor = TntSailor::builder();
    if let Some(page_size) = page_size {
        sailor.page_size(page_size);
    }
    let sailor = sailor
        .build()
        .map_err(Error::msg)
        .context("failed to initialize sailor")?;
    let postcode = args()
        .nth(1)
        .ok_or_else(|| format_err!("missing postcode"))?;

    let runtime = Runtime::new().context("failed to initialize runtime")?;
    let products = runtime.block_on(async {
        sailor
            .get_sale_products(postcode)
            .await
            .context("failed to get sale items")
    })?;

    products.for_each(|product| {
        let Product { name, .. } = &product;
        let price = product.price();
        match product.discount() {
            Some(discount) => {
                println!("{} (${:.2}, you save ${:.2})", name, price, discount)
            }
            None => println!("{} (${:.2})", name, price,),
        };
    });
    Ok(())
}
