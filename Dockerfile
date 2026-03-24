FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o decoy .

FROM scratch
COPY --from=builder /app/decoy /decoy
COPY --from=builder /app/config/config.yaml /config/config.yaml
ENTRYPOINT ["/decoy", "-config", "/config/config.yaml"]
