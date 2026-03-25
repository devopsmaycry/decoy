# Decoy

A lightweight honeypot service written in Go. It listens on configurable ports and logs all incoming connection attempts with structured JSON output. Supports SSH credential capture, full HTTP/HTTPS request logging, and generic TCP listeners — with optional forwarding to a syslog server.

## Features

- **SSH** — Captures username and password from every login attempt. Presents a configurable OpenSSH banner to avoid trivial fingerprinting.
- **HTTP** — Logs method, URI, query parameters, headers, and request body for every request.
- **HTTPS** — Same as HTTP with TLS. Provide your own certificate and key.
- **TCP** — Logs the remote IP and number of received bytes for any raw TCP connection.
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
# Add ssl: true to an http listener to enable TLS (requires https section below)
listeners:
  - port: "2222"
    type: ssh
  - port: "8080"
    type: http
  - port: "4343"
    type: http
    ssl: true
  - port: "9000"
    type: tcp

# SSH-specific options
ssh:
  logUsername: true   # log the attempted username (false = redact as ******** )
  logPassword: true   # log the attempted password (false = redact as ******** )
  sshShowedVersion: "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6"  # banner shown to clients

# HTTPS certificate (required when any listener has ssl: true)
https:
  serverCertificate: "/certs/server.crt"
  serverCertificateKey: "/certs/server.key"

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
| Key | Type | Default | Description |
|---|---|---|---|
| `listeners[].port` | string | — | Port to listen on |
| `listeners[].type` | string | — | Listener type: `ssh`, `http`, or `tcp` |
| `listeners[].ssl` | bool | `false` | Enable TLS on an `http` listener |
| `ssh.logUsername` | bool | `false` | Log SSH usernames in plaintext |
| `ssh.logPassword` | bool | `false` | Log SSH passwords in plaintext |
| `ssh.sshShowedVersion` | string | `SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6` | SSH banner shown to clients |
| `https.serverCertificate` | string | — | Path to TLS certificate file (PEM) |
| `https.serverCertificateKey` | string | — | Path to TLS private key file (PEM) |
| `syslog.cliEnabled` | bool | `true` | Enable stdout logging |
| `syslog.enabled` | bool | `false` | Enable syslog forwarding |
| `syslog.server` | string | — | Syslog server address |
| `syslog.port` | string | — | Syslog server UDP port |

## Running locally

```bash
go run . -config config/config.yaml
```

The `-config` flag defaults to `config/config.yaml` relative to the working directory.

## HTTPS setup

To use an HTTPS listener you need a TLS certificate and private key. For testing, generate a self-signed certificate with openssl:

```bash
openssl req -x509 -newkey ec \
  -pkeyopt ec_paramgen_curve:P-384 \
  -keyout server.key \
  -out server.crt \
  -days 3650 \
  -nodes \
  -subj "/CN=localhost"
```

| Flag | Description |
|---|---|
| `-newkey ec -pkeyopt ec_paramgen_curve:P-384` | ECDSA key on P-384 (modern, fast) |
| `-days 3650` | Valid for ~10 years |
| `-nodes` | No passphrase on the key (required for unattended startup) |
| `-subj "/CN=localhost"` | Minimal subject — adjust for your hostname |

Point the config to the generated files:

```yaml
https:
  serverCertificate: "server.crt"
  serverCertificateKey: "server.key"
```

