FROM golang:1.20 AS builder

ENV GO111MODULE=on
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o foodtinder cmd/main.go

# run
FROM alpine:latest

RUN apk add --no-cache tzdata

COPY --from=builder /app/foodtinder /app/foodtinder
COPY --from=builder /app/docs/swaggerdocs/swagger.yaml /app/docs/swaggerdocs/swagger.yaml

EXPOSE 8080

WORKDIR /app

ENTRYPOINT ["./foodtinder"]