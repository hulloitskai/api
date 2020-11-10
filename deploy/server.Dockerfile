# == BUILDER ==
FROM ekidd/rust-musl-builder:1.47.0 AS builder

# Compile dependencies:
WORKDIR /src
RUN sudo chown rust:rust ./ && \
    USER=rust cargo init --bin . && \
    mkdir -p src/bin/server && \
    mv src/main.rs src/bin/server/main.rs
COPY --chown=rust:rust Cargo.toml Cargo.lock ./
RUN cargo build --bin server --release --target x86_64-unknown-linux-musl && \
    rm -rf src

# Copy source:
COPY --chown=rust:rust .git/ .git/
COPY --chown=rust:rust src/ src/
COPY --chown=rust:rust build.rs ./

# Build binaries:
ENV BUILD_VERSION_DIRTY_SUFFIX=""
RUN cargo build --bin server --release --target x86_64-unknown-linux-musl


# == RUNNER ==
FROM alpine:3.12

# Install system dependencies:
RUN apk add --update ca-certificates curl

# Copy built binary:
ENV CMD=server
COPY --from=builder /src/target/x86_64-unknown-linux-musl/release/${CMD} /usr/local/bin/${CMD}

# Configure ports:
ENV API_PORT=80
EXPOSE $API_PORT

# Configure healthcheck and entrypoint:
HEALTHCHECK --interval=10s --timeout=1s --start-period=5s --retries=3 CMD curl -f http://localhost || exit 1
ENTRYPOINT $CMD
