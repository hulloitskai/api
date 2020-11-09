use crate::prelude::*;

use std::fmt::{Display, Formatter, Result as FmtResult};

pub mod tnt;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Product {
    pub name: String,
    pub units: Option<String>,
    pub prices: Prices,
    pub vendor: String,
    pub vendor_id: String,
    pub vendor_sku: String,
}

impl Product {
    pub fn price(&self) -> &Decimal {
        self.prices.current()
    }

    pub fn discount(&self) -> Option<Decimal> {
        self.prices.discount()
    }
}

impl Display for Product {
    fn fmt(&self, f: &mut Formatter<'_>) -> FmtResult {
        f.write_fmt(format_args!("{} (${:.2})", &self.name, &self.price()))
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Prices {
    pub original: Decimal,
    pub sale: Option<Decimal>,
}

impl Prices {
    pub fn current(&self) -> &Decimal {
        match &self.sale {
            Some(amount) => amount,
            None => &self.original,
        }
    }

    pub fn discount(&self) -> Option<Decimal> {
        let sale = match &self.sale {
            Some(amount) => amount,
            None => return None,
        };
        Some(&self.original - sale)
    }
}
#[async_trait]
pub trait Sailor {
    async fn get_sale_products(&self) -> Result<Vec<Product>>;
}
