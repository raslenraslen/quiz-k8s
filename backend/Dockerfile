FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN go build -o quiz-backend .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/quiz-backend .
EXPOSE 8080
CMD ["./quiz-backend"]
