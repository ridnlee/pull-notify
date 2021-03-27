FROM golang:alpine

WORKDIR /service

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build .

CMD ["./test-generator"]