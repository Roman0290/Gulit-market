FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /pocket-market-api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /pocket-market-api /app/pocket-market-api
COPY --from=build /app/public ./public
EXPOSE 8080
ENTRYPOINT ["/app/pocket-market-api"]
