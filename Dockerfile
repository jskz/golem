FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY migrations /migrations
COPY src/*.go ./

RUN go build -o /golem

VOLUME /etc

EXPOSE 8080

CMD [ "/golem" ]