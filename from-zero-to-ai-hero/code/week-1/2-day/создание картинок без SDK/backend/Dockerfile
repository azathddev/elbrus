FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/banner-api .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && update-ca-certificates
RUN adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /bin/banner-api /app/banner-api

ENV PORT=8080
EXPOSE 8080

USER appuser
ENTRYPOINT ["/app/banner-api"]
