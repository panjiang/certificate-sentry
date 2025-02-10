FROM --platform=${BUILDPLATFORM} docker.shiyou.kingsoft.com/library/golang:1.22.10 AS builder

WORKDIR /app
COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o certificate-sentry

FROM docker.shiyou.kingsoft.com/library/alpine:3.14.0

COPY --from=builder /app/certificate-sentry /app
COPY ./config.yaml /app
COPY ./*.tmpl /app

ENTRYPOINT ["/app/certificate-sentry"]
