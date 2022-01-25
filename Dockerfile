# builder

FROM golang:1.17-alpine AS builder

WORKDIR /root/
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o miaokeeper .

# image

FROM alpine:latest
LABEL maintainer=BBAliance

# disabling cgo when built, so no need to install libc6-compat
RUN apk add --no-cache ca-certificates

WORKDIR /miaokeeper/
COPY --from=builder /root/miaokeeper /miaokeeper/miaokeeper
VOLUME /miaokeeper/

EXPOSE 9876

ENTRYPOINT ["/miaokeeper/miaokeeper"]
