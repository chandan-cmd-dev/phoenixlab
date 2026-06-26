FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o phoenixlab ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/phoenixlab .

COPY --from=builder /app/conf      ./conf
COPY --from=builder /app/views     ./views
COPY --from=builder /app/static    ./static
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./phoenixlab"]
