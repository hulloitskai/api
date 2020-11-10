# == BUILDER ==
FROM ekidd/rust-musl-builder:1.47.0 AS builder

# Compile dependencies:
WORKDIR /src
RUN sudo chown rust:rust ./ && USER=rust cargo init --bin .
COPY --chown=rust:rust Cargo.toml Cargo.lock ./
RUN cargo build --release --target x86_64-unknown-linux-musl && rm -rf src

# Copy source:
COPY --chown=rust:rust .git/ .git/
COPY --chown=rust:rust src/ src/
COPY --chown=rust:rust build.rs ./

# Build binaries:
ENV BUILD_VERSION_DIRTY_SUFFIX=""
RUN cargo build --bin api --release --target x86_64-unknown-linux-musl


# == RUNNER ==
FROM alpine:3.12

# Install system dependencies:
RUN apk add --update ca-certificates curl

# Copy built binary:
COPY --from=builder /src/target/x86_64-unknown-linux-musl/release/api /bin/api

# Configure ports:
ENV API_SERVER_PORT=80
EXPOSE $API_SERVER_PORT

# Configure healthcheck and entrypoint:
HEALTHCHECK --interval=10s --timeout=1s --start-period=5s --retries=3 CMD curl -f http://localhost/ || exit 1
ENTRYPOINT $CMD
