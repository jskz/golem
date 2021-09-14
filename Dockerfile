FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY migrations /migrations
COPY src/*.go ./
COPY etc/greeting.ansi /greeting.ansi
COPY etc/motd.ansi /motd.ansi
COPY etc/death.ansi /death.ansi

RUN go build -o /golem

VOLUME /etc
VOLUME /scripts

EXPOSE 8080

CMD [ "/golem" ]