# syntax=docker/dockerfile:1
FROM golang:1.17-alpine
ENV PORT 8080
ENV HOSTDIR 0.0.0.0

EXPOSE 8080
WORKDIR /app

COPY src/data_service/go.mod ./ 
COPY src/data_service/go.sum ./

RUN go mod tidy

COPY src/data_service/. .

RUN go build -o /main
CMD ["/main"]
