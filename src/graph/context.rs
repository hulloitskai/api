use diesel::r2d2::{ConnectionManager, Pool};
use diesel::PgConnection;
use graphql::Context as GQLContext;

pub type DbPool = Pool<ConnectionManager<PgConnection>>;

pub struct Context {
    pub db: DbPool,
}

impl Context {
    pub fn new(db: DbPool) -> Context {
        Context { db }
    }

    pub fn from<'a>(ctx: &'a GQLContext) -> &'a Self {
        ctx.data_unchecked::<Self>()
    }
}
