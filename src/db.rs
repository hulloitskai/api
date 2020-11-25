pub use diesel::r2d2::{
    ConnectionManager as DbConnectionManager, Pool as DbPool,
};
pub use diesel::{Connection as DbConnection, PgConnection};

pub type PgPool = DbPool<DbConnectionManager<PgConnection>>;
