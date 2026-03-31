# Decoy

A lightweight honeypot service written in Go. It listens on configurable ports and logs all connection attempts as structured JSON. Designed to detect reconnaissance, credential stuffing, and lateral movement attempts inside a network.

## Features

- **SSH** — Captures username and password from every login attempt. Configurable server version banner to avoid trivial fingerprinting. Credentials can be logged in plaintext or redacted.
- **HTTP / HTTPS** — Logs method, URI, query parameters, all headers, and the full request body. POST submissions are parsed for `username` and `password` fields. A realistic login page is served to keep attackers engaged. Any request to a path not matching the configured login path is logged as `http_probe`.
- **TCP service emulation** — Sends realistic protocol banners for SMTP, FTP, Redis, MySQL (Protocol v10 handshake), and MSSQL (TDS Pre-Login response). Logs the first 4 KB received from the client.
- **Rate limiting** — Built-in protection against log flooding and resource exhaustion: max 500 concurrent TCP/SSH connections globally, max 30 connections per IP per minute.
- **Structured JSON logging** — Every event is a single JSON line written to stdout and/or forwarded to a syslog server via UDP.
- **Proxy-aware** — HTTP listeners extract the real client IP from `X-Forwarded-For` and `X-Real-IP` headers when running behind a load balancer or reverse proxy.
- **Configurable via YAML** — All ports, banners, and output options are controlled through a single config file.
- **Static binary / container** — Ships as a single ~6 MB static binary. Docker image based on `scratch`.

## Requirements

- Go 1.25+ (for local builds)
- Docker (for container builds)

## Quick start

```bash
git clone https://github.com/your-org/decoy
cd decoy
go run . -config config/config.yaml
```

## Configuration

All configuration lives in `config/config.yaml`. The config version must be `"1.2"`.

### Minimal example

```yaml
version: "1.2"

listeners:
  - port: "2222"
    type: ssh

httpListeners:
  - port: "8080"
    path: "/admin/login"
    websiteEnabled: true

syslog:
  cliEnabled: true
```

### Full reference example

```yaml
version: "1.2"

# SSH and TCP listeners
listeners:
  - port: "2222"
    type: ssh
  - port: "25"
    type: tcp
    service: smtp
  - port: "21"
    type: tcp
    service: ftp
  - port: "6379"
    type: tcp
    service: redis
  - port: "3306"
    type: tcp
    service: mysql
  - port: "1433"
    type: tcp
    service: mssql

# HTTP/HTTPS listeners (separate section — not part of listeners[])
httpListeners:
  - port: "8080"
    path: "/admin/login"
    Server: "Apache/2.4.51 (Debian)"
    X-Powered-By: "PHP/7.4.33"
    websiteEnabled: true
  - port: "4343"
    path: "/admin/login"
    websiteEnabled: true
    sslEnabled: true
    serverCertificate: "/certs/server.crt"
    serverCertificateKey: "/certs/server.key"

ssh:
  logUsername: true
  logPassword: true
  sshShowedVersion: "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6"

service:
  ftpBanner:   "220 Microsoft FTP Service"
  redisBanner: "-NOAUTH Authentication required."
  smtpBanner:  "220 mail.corp.local ESMTP Postfix (Debian/GNU)"

syslog:
  cliEnabled: true
  enabled: false
  server: "192.168.10.10"
  port: "514"
```

### Configuration reference

#### `listeners[]` — SSH and TCP

| Field | Type | Required | Description |
|---|---|---|---|
| `port` | string | yes | TCP port to listen on |
| `type` | string | yes | `ssh` or `tcp` |
| `service` | string | only for `tcp` | Service to emulate: `smtp`, `ftp`, `redis`, `mysql`, `mssql` |

> HTTP listeners are defined separately in `httpListeners[]`, not here.

#### `httpListeners[]` — HTTP / HTTPS

| Field | Type | Default | Description |
|---|---|---|---|
| `port` | string | — | TCP port to listen on |
| `path` | string | — | URL path that serves the login page (required) |
| `Server` | string | `Apache/2.2.22 (Debian)` | `Server` response header |
| `X-Powered-By` | string | `PHP/5.6.40` | `X-Powered-By` response header |
| `websiteEnabled` | bool | `false` | Serve the built-in login page HTML |
| `redirectUrl` | string | — | Redirect all GET requests to this URL |
| `sslEnabled` | bool | `false` | Enable TLS |
| `serverCertificate` | string | — | Path to PEM certificate (required when `sslEnabled: true`) |
| `serverCertificateKey` | string | — | Path to PEM private key (required when `sslEnabled: true`) |

