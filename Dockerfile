FROM golang:1.26.5-alpine AS builder
WORKDIR /app
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN addgroup -S app && adduser -S -G app app
COPY --from=builder --chown=app:app /app/server .
USER app
EXPOSE 8080
CMD ["./server"]
