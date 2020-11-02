# api

_A personal GraphQL API, implemented in Rust._

> This branch is a **WIP**! For previous implementations in Go, see the [`v2`](https://github.com/stevenxie/api/tree/v2) branch.

<br />

## Development

### Setup

1. Ensure you have the required development dependencies:

    - [Rust](https://www.rust-lang.org/tools/install)
    - [Docker Engine](https://docs.docker.com/get-docker/)
    - [Docker Compose](https://docs.docker.com/compose/)

2. Clone and enter repo:

    ```bash
    git clone git@github.com:stevenxie/api
    cd api
    ```

3. Configure `.env`:

    ```bash
    cp .env.example .env  # copy template
    vi .env               # edit template
    ```

4. Install dependencies, setup hooks, and build target:

    ```bash
    cargo build
    ```

### Workflow

```bash
# Start dependencies (database, etc.):
docker-compose up -d

# Start server:
cargo run

# Format source code:
cargo fmt

# Run tests:
cargo test

# Check and lint source code:
cargo check && cargo clippy
```