> For a production honeypot use a real certificate (e.g. Let's Encrypt) so the TLS handshake looks legitimate to scanners.

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
  -p 4343:4343 \
  -p 9000:9000 \
  -v $(pwd)/config/config.yaml:/config/config.yaml:ro \
  decoy
```

### With HTTPS (mount certificates)

```bash
docker run -d \
  --name decoy \
  -p 2222:2222 \
  -p 8080:8080 \
  -p 4343:4343 \
  -p 9000:9000 \
  -v $(pwd)/config/config.yaml:/config/config.yaml:ro \
  -v $(pwd)/certs/server.crt:/certs/server.crt:ro \
  -v $(pwd)/certs/server.key:/certs/server.key:ro \
  decoy
```

Make sure the paths in `config.yaml` match the mount targets inside the container (e.g. `/certs/server.crt`).

### View logs

```bash
docker logs -f decoy
```

## Pre-compiled deployment

Build a static binary and deploy it directly on any Linux host without Docker or Go.

### Build

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o decoy .
```

For ARM64 (e.g. Raspberry Pi, AWS Graviton):

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o decoy .
```

### Ubuntu / Debian

```bash
# Copy binary and config
sudo cp decoy /usr/local/bin/decoy
sudo chmod 755 /usr/local/bin/decoy
sudo mkdir -p /etc/decoy
sudo cp config/config.yaml /etc/decoy/config.yaml

# Create dedicated system user (no login shell, no home dir)
sudo useradd --system --no-create-home --shell /usr/sbin/nologin decoy

# Allow binding to privileged ports (<1024) without root
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/decoy
```

Create the systemd unit file:

```bash
sudo tee /etc/systemd/system/decoy.service > /dev/null <<'EOF'
[Unit]
Description=Decoy Honeypot Service
After=network.target
Wants=network.target

[Service]
Type=simple
User=decoy
Group=decoy
ExecStart=/usr/local/bin/decoy -config /etc/decoy/config.yaml
Restart=on-failure
RestartSec=5s

# Hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/decoy
PrivateTmp=true
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now decoy
sudo systemctl status decoy

# Follow logs
sudo journalctl -u decoy -f
```

### Red Hat / Rocky Linux / AlmaLinux

Same steps as Ubuntu/Debian. Additionally open firewall ports:

```bash
sudo firewall-cmd --permanent --add-port=2222/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=4343/tcp
sudo firewall-cmd --permanent --add-port=9000/tcp
sudo firewall-cmd --reload
```

If SELinux is enforcing, label the binary and config:

```bash
sudo semanage fcontext -a -t bin_t '/usr/local/bin/decoy'
sudo restorecon -v /usr/local/bin/decoy
sudo semanage fcontext -a -t etc_t '/etc/decoy(/.*)?'
sudo restorecon -Rv /etc/decoy
```

### Alpine Linux

```bash
# Copy binary and config
cp decoy /usr/local/bin/decoy
chmod 755 /usr/local/bin/decoy
mkdir -p /etc/decoy
cp config/config.yaml /etc/decoy/config.yaml

# Create system user
adduser -S -H -s /sbin/nologin decoy

# Allow binding to privileged ports
setcap 'cap_net_bind_service=+ep' /usr/local/bin/decoy
```

Create an OpenRC init script:

```bash
cat > /etc/init.d/decoy <<'EOF'
#!/sbin/openrc-run

description="Decoy Honeypot Service"
command="/usr/local/bin/decoy"
command_args="-config /etc/decoy/config.yaml"
command_user="decoy"
pidfile="/run/decoy.pid"
command_background=true

depend() {
    need net
    after firewall
}
EOF

chmod +x /etc/init.d/decoy
rc-update add decoy default
rc-service decoy start

# Follow logs
tail -f /var/log/messages | grep decoy
```

---

## Logstash Integration

Decoy sends structured JSON over syslog (UDP). Below are Logstash pipeline examples for each event type.

### Input

```ruby
input {
  udp {
    port  => 514
    codec => plain
  }
}
```

### Filter

The syslog message wraps the JSON payload. The filter strips the syslog envelope and parses the JSON:

```ruby
filter {
  # Strip syslog header, extract the JSON part
  grok {
    match => { "message" => "%{SYSLOGTIMESTAMP} %{SYSLOGHOST} decoy: %{GREEDYDATA:json_payload}" }
  }

  json {
    source => "json_payload"
    target => "decoy"
  }

  # Map decoy.time to @timestamp
  date {
    match => [ "[decoy][time]", "ISO8601" ]
    target => "@timestamp"
  }

  # Enrich SSH auth attempts
  if [decoy][event] == "ssh_auth_attempt" {
    mutate {
      add_tag => [ "honeypot", "ssh" ]
      rename  => {
        "[decoy][remote_ip]"      => "[source][address]"
        "[decoy][username]"       => "[user][name]"
        "[decoy][password]"       => "[user][password]"
        "[decoy][client_version]" => "[user_agent][original]"
        "[decoy][port]"           => "[destination][port]"
      }
    }
  }

  # Enrich HTTP requests
  if [decoy][event] == "http_request" {
    mutate {
      add_tag => [ "honeypot", "http" ]
      rename  => {
        "[decoy][remote_ip]" => "[source][address]"
        "[decoy][method]"    => "[http][request][method]"
        "[decoy][uri]"       => "[url][path]"
        "[decoy][query]"     => "[url][query]"
        "[decoy][body]"      => "[http][request][body][content]"
        "[decoy][port]"      => "[destination][port]"
      }
    }
  }

  # Enrich generic TCP connections
  if [decoy][event] == "tcp_connection" {
    mutate {
      add_tag => [ "honeypot", "tcp" ]
      rename  => {
        "[decoy][remote_ip]" => "[source][address]"
        "[decoy][port]"      => "[destination][port]"
      }
    }
  }

  # Split source IP and port into separate fields
  if [source][address] {
    grok {
      match => { "[source][address]" => "%{IP:[source][ip]}:%{INT:[source][port]}" }
    }
  }

  mutate {
    remove_field => [ "message", "json_payload" ]
  }
}
```

### Output

#### Elasticsearch

```ruby
output {
  elasticsearch {
    hosts     => ["https://your-elasticsearch:9200"]
    index     => "decoy-honeypot-%{+YYYY.MM.dd}"
    user      => "logstash_writer"
    password  => "${LOGSTASH_ES_PASSWORD}"
  }
}
```

#### Splunk

Use the [Logstash output for Splunk HEC](https://github.com/splunk/logstash-output-splunk_hec):

```bash
logstash-plugin install logstash-output-splunk_hec
```

```ruby
output {
  splunk_hec {
    host  => "splunk.example.com"
    port  => 8088
    token => "${SPLUNK_HEC_TOKEN}"

    # Optional: route to a specific index and sourcetype
    index      => "honeypot"
    sourcetype => "decoy:json"

    # Use @timestamp from the filter stage
    use_ack => false
    ssl     => true
  }
}
```

In Splunk, search and report on events with SPL:

```spl
index=honeypot sourcetype="decoy:json"
| eval src_ip=replace(remote_ip, ":.*", "")
| stats count by event, src_ip
| sort -count

| where event="ssh_auth_attempt"
| table _time, remote_ip, username, password, client_version, port

| where event="http_request"
| table _time, remote_ip, method, uri, query, body, port, ssl
```

> Create a dedicated HEC token scoped to the `honeypot` index with `sourcetype=decoy:json` in Splunk Settings → Data Inputs → HTTP Event Collector.

#### Azure Log Analytics Workspace

Requires the [logstash-output-azure_loganalytics](https://github.com/Azure/Azure-Sentinel/tree/master/DataConnectors/Logstash) plugin:

```bash
logstash-plugin install logstash-output-azure_loganalytics
```

```ruby
output {
  azure_loganalytics {
    workspace_id  => "${AZURE_WORKSPACE_ID}"   # Log Analytics Workspace ID
    workspace_key => "${AZURE_WORKSPACE_KEY}"  # Primary or secondary key
    custom_log_table_name => "Decoy"           # Custom log table (suffix _CL added by Azure)
  }
}
```

Events will appear in the Log Analytics table `Decoy_CL`. Example KQL queries:

```kql
// All SSH credential attempts
Decoy_CL
| where event_s == "ssh_auth_attempt"
| project TimeGenerated, source_ip_s, user_name_s, user_password_s, user_agent_original_s
| order by TimeGenerated desc

// Top attacking IPs
Decoy_CL
| where event_s in ("ssh_auth_attempt", "http_request", "tcp_connection")
| summarize count() by source_ip_s
| order by count_ desc

// HTTP requests by URI
Decoy_CL
| where event_s == "http_request"
| summarize count() by url_path_s, http_request_method_s
| order by count_ desc
```

> **Note:** Azure appends `_s`, `_d`, or `_b` to field names depending on their type (string, double, boolean). The workspace key is sensitive — always pass it via an environment variable, never hardcode it.

### Resulting ECS fields per event type

| Event | ECS fields populated |
|---|---|
| `ssh_auth_attempt` | `source.ip`, `source.port`, `user.name`, `user.password`, `user_agent.original`, `destination.port` |
| `http_request` | `source.ip`, `source.port`, `http.request.method`, `url.path`, `url.query`, `http.request.body.content`, `destination.port` |
| `tcp_connection` | `source.ip`, `source.port`, `destination.port` |

---

## Example log output

```json
{"event":"decoy_started","listener_count":4,"time":"2026-03-24T10:00:00Z"}
{"event":"ssh_listening","port":"2222","time":"2026-03-24T10:00:00Z"}
{"event":"http_listening","port":"8080","ssl":false,"time":"2026-03-24T10:00:00Z"}
{"event":"http_listening","port":"4343","ssl":true,"time":"2026-03-24T10:00:00Z"}
{"event":"tcp_listening","port":"9000","time":"2026-03-24T10:00:00Z"}
{"client_version":"SSH-2.0-OpenSSH_8.2p1","event":"ssh_auth_attempt","password":"admin123","port":"2222","remote_ip":"1.2.3.4:54321","time":"2026-03-24T10:01:00Z","username":"root"}
{"body":"","event":"http_request","headers":{"User-Agent":"curl/7.88.1"},"method":"GET","port":"8080","remote_ip":"1.2.3.4:12345","ssl":false,"time":"2026-03-24T10:02:00Z","uri":"/admin","query":""}
{"body":"user=admin&pass=secret","event":"http_request","headers":{"User-Agent":"Mozilla/5.0"},"method":"POST","port":"4343","remote_ip":"1.2.3.4:54123","ssl":true,"time":"2026-03-24T10:02:05Z","uri":"/login","query":""}
{"event":"tcp_connection","port":"9000","raw_bytes":12,"remote_ip":"1.2.3.4:9876","time":"2026-03-24T10:03:00Z"}
```
