# request-service/Dockerfile
FROM golang:1.20-alpine
WORKDIR /app
COPY . .
RUN go mod download && go build -o request-service
EXPOSE 50025
CMD ["./request-service"]

