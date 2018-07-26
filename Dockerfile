# Builder Stage (build stage for base image)
FROM golang:1.10-alpine as builder
MAINTAINER Jesse Lee jesse.lee@groundx.xyz

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /go-gxplatform
RUN cd /go-gxplatform && make gxp

# Container Stage (run gxp)
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-gxplatform/build/bin/gxp /usr/local/bin/

EXPOSE 8545 8546 30303 61001 30303/udp
ENTRYPOINT ["gxp"]
