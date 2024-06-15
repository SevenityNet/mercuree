FROM golang:1.22-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o mercuree .

FROM alpine:latest as runtime

COPY --from=builder /app/mercuree /usr/local/bin/mercuree

ENTRYPOINT [ "mercuree" ]