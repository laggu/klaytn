# Builder Stage (build stage for base image)
FROM golang:1.10-alpine as builder
MAINTAINER Jesse Lee jesse.lee@groundx.xyz

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /klaytn
RUN cd /klaytn && make klay

# Container Stage (run klay)
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /klaytn/build/bin/klay /usr/local/bin/

EXPOSE 8545 8546 30303 61001 30303/udp
ENTRYPOINT ["klay"]
