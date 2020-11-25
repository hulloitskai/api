mod common {
    pub use crate::common::*;
    pub use crate::db::*;

    pub use graphql::{Context, Error as FieldError, Result as FieldResult};
    pub use graphql::{InputObject, MergedObject, Object, SimpleObject};

    pub use diesel::insert_into;
    pub use diesel::prelude::*;
    pub use tokio::task::spawn_blocking;

    pub fn format_error(error: Error) -> FieldError {
        let message = format!("{:#}", error);
        FieldError::new(message)
    }

    pub trait IntoFieldResult<T> {
        fn into_field_result(self) -> FieldResult<T>;
    }

    impl<T, R> IntoFieldResult<T> for R
    where
        R: Into<Result<T>>,
    {
        fn into_field_result(self) -> FieldResult<T> {
            let result: Result<T> = self.into();
            result.map_err(format_error)
        }
    }
}

pub mod query;
pub use query::*;

pub mod subscription;
pub use subscription::*;

pub mod meta;
pub use self::meta::*;

pub mod contact;
pub use contact::*;
