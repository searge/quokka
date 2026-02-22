# syntax=docker/dockerfile:1.7

ARG GO_BUILDER_IMAGE=docker.io/library/golang:1.26-alpine
ARG RUNTIME_IMAGE=docker.io/library/alpine:3.21

FROM ${GO_BUILDER_IMAGE} AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/quokka-api ./cmd/api

FROM ${RUNTIME_IMAGE} AS runtime

RUN addgroup -S quokka && adduser -S -G quokka quokka && \
    apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/quokka-api /app/quokka-api

USER quokka

EXPOSE 8080

ENTRYPOINT ["/app/quokka-api"]
