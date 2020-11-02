use anyhow::{Context, Result};
use chrono::Local;
use git::{DescribeFormatOptions, DescribeOptions, Repository};

fn main() -> Result<()> {
    // Set build timestamp.
    set_env("BUILD_TIMESTAMP", &Local::now().to_rfc3339());

    // Set build version from git.
    set_env(
        "BUILD_VERSION",
        match &git_version() {
            Ok(version) => version,
            Err(error) => {
                eprintln!("Failed to describe git version: {}", error);
                ""
            }
        },
    );

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

fn set_env(key: &str, val: &str) {
    println!("cargo:rustc-env={}={}", key, val);
}
