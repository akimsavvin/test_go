# Copyright (c) 2024, TrustMe
# All rights reserved.
# This file is the property of TrustMe and contains confidential information.
# Unauthorized distribution or use of this file is strictly prohibited without explicit written permission from TrustMe.

FROM golang:1.23.0 as build

WORKDIR /build

ADD ./go.mod ./go.sum ./
RUN go mod download

LABEL maintainer="Akim Savvin <akim@savvin.io>"
WORKDIR /app

COPY . .

ENV CGO_ENABLED=0 GOOS=linux

RUN go build -o ./binary/lure ./cmd/lure


FROM alpine:latest as run

COPY --from=build /build/binary/lure ./lure

ENTRYPOINT ["./lure"]