FROM golang:1.23.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/config/.env /config/.env

RUN chmod +x ./main

EXPOSE 8080

CMD ["./main"]
