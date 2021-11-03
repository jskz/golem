FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY src/*.go ./
COPY etc /app/etc
COPY migrations /app/migrations

RUN go build -o /app/golem

VOLUME ./etc
VOLUME ./scripts

EXPOSE 8080
EXPOSE 6060
EXPOSE 9000

CMD [ "./golem" ]