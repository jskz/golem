FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY migrations /migrations
COPY src/*.go ./
COPY etc /app/etc

RUN go build -o /app/golem

VOLUME /etc
VOLUME /scripts

EXPOSE 8080

CMD [ "./golem" ]