FROM golang:1.25.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app ./cmd/app

FROM alpine:3.22

COPY --from=builder /out/app /usr/local/bin/app
COPY --from=builder /src/internal/static/files /var/lib/confectionery/static

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/app"]
