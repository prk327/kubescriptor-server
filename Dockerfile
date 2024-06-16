# backend/Dockerfile
FROM golang:1.16-alpine

WORKDIR /kubescriptor

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /main

EXPOSE 8080

CMD ["/main"]