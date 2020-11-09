use super::{Prices, Product, Sailor};
use crate::prelude::*;

use json::Value as JsonValue;
use json_dotpath::DotPaths;

use bigdecimal::Zero;
use request::Client;
use std::fmt::{Display, Formatter, Result as FmtResult};
use tokio_compat::FutureExt;

/// A `TntSailor` finds sales at T&T Supermarket.
#[derive(Builder)]
pub struct TntSailor {
    #[builder(default = "build_client()")]
    client: Client,

    #[builder(default = "25")]
    page_size: u32,
}

fn build_client() -> Client {
    Client::builder().cookie_store(true).build().unwrap()
}

impl TntSailor {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn builder() -> TntSailorBuilder {
        TntSailorBuilder::default()
    }
}

impl Default for TntSailor {
    fn default() -> Self {
        Self::builder().build().unwrap()
    }
}

const TNT_API_URL: &str = "https://www.tntsupermarket.com/rest/V1";

#[async_trait]
impl Sailor for TntSailor {
    async fn get_sale_products(&self) -> Result<Vec<Product>> {
        // Set location.
        // TODO: Split this into a separate helper function.
        let url =
            format!("{}/tntzone/location/getpreferedstorecode", TNT_API_URL);
        let postcode = "N2L";
        let response = self
            .client
            .get(&url)
            .query(&[("postcode", &postcode)])
            .send()
            .compat()
            .await
            .context("send request")?;
        if !response.status().is_success() {
            bail!("bad response");
        }

        // Get weekly specials.
        let url = format!("{}/xmapi/app-weekly-special", TNT_API_URL);
        let page = "1";
        let page_size = self.page_size.to_string();
        let response = self
            .client
            .get(&url)
            .query(&[("page", page), ("pageSize", &page_size)])
            .send()
            .compat()
            .await
            .context("send request")?;
        if !response.status().is_success() {
            bail!("bad response");
        }
        let value: JsonValue =
            response.json().await.context("parse response")?;

        let products: Vec<TntProduct> =
            match value.dot_get("data.category.items") {
                Ok(products) => products.ok_or_else(|| anyhow!("missing data")),
                Err(error) => Err(error.into()),
            }
            .context("parse products")?;
        let products: Vec<TntProduct> = products
            .into_iter()
            .filter(|product| {
                if (product.is_available == 0) {
                    info!("Found an unavailable product: {}", product);
                    return false;
                }
                if (product.is_saleable == 0) {
                    info!("Found an unsellable product: {}", product);
                    return false;
                }
                if (product.prices.final_price.amount.is_zero()) {
                    info!("Found an unpriced product: {}", product);
                    return false;
                }
                true
            })
            .collect();
        let products: Vec<Product> =
            products.into_iter().map(Into::into).collect();
        Ok(products)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct TntProduct {
    id: String,
    sku: String,
    name: String,
    prices: TntPrices,
    weight_uom: String,
    is_available: u8,
    is_saleable: u8,
}

impl Display for TntProduct {
    fn fmt(&self, f: &mut Formatter<'_>) -> FmtResult {
        f.write_fmt(format_args!("{} (SKU: {})", &self.name, &self.sku))
    }
}

impl From<TntProduct> for Product {
    fn from(
        TntProduct {
            id,
            name,
            prices,
            weight_uom,
            sku,
            ..
        }: TntProduct,
    ) -> Self {
        Product {
            name,
            units: Some(weight_uom).filter(|string| !string.is_empty()),
            prices: prices.into(),
            vendor: "T&T".to_owned(),
            vendor_id: id,
            vendor_sku: sku,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct TntPrices {
    old_price: TntPrice,
    final_price: TntPrice,
}

impl From<TntPrices> for Prices {
    fn from(
        TntPrices {
            old_price,
            final_price,
        }: TntPrices,
    ) -> Self {
        let orig = old_price.amount;
        let sale = final_price.amount;
        Prices {
            sale: if sale != orig { Some(sale) } else { None },
            original: orig,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct TntPrice {
    amount: Decimal,
}
