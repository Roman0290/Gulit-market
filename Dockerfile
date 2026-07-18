FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /pocket-market-api ./cmd/api

FROM alpine:3.20
COPY --from=build /pocket-market-api /pocket-market-api
EXPOSE 8080
ENTRYPOINT ["/pocket-market-api"]
