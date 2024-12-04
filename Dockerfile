FROM golang:1.23.3-alpine3.20 AS builder

COPY . /github.com/titoffon/auth-service-task-BackDev/
WORKDIR /github.com/titoffon/auth-service-task-BackDev/

RUN go mod download
RUN go build -o ./bin/auth cmd/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /github.com/titoffon/auth-service-task-BackDev/bin/auth .
COPY --from=builder /github.com/titoffon/auth-service-task-BackDev/cmd/.env .

CMD ["./auth"]