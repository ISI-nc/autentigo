FROM golang:1.15-alpine as build-env
ENV CGO_ENABLED=0
WORKDIR /src
ADD go.mod go.sum ./
RUN go mod download
ADD . ./
RUN set -ex; \
  if grep -q '^package main *' *.go; then go install .; fi; \
  if [ -d cmd ]; then go install ./cmd/...; fi

FROM alpine:3.18.3
RUN apk add --no-cache curl tzdata
ENTRYPOINT ["/bin/autentigo"]
COPY --from=build-env /go/bin/ /bin/