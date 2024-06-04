FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go


FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY config/config.json config/
EXPOSE 8181
CMD ["./main"]
