FROM --platform=${BUILDPLATFORM} docker.shiyou.kingsoft.com/library/golang:1.22.10 AS builder

WORKDIR /app
COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o cert-checker

FROM docker.shiyou.kingsoft.com/library/alpine:3.14.0

COPY --from=builder /app/cert-checker /app
COPY ./config.yaml /app
COPY ./*.tmpl /app

ENTRYPOINT ["/app/cert-checker"]
