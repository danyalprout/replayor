FROM golang:1.21 as geth

WORKDIR /app

ENV REPO=https://github.com/danyalprout/op-geth.git

RUN git clone $REPO . && git checkout 882d0f22749bbc36ae17c414fb7dde06f6e91035 && go run build/ci.go install -static ./cmd/geth

FROM golang:1.21

RUN apt-get update && \
    apt-get install -y jq curl && \
    rm -rf /var/lib/apt/lists

WORKDIR /app

COPY --from=geth /app/build/bin/geth ./