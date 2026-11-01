FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /rate-limiter ./cmd/server
FROM alpine:3.20
COPY --from=build /rate-limiter /rate-limiter
EXPOSE 8080
ENTRYPOINT ["/rate-limiter"]
