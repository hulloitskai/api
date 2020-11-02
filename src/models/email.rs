use crate::prelude::*;

lazy_static! {
    static ref EMAIL_REGEX: Regex = Regex::new(
        r"^([a-z0-9_+]([a-z0-9_+.]*[a-z0-9_+])?)@([a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,6})",
    )
    .unwrap();
}

#[derive(Debug, Clone, Hash, Serialize, Deserialize)]
pub struct Email(String);

impl Email {
    pub fn new(s: String) -> Result<Self> {
        if EMAIL_REGEX.is_match(&s) {
            Ok(Email(s))
        } else {
            Err(format_err!("invalid email address"))
        }
    }
}

impl ToString for Email {
    fn to_string(&self) -> String {
        self.0.to_owned()
    }
}

impl From<Email> for String {
    fn from(email: Email) -> Self {
        email.0
    }
}
