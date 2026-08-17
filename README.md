# Protheus MCP

> **AI-assisted troubleshooting for TOTVS Protheus.**

**Protheus MCP** is an open-source MCP (Model Context Protocol) server written in **Go** that gives AI assistants real-time, **read-only** visibility into TOTVS Protheus environments.

Instead of asking an AI a generic question such as *“what can make Protheus slow?”*, Protheus MCP enables a much more useful workflow:

> **“Investigate why my Protheus environment is slow right now.”**

The AI can collect evidence from the Windows host, Protheus AppServer processes, and SQL Server, then correlate those signals to assist with troubleshooting.

![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)
![MCP](https://img.shields.io/badge/MCP-Server-5A45FF)
![Windows](https://img.shields.io/badge/Windows-primary-0078D4?logo=windows)
![SQL Server](https://img.shields.io/badge/SQL%20Server-supported-CC2927?logo=microsoftsqlserver&logoColor=white)
![Read Only](https://img.shields.io/badge/security-read--only-success)
![License](https://img.shields.io/badge/license-MIT-blue)

## Why Protheus MCP?

Troubleshooting ERP performance often means jumping between Task Manager, Windows services, SQL Server DMVs, AppServer processes, logs, and monitoring tools.

Protheus MCP starts building a single bridge between that operational context and an MCP-compatible AI assistant.

```text
                    MCP-compatible AI
                           │
                           │ MCP / stdio
                           ▼
                   ┌────────────────┐
                   │  Protheus MCP  │
                   │       Go       │
                   └───────┬────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
           Windows      Protheus     SQL Server
              │         AppServer        │
              └────────────┼─────────────┘
                           ▼
                    Operational data
                           │
                           ▼
                  AI-assisted diagnosis
```

## v0.1.0-alpha capabilities

The first public alpha exposes six read-only MCP tools:

| Tool | Purpose |
| --- | --- |
| `get_system_health` | CPU, memory, OS uptime, and local disk usage |
| `get_protheus_processes` | Finds AppServer processes and returns PID, CPU, memory, and uptime |
| `get_database_health` | SQL Server connectivity, latency, active requests, blocked requests, and long-running requests |
| `get_long_running_queries` | Lists currently executing requests above a configurable duration threshold |
| `get_blocking_sessions` | Shows current SQL Server blocking relationships and blocking/blocked session IDs |
| `get_session_details` | Deep-dives into a SQL Server session, including waits, blocker, CPU, and SQL text |

### Example investigation

```text
User: My Protheus environment is slow. Investigate it.

AI
 ├─ get_system_health()
 │    └─ CPU: normal / Memory: normal / Disk: normal
 ├─ get_database_health()
 │    └─ blocked_requests: 4
 ├─ get_blocking_sessions()
 │    └─ session 84 is blocking sessions 117, 132 and 141
 └─ get_session_details(session_id=84)
      └─ TOTVS Application Server / long-running SQL request

AI: The strongest current indicator is database contention. Session 84,
    originated by TOTVS Application Server, is blocking three requests.
    Host resources are currently within normal ranges.
```

The MCP provides the **evidence**. The AI assistant provides the reasoning and explanation.

## Security first: read-only by design

The public alpha is intentionally diagnostic-only. It does **not** execute arbitrary SQL, run `KILL`, terminate AppServer processes, restart services, modify `appserver.ini`, modify ERP/database data, or return credentials through MCP tools.

SQL text returned by diagnostic tools is truncated to reduce unnecessary exposure and context usage.

> Review the privacy/data-handling policy of the AI client you connect to the MCP. Operational metadata and SQL text returned by tools may become part of that client's model context.

## Requirements

### Build from source
- Go 1.23+
- Windows is the primary target for the first alpha
- SQL Server is optional; system/process tools work without database configuration

### Compiled release
No Go runtime is required. Download the Windows executable from a project release and configure it in your MCP client.

## Build on Windows

```powershell
go mod tidy
go test ./...
go build -o protheus-mcp.exe ./cmd/protheus-mcp
```

Or:

```powershell
.\scripts\build-windows.ps1
```

## Configuration

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `PROTHEUS_PROCESS` | No | `appserver` | Process name/pattern used to find Protheus AppServers |
| `DB_HOST` | SQL tools | — | SQL Server host |
| `DB_PORT` | No | `1433` | SQL Server TCP port |
| `DB_NAME` | SQL tools | — | Protheus database name |
| `DB_USER` | SQL tools | — | Monitoring login |
| `DB_PASSWORD` | SQL tools | — | Monitoring password |
| `DB_ENCRYPT` | No | `true` | Enables encrypted SQL Server connection |
| `DB_TRUST_SERVER_CERTIFICATE` | No | `false` | Trusts the presented SQL Server certificate without CA validation |
| `QUERY_TIMEOUT_SECONDS` | No | `5` | Timeout for monitoring queries |

```cmd
set PROTHEUS_PROCESS=appserver
set DB_HOST=localhost
set DB_PORT=1433
set DB_NAME=PROTHEUS
set DB_USER=protheus_monitor
set DB_PASSWORD=CHANGE_ME
set DB_ENCRYPT=true
set DB_TRUST_SERVER_CERTIFICATE=false
```

For SQL Server environments using an internal/self-signed certificate, `DB_TRUST_SERVER_CERTIFICATE=true` may be necessary. Prefer a certificate trusted by the client whenever possible.

### Least privilege

Use a dedicated **monitoring-only login**. Do not use `sa`, the Protheus application credential, or an account with unnecessary write privileges. The alpha performs only diagnostic reads/DMV queries.

## Testing with MCP Inspector

```cmd
npx @modelcontextprotocol/inspector ^
  -e DB_HOST=localhost ^
  -e DB_PORT=1433 ^
  -e DB_NAME=PROTHEUS ^
  -e DB_USER=protheus_monitor ^
  -e DB_PASSWORD=CHANGE_ME ^
  -e DB_ENCRYPT=true ^
  -e DB_TRUST_SERVER_CERTIFICATE=false ^
  .\protheus-mcp.exe
```

After connecting, the six tools should be visible in the Inspector.

## MCP client configuration

Protheus MCP currently uses MCP over `stdio`.

```json
{
  "mcpServers": {
    "protheus": {
      "command": "C:\\tools\\protheus-mcp\\protheus-mcp.exe",
      "env": {
        "PROTHEUS_PROCESS": "appserver",
        "DB_HOST": "localhost",
        "DB_PORT": "1433",
        "DB_NAME": "PROTHEUS",
        "DB_USER": "protheus_monitor",
        "DB_PASSWORD": "CHANGE_ME",
        "DB_ENCRYPT": "true",
        "DB_TRUST_SERVER_CERTIFICATE": "false"
      }
    }
  }
}
```

Because the server uses `stdio`, stdout is reserved for MCP JSON-RPC traffic. Application logs are written to stderr.

## Suggested prompts

```text
Analyze the current health of my Protheus environment. Use the available
read-only Protheus MCP tools and explain the evidence before recommending
what I should investigate.
```

```text
My Protheus environment is slow right now. Determine whether the strongest
indicator is the Windows host, AppServer processes, or SQL Server.
```

```text
Check for SQL Server blocking. If you find a blocker, inspect the blocking
session and explain what it is doing. Do not perform any corrective action.
```

## Roadmap

### v0.1.x
- [x] MCP server written in Go
- [x] Windows host health
- [x] Protheus AppServer process discovery
- [x] SQL Server health
- [x] Long-running query diagnostics
- [x] Blocking-session diagnostics
- [x] Session deep-dive
- [x] Configurable SQL Server TLS certificate trust
- [ ] Expanded automated test coverage
- [ ] Friendlier configuration/bootstrap experience

### v0.2
- [ ] Windows service discovery
- [ ] `appserver.ini` parser and environment context
- [ ] DBAccess / License Server / REST connectivity checks
- [ ] SQL Server wait diagnostics

### Future
- [ ] PostgreSQL provider
- [ ] AppServer log diagnostics
- [ ] Multi-host environment correlation
- [ ] Additional Protheus operational context

## Contributing

Issues, feedback, test results, and pull requests are welcome. Early feedback from different Protheus topologies is especially valuable while the project is still in alpha.

See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## Disclaimer

This is an independent open-source project and is **not affiliated with or endorsed by TOTVS S.A.** TOTVS and Protheus are trademarks of their respective owners.

The project is experimental software. Validate diagnostics in controlled environments before relying on them in production.

## License

MIT — see [`LICENSE`](LICENSE).
