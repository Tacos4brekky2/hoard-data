# data-service/Dockerfile
FROM golang:1.20-alpine
WORKDIR /app
COPY . .
RUN go mod download && go build -o data-service
EXPOSE 50052
CMD ["./data-service"]