#### `ssh`

| Field | Type | Default | Description |
|---|---|---|---|
| `logUsername` | bool | `false` | Log SSH usernames in plaintext (`false` → redacted as `********`) |
| `logPassword` | bool | `false` | Log SSH passwords in plaintext (`false` → redacted as `********`) |
| `sshShowedVersion` | string | `SSH-2.0-OpenSSH_8.9p1 Debian-3` | SSH banner presented to clients |

#### `service` — TCP banner overrides

| Field | Type | Default | Description |
|---|---|---|---|
| `ftpBanner` | string | `220 Microsoft FTP Service` | FTP greeting (without `\r\n`) |
| `redisBanner` | string | `-NOAUTH Authentication required.` | Redis response (without `\r\n`) |
| `smtpBanner` | string | `220 mail.corp.local ESMTP Postfix (Debian/GNU)` | SMTP greeting (without `\r\n`) |

MySQL and MSSQL banners are fixed binary protocol responses and cannot be overridden via config.

#### `syslog`

| Field | Type | Default | Description |
|---|---|---|---|
| `cliEnabled` | bool | `true` | Write JSON log lines to stdout |
| `enabled` | bool | `false` | Forward logs to a syslog server via UDP |
| `server` | string | — | Syslog server address |
| `port` | string | — | Syslog server UDP port |

---

## Log events

Every log entry is a JSON object with at minimum a `time` (RFC 3339 UTC) and `event` field.

### Lifecycle

| Event | Description |
|---|---|
| `decoy_started` | Service started. Includes `listener_count`. |
| `decoy_stopped` | Service stopped (SIGTERM / SIGINT). |
| `ssh_listening` | SSH listener ready. |
| `http_listening` | HTTP listener ready. |
| `tcp_listening` | TCP listener ready. |

### SSH

| Event | Fields | Description |
|---|---|---|
| `ssh_auth_attempt` | `port`, `remote_ip`, `username`, `password`, `client_version` | Login attempt received. Username/password are redacted if logging is disabled. |
| `ssh_rate_limited` | `port`, `remote_ip` | Connection dropped — per-IP rate limit (30/min) or global limit (500 concurrent) exceeded. |
| `ssh_accept_error` | `port`, `error` | Error accepting a new TCP connection. |

### HTTP

| Event | Fields | Description |
|---|---|---|
| `http_request` | `port`, `remote_ip`, `method`, `uri`, `query`, `headers`, `body`, `ssl`, `username`*, `password`* | Request to the configured login path. `username`/`password` are extracted from POST form body. |
| `http_probe` | `port`, `remote_ip`, `method`, `uri`, `ssl` | Request to any path other than the configured login path. |
| `http_rate_limited` | `port`, `remote_ip` | Request dropped — per-IP rate limit exceeded. |

*\* Only present on POST requests with a URL-encoded body.*

### TCP

| Event | Fields | Description |
|---|---|---|
| `tcp_connection` | `port`, `remote_ip`, `service`, `data`* | Client connected and sent data after receiving the banner. |
| `tcp_rate_limited` | `port`, `remote_ip`, `service` | Connection dropped — rate limit exceeded. |
| `tcp_write_error` | `port`, `remote_ip`, `error` | Failed to send the banner to the client. |

*\* Only present when the client sent data.*

---

## Example log output

```json
{"event":"decoy_started","listener_count":3,"time":"2026-03-31T08:00:00Z"}
{"event":"ssh_listening","port":"2222","time":"2026-03-31T08:00:00Z"}
{"event":"http_listening","port":"8080","ssl":false,"time":"2026-03-31T08:00:00Z"}
{"event":"tcp_listening","port":"25","service":"smtp","time":"2026-03-31T08:00:00Z"}
{"client_version":"SSH-2.0-PuTTY_Release_0.79","event":"ssh_auth_attempt","password":"admin123","port":"2222","remote_ip":"203.0.113.42","time":"2026-03-31T08:01:14Z","username":"root"}
{"event":"http_probe","method":"GET","port":"8080","remote_ip":"203.0.113.42","ssl":false,"time":"2026-03-31T08:01:20Z","uri":"/.env"}
{"body":"username=admin&password=letmein","event":"http_request","headers":{"User-Agent":"Mozilla/5.0","Content-Type":"application/x-www-form-urlencoded"},"method":"POST","password":"letmein","port":"8080","query":"","remote_ip":"203.0.113.42","ssl":false,"time":"2026-03-31T08:01:21Z","uri":"/admin/login","username":"admin"}
{"data":"EHLO attacker.example.com\r\n","event":"tcp_connection","port":"25","remote_ip":"203.0.113.7","service":"smtp","time":"2026-03-31T08:02:05Z"}
{"event":"ssh_rate_limited","port":"2222","remote_ip":"198.51.100.9","time":"2026-03-31T08:02:30Z"}
```

