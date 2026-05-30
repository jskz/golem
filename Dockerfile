  FROM golang:1.26.3-alpine AS build

  WORKDIR /src

  COPY go.mod go.sum ./
  RUN go mod download

  COPY src ./src
  RUN CGO_ENABLED=0 go build -o /out/golem ./src

  FROM alpine:3.22

  WORKDIR /app

  RUN adduser -D -H -s /sbin/nologin golem

  COPY --from=build /out/golem /app/golem
  COPY etc /app/etc
  COPY migrations /app/migrations
  COPY scripts /app/scripts

  USER golem

  EXPOSE 4000 6060 9000

  CMD ["/app/golem"]