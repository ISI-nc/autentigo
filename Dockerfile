#syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build-env
RUN apk upgrade --no-cache --available
ENV CGO_ENABLED=0
WORKDIR /src
COPY go.mod go.sum ./
ARG GOPROXY
ARG GONOSUMDB
RUN go mod download
COPY . ./
RUN set -ex; \
  go vet ./... ;\
  go install ./cmd/...

FROM alpine AS apps
RUN set -ex ;\
  apk upgrade --no-cache --available ;\
  apk add --no-cache curl tzdata
COPY --from=build-env /go/bin/ /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/"]
USER 1001
