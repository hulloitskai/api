pub mod prelude;

pub use diesel::r2d2::{ConnectionManager, Pool};
pub use diesel::PgConnection;
pub type DbPool = Pool<ConnectionManager<PgConnection>>;

pub mod build;
pub mod contact;
pub mod query;
pub mod subscription;

pub use build::*;
pub use contact::*;
pub use query::*;
pub use subscription::*;
