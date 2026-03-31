FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o decoy .

FROM scratch
COPY --from=builder /app/decoy /decoy
COPY --from=builder /app/config/config.yaml /config/config.yaml
# Run as unprivileged user (nobody: uid=65534).
# No /etc/passwd exists in scratch, so a numeric UID is required.
# Grant CAP_NET_BIND_SERVICE at runtime to allow ports < 1024 without root.
USER 65534:65534
ENTRYPOINT ["/decoy", "-config", "/config/config.yaml"]
