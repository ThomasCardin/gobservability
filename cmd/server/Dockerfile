FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -installsuffix cgo -o server ./cmd/server

FROM gcr.io/distroless/static-debian11

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/cmd/server/templates ./cmd/server/templates

EXPOSE 8080

CMD ["./server"]