---

## Running locally

```bash
go run . -config config/config.yaml
```

The `-config` flag defaults to `config/config.yaml` relative to the working directory.

---

## HTTPS setup

Generate a self-signed certificate for testing:

```bash
openssl req -x509 -newkey ec \
  -pkeyopt ec_paramgen_curve:P-384 \
  -keyout server.key \
  -out server.crt \
  -days 3650 \
  -nodes \
  -subj "/CN=localhost"
```

Point the config to the generated files:

```yaml
httpListeners:
  - port: "4343"
    path: "/admin/login"
    websiteEnabled: true
    sslEnabled: true
    serverCertificate: "server.crt"
    serverCertificateKey: "server.key"
```

> For a production honeypot, use a real certificate (e.g. Let's Encrypt) so the TLS handshake looks legitimate to automated scanners.

---

## Binding to privileged ports (< 1024)

To listen on standard ports like 22, 25, or 80 without running as root, grant the binary the `CAP_NET_BIND_SERVICE` capability:

```bash
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/decoy
```

This is sufficient for a dedicated system user. No need to run as root or use `sudo`.

---

## Hardening

A honeypot is an intentionally exposed service. Hardening the container and host is critical: if an attacker finds a vulnerability in the process itself, the blast radius must be as small as possible.

### Why block outbound traffic?

Decoy accepts connections but never needs to initiate them. Blocking all outbound connections from the process prevents:

- Data exfiltration if the binary is exploited
- The honeypot being used as a scanning or attack relay
- DNS callbacks or C2 beaconing from exploited vulnerabilities

---

### Container hardening (Docker)

Use the hardened `docker run` flags in production. The image already runs as `UID 65534` (nobody) — the flags below add additional isolation layers:

```bash
docker run -d \
  --name decoy \
  --read-only \
  --security-opt no-new-privileges:true \
  --cap-drop ALL \
  --cap-add NET_BIND_SERVICE \
  --pids-limit 200 \
  --memory 128m \
  --ulimit nofile=4096:4096 \
  -p 2222:2222 \
  -p 8080:8080 \
  -v $(pwd)/config/config.yaml:/config/config.yaml:ro \
  decoy
```

| Flag | Purpose |
|---|---|
| `--read-only` | Container filesystem is read-only; prevents writing exploit payloads |
| `--security-opt no-new-privileges:true` | Blocks `setuid`/`setgid` privilege escalation |
| `--cap-drop ALL` | Drops all Linux capabilities |
| `--cap-add NET_BIND_SERVICE` | Restores only the capability to bind ports < 1024. Remove if all ports are ≥ 1024. |
| `--pids-limit 200` | Prevents fork bombs |
| `--memory 128m` | Caps RAM usage |
| `--ulimit nofile=4096:4096` | Caps open file descriptors (connections) |

#### Block outbound traffic — iptables (Docker host)

Docker's `DOCKER-USER` chain runs before Docker's own rules and applies to all traffic from the bridge. The rules below allow responses to inbound connections but drop any new connection the container initiates:

```bash
# Allow responses to inbound connections (established / related)
sudo iptables -I DOCKER-USER -i docker0 ! -o docker0 \
  -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Drop all new outbound connections initiated by the container
sudo iptables -I DOCKER-USER -i docker0 ! -o docker0 \
  -m conntrack --ctstate NEW -j DROP
```

Make the rules persistent across reboots:

```bash
# Debian / Ubuntu
sudo apt-get install -y iptables-persistent
sudo netfilter-persistent save

# Red Hat / Rocky / AlmaLinux
sudo service iptables save
```

If you use a named Docker network instead of the default `docker0` bridge, find the bridge interface name first:

```bash
# Replace "decoy-net" with your network name
BRIDGE="br-$(docker network inspect decoy-net -f '{{.Id}}' | cut -c1-12)"
echo $BRIDGE   # e.g. br-a1b2c3d4e5f6
```

Then substitute `docker0` with `$BRIDGE` in the iptables commands above.

---

### Bare-metal / VM hardening

#### Block outbound traffic — nftables (Linux ≥ 5.x, recommended)

This rule matches traffic by the OS user running the decoy process. New outbound connections are dropped; responses to inbound probes are allowed:

```bash
sudo tee /etc/nftables.d/decoy.nft > /dev/null <<'EOF'
table inet decoy_outbound {
    chain output {
        type filter hook output priority 0; policy accept;
        meta skuid "decoy" ct state new drop
        meta skuid "decoy" ct state established,related accept
    }
}
EOF

# Load immediately
sudo nft -f /etc/nftables.d/decoy.nft

# Make persistent (include in main config)
echo 'include "/etc/nftables.d/decoy.nft"' | sudo tee -a /etc/nftables.conf
```

#### Block outbound traffic — iptables

```bash
# Allow responses to inbound connections
sudo iptables -A OUTPUT -m owner --uid-owner decoy \
  -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Drop all new outbound connections from the decoy process
sudo iptables -A OUTPUT -m owner --uid-owner decoy \
  -m conntrack --ctstate NEW -j DROP

# Persist
sudo netfilter-persistent save   # Debian/Ubuntu
sudo service iptables save       # Red Hat/Rocky
```

#### Block outbound traffic — ufw

ufw does not natively support `--uid-owner` matching. Use the nftables or iptables approach above instead. If ufw is active, it will preserve these custom rules as long as they are inserted into the `OUTPUT` chain directly (not via `ufw` commands).

#### Additional systemd sandboxing

Extend the systemd unit with these directives for deeper process isolation:

```ini
[Service]
# ...existing directives (see deployment section below)...

# Prevent writing to the filesystem outside of explicitly allowed paths
ReadOnlyPaths=/
InaccessiblePaths=/home /root /boot /media /mnt

# Restrict dangerous syscalls
SystemCallFilter=@system-service
SystemCallFilter=~@privileged @resources @debug

# Misc hardening
LockPersonality=true
MemoryDenyWriteExecute=true
RestrictRealtime=true
RestrictSUIDSGID=true
RemoveIPC=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectClock=true
```

After editing the unit file:

```bash
sudo systemctl daemon-reload
sudo systemctl restart decoy
```

---

## Docker

### Build the image

```bash
docker build -t decoy .
```

### Run with default config

```bash
docker run -d \
  --name decoy \
  -p 2222:2222 \
  -p 8080:8080 \
  -p 25:25 \
  decoy
```

### Use a custom config

```bash
docker run -d \
  --name decoy \
  -p 2222:2222 \
  -p 8080:8080 \
  -p 4343:4343 \
  -p 25:25 \
  -p 21:21 \
  -p 6379:6379 \
  -p 3306:3306 \
  -p 1433:1433 \
  -v $(pwd)/config/config.yaml:/config/config.yaml:ro \
  decoy
```

### With HTTPS

```bash
docker run -d \
  --name decoy \
  -p 2222:2222 \
  -p 8080:8080 \
  -p 4343:4343 \
  -v $(pwd)/config/config.yaml:/config/config.yaml:ro \
  -v $(pwd)/certs/server.crt:/certs/server.crt:ro \
  -v $(pwd)/certs/server.key:/certs/server.key:ro \
  decoy
```

Ensure the certificate paths in `config.yaml` match the mount targets inside the container (e.g. `/certs/server.crt`).

### View logs

```bash
docker logs -f decoy
```

---

## Pre-compiled deployment

Build a static binary for any Linux target without Docker or Go installed on the host.

### Build

```bash
# amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o decoy .

# arm64 (Raspberry Pi, AWS Graviton)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o decoy .
```

### Ubuntu / Debian

```bash
sudo cp decoy /usr/local/bin/decoy
sudo chmod 755 /usr/local/bin/decoy
sudo mkdir -p /etc/decoy
sudo cp config/config.yaml /etc/decoy/config.yaml

# Dedicated system user — no login shell, no home dir
sudo useradd --system --no-create-home --shell /usr/sbin/nologin decoy

# Allow binding to ports < 1024 without root
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/decoy
```

Create the systemd unit:

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
PrivateTmp=true
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now decoy
sudo journalctl -u decoy -f
```

### Red Hat / Rocky Linux / AlmaLinux

Same steps as above. Additionally open firewall ports:

```bash
sudo firewall-cmd --permanent --add-port=2222/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=4343/tcp
sudo firewall-cmd --reload
```

If SELinux is enforcing:

```bash
sudo semanage fcontext -a -t bin_t '/usr/local/bin/decoy'
sudo restorecon -v /usr/local/bin/decoy
sudo semanage fcontext -a -t etc_t '/etc/decoy(/.*)?'
sudo restorecon -Rv /etc/decoy
```

### Alpine Linux

```bash
cp decoy /usr/local/bin/decoy
chmod 755 /usr/local/bin/decoy
mkdir -p /etc/decoy
cp config/config.yaml /etc/decoy/config.yaml
adduser -S -H -s /sbin/nologin decoy
setcap 'cap_net_bind_service=+ep' /usr/local/bin/decoy
```

OpenRC init script:

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
```

---

## Logstash integration

Decoy forwards structured JSON over UDP syslog. The filter below strips the syslog envelope, parses the JSON, and maps fields to ECS.

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

```ruby
filter {
  grok {
    match => { "message" => "%{SYSLOGTIMESTAMP} %{SYSLOGHOST} decoy: %{GREEDYDATA:json_payload}" }
  }
  json {
    source => "json_payload"
    target => "decoy"
  }
  date {
    match  => [ "[decoy][time]", "ISO8601" ]
    target => "@timestamp"
  }

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

  if [decoy][event] in ["http_request", "http_probe"] {
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

  if [decoy][event] == "tcp_connection" {
    mutate {
      add_tag => [ "honeypot", "tcp" ]
      rename  => {
        "[decoy][remote_ip]" => "[source][address]"
        "[decoy][port]"      => "[destination][port]"
      }
    }
  }

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

### Output — Elasticsearch

```ruby
output {
  elasticsearch {
    hosts    => ["https://your-elasticsearch:9200"]
    index    => "decoy-honeypot-%{+YYYY.MM.dd}"
    user     => "logstash_writer"
    password => "${LOGSTASH_ES_PASSWORD}"
  }
}
```

### Output — Splunk HEC

```bash
logstash-plugin install logstash-output-splunk_hec
```

```ruby
output {
  splunk_hec {
    host       => "splunk.example.com"
    port       => 8088
    token      => "${SPLUNK_HEC_TOKEN}"
    index      => "honeypot"
    sourcetype => "decoy:json"
    ssl        => true
  }
}
```

Useful SPL queries:

```spl
index=honeypot sourcetype="decoy:json" event="ssh_auth_attempt"
| table _time, remote_ip, username, password, client_version, port
| sort -_time

index=honeypot sourcetype="decoy:json"
| eval src_ip=replace(remote_ip, ":.*", "")
| stats count by event, src_ip
| sort -count

index=honeypot sourcetype="decoy:json" event="http_request"
| table _time, remote_ip, method, uri, username, password, ssl
```

### Output — Azure Log Analytics

```bash
logstash-plugin install logstash-output-azure_loganalytics
```

```ruby
output {
  azure_loganalytics {
    workspace_id           => "${AZURE_WORKSPACE_ID}"
    workspace_key          => "${AZURE_WORKSPACE_KEY}"
    custom_log_table_name  => "Decoy"
  }
}
```

Events appear in the `Decoy_CL` table. Example KQL:

```kql
// SSH credential attempts
Decoy_CL
| where event_s == "ssh_auth_attempt"
| project TimeGenerated, source_ip_s, user_name_s, user_password_s, user_agent_original_s
| order by TimeGenerated desc

// Top attacking IPs across all event types
Decoy_CL
| where event_s in ("ssh_auth_attempt", "http_request", "tcp_connection")
| summarize count() by source_ip_s
| order by count_ desc

// HTTP requests by path and method
Decoy_CL
| where event_s in ("http_request", "http_probe")
| summarize count() by url_path_s, http_request_method_s
| order by count_ desc
```

### ECS fields per event

| Event | ECS fields |
|---|---|
| `ssh_auth_attempt` | `source.ip`, `source.port`, `user.name`, `user.password`, `user_agent.original`, `destination.port` |
| `http_request` | `source.ip`, `source.port`, `http.request.method`, `url.path`, `url.query`, `http.request.body.content`, `destination.port` |
| `http_probe` | `source.ip`, `source.port`, `http.request.method`, `url.path`, `destination.port` |
| `tcp_connection` | `source.ip`, `source.port`, `destination.port` |
