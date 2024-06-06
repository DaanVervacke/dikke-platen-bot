# Build
FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/

# Run
FROM alpine:latest

RUN apk --no-cache add ca-certificates bash curl

SHELL ["/bin/bash", "-c"]

COPY --from=build /main /main

EXPOSE 8080

CMD ["/main"]