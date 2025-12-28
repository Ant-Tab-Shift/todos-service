FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd


FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server .
CMD ["./server"]
