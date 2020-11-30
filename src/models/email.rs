use crate::common::*;

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
            bail!("bad format");
        }
    }

    pub fn as_string(&self) -> &String {
        &self.0
    }

    pub fn as_str(&self) -> &str {
        &self.0
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

impl FromStr for Email {
    type Err = Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        Email::new(s.to_owned())
    }
}
