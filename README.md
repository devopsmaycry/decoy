# Decoy

A lightweight honeypot service written in Go. It listens on configurable ports and logs all incoming connection attempts with structured JSON output. Supports SSH credential capture, full HTTP request logging, and generic TCP listeners — with optional forwarding to a syslog server.

## Features

- **SSH** — Captures username and password from every login attempt. Presents a realistic OpenSSH banner to avoid trivial fingerprinting.
- **HTTP** — Logs method, URI, query parameters, headers, and request body for every request.
- **TCP** — Logs the remote IP for any raw TCP connection.
- **Structured JSON logging** — Every event is written as a single JSON line to stdout and/or a syslog server.
- **Configurable via YAML** — All ports, listener types, and output options are controlled through a single config file.

## Requirements

- Go 1.25+ (for local builds)
- Docker (for container builds)

## Configuration

All configuration lives in `config/config.yaml`.

```yaml
version: "1.0"

# Define one or more listeners.
# Supported types: ssh, http, tcp
listeners:
  - port: "2222"
    type: ssh
  - port: "8080"
    type: http
  - port: "9000"
    type: tcp

# SSH-specific options
ssh:
  logUsername: true   # log the attempted username (false = log as ******** )
  logPassword: true   # log the attempted password (false = log as ******** )

# Output options
syslog:
  cliEnabled: true    # write JSON logs to stdout
  enabled: false      # forward logs to a syslog server via UDP
  server: "192.168.10.10"
  port: "514"
```

### Configuration reference

| Key | Type | Default | Description |
|---|---|---|---|
| `listeners[].port` | string | — | Port to listen on |
| `listeners[].type` | string | — | Listener type: `ssh`, `http`, or `tcp` |
| `ssh.logUsername` | bool | `false` | Log SSH usernames in plaintext |
| `ssh.logPassword` | bool | `false` | Log SSH passwords in plaintext |
| `syslog.cliEnabled` | bool | `true` | Enable stdout logging |
| `syslog.enabled` | bool | `false` | Enable syslog forwarding |
| `syslog.server` | string | — | Syslog server address |
| `syslog.port` | string | — | Syslog server UDP port |

## Running locally

```bash
go run . -config config/config.yaml
```

The `-config` flag defaults to `config/config.yaml` relative to the working directory.

## Docker

### Build the image

```bash
docker build -t decoy .
```

### Run the container

```bash
docker run -d \
  --name decoy \
  -p 2222:2222 \
  -p 8080:8080 \
  -p 9000:9000 \
  decoy
```

### Use a custom config

Mount your own config file to override the default:

```bash
docker run -d \
  --name decoy \
  -p 2222:2222 \
  -p 8080:8080 \
  -p 9000:9000 \
  -v $(pwd)/config/config.yaml:/config/config.yaml:ro \
  decoy
```

### View logs

```bash
docker logs -f decoy
```

## Example log output

```json
{"event":"decoy_started","listener_count":3,"time":"2026-03-24T10:00:00Z"}
{"event":"ssh_listening","port":"2222","time":"2026-03-24T10:00:00Z"}
{"event":"http_listening","port":"8080","time":"2026-03-24T10:00:00Z"}
{"event":"tcp_listening","port":"9000","time":"2026-03-24T10:00:00Z"}
{"client_version":"SSH-2.0-OpenSSH_8.2p1","event":"ssh_auth_attempt","password":"admin123","port":"2222","remote_ip":"1.2.3.4:54321","time":"2026-03-24T10:01:00Z","username":"root"}
{"body":"","event":"http_request","headers":{"User-Agent":"curl/7.88.1"},"method":"GET","port":"8080","remote_ip":"1.2.3.4:12345","time":"2026-03-24T10:02:00Z","uri":"/admin","query":""}
{"event":"tcp_connection","port":"9000","remote_ip":"1.2.3.4:9876","time":"2026-03-24T10:03:00Z"}
```
