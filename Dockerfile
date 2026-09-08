# Build
FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/clinic ./cmd/web

# Run
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata postgresql-client
COPY --from=builder /out/clinic /app/clinic
COPY templates /app/templates
COPY static /app/static
COPY sborka /app/sborka
COPY scripts/migrate-and-run.sh /app/migrate-and-run.sh
RUN chmod +x /app/migrate-and-run.sh
ENV PORT=80
EXPOSE 80
CMD ["/app/migrate-and-run.sh"]
