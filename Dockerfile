FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gauditor ./cmd/gauditor

FROM alpine:3.20
RUN adduser -D -g '' app
USER app
COPY --from=build /out/gauditor /usr/local/bin/gauditor
EXPOSE 8091
ENTRYPOINT ["gauditor", "-addr", ":8091"]


