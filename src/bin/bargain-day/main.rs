use api::prelude::*;

use logger::try_init as init_logger;
use std::env::VarError as EnvVarError;
use tokio::main as tokio;

use api::env::{load as load_env, var as env_var};
use api::grocery::tnt::TntSailor;
use api::grocery::Sailor;

#[tokio]
async fn main() -> Result<()> {
    load_env().context("load environment variables")?;
    init_logger().context("init logger")?;

    let page_size = match env_var("TNT_PAGE_SIZE") {
        Ok(s) => Ok(Some(s)),
        Err(error) => match error {
            EnvVarError::NotPresent => Ok(None),
            error => Err(error),
        },
    }
    .context("get page size")?;
    let page_size: Result<Option<u32>> =
        page_size.map_or(Ok(None as Option<u32>), |s| {
            let size: u32 = s.parse()?;
            Ok(Some(size))
        });
    let page_size = page_size.context("parse page size")?;

    let mut sailor = TntSailor::builder();
    if let Some(page_size) = page_size {
        sailor.page_size(page_size);
    }
    let sailor = sailor
        .build()
        .map_err(|message| anyhow!(message))
        .context("build sailor")?;

    let products =
        sailor.get_sale_products().await.context("get sale items")?;
    products
        .iter()
        .for_each(|product| match product.discount() {
            Some(discount) => println!(
                "{} (${:.2}, you save ${:.2})",
                &product.name,
                product.price(),
                discount
            ),
            None => println!("{} (${:.2})", &product.name, product.price(),),
        });
    Ok(())
}
