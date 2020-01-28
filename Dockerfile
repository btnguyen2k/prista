## Sample build command:
## docker build --force-rm --squash -t prista:0.1.0 .

FROM golang:1.13-alpine AS builder
MAINTAINER Thanh Nguyen <btnguyen2k@gmail.com>
RUN apk add git \
    && mkdir /build
COPY . /build
RUN cd /build && go build -o main

FROM alpine:3.10
RUN mkdir /app
COPY --from=builder /build/main /app/main
COPY --from=builder /build/config /app/config
RUN apk add --no-cache -U tzdata bash ca-certificates \
    && update-ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime \
    && chmod 711 /app/main \
    && rm -rf /var/cache/apk/*
WORKDIR /app
CMD ["/app/main"]
#ENTRYPOINT /app/main
