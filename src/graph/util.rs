use super::common::*;

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
