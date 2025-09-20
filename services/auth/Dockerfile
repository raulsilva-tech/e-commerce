# Build stage
FROM golang:1.24-alpine AS build
WORKDIR /app
ENV CGO_ENABLED=0 \
GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /auth ./cmd/server

# Runner
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /auth /auth
EXPOSE 8080
ENTRYPOINT [ "/auth" ]
