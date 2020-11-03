use anyhow::{Context, Result};
use chrono::Local;
use git::{DescribeFormatOptions, DescribeOptions, Repository};
use semver::Version;

fn main() -> Result<()> {
    // Set build timestamp.
    set_env("BUILD_TIMESTAMP", &Local::now().to_rfc3339());

    // Set build version from git.
    let version = git_version();
    let version = match version {
        Ok(version) => version,
        Err(error) => {
            eprintln!("Failed to describe git version: {}", error);
            String::new()
        }
    };
    let version = fmt_version(&version);
    set_env("BUILD_VERSION", &version);

    Ok(())
}

fn git_version() -> Result<String> {
    let repo = Repository::open(".").context("open repository")?;
    let desc = repo
        .describe(
            DescribeOptions::default()
                .describe_tags()
                .show_commit_oid_as_fallback(true),
        )
        .context("describe HEAD")?;
    desc.format(Some(
        DescribeFormatOptions::default().dirty_suffix("-dirty"),
    ))
    .context("format describe result")
}

fn fmt_version(version: &str) -> String {
    let trimmed = if let Some(version) = version.strip_prefix("v") {
        version
    } else {
        return version.to_owned();
    };

    let version = if let Ok(version) = Version::parse(trimmed) {
        version
    } else {
        return trimmed.to_owned();
    };

    version.to_string()
}

fn set_env(key: &str, val: &str) {
    println!("cargo:rustc-env={}={}", key, val);
}
