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
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/clinic /app/clinic
COPY templates /app/templates
COPY static /app/static
ENV PORT=80
EXPOSE 80
CMD ["/app/clinic"]
