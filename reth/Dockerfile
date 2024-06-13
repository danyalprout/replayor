FROM rust:1.78
WORKDIR /app

RUN apt-get update && apt-get -y upgrade && apt-get install -y git libclang-dev pkg-config curl build-essential
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

ENV REPO=https://github.com/paradigmxyz/reth.git
RUN git clone $REPO . && git checkout 5e778317fb6ef0809fbd399a583ac6fb9befad2b

# Note: Not including `asm-keccak` since it's not supported for some local machines
RUN cargo build --bin op-reth --locked --features jemalloc,optimism --profile maxperf

RUN cp /app/target/maxperf/op-reth /app/op-reth

COPY ./op-reth.sh /app/op-reth.sh
COPY ./reth.toml /app/reth.toml

ENTRYPOINT ["/app/op-reth.sh"]
