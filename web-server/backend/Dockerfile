FROM golang:1.24-alpine AS builder

WORKDIR /backend

COPY ./backend .

RUN go build ./cmd/main.go

FROM alpine AS runner

# Install curl for health checks
RUN apk add --no-cache curl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /backend/main /main
COPY --from=builder /backend/config/config.yaml /config.yaml

# Copy frontend files
COPY ./frontend-simple /frontend-simple

EXPOSE 8080
EXPOSE 8081

CMD ["/main"]
