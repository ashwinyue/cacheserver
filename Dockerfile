FROM golang:1.22-alpine AS builder

RUN apk add --no-cache make git

COPY . /src
WORKDIR /src

RUN go mod download
RUN make build

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./cacheserver", "-conf", "/data/conf"]
