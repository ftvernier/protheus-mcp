package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"time"

	_ "github.com/microsoft/go-mssqldb"

	"github.com/ftvernier/protheus-mcp/internal/config"
)

type Health struct {
	Configured        bool    `json:"configured" jsonschema:"whether database monitoring is configured"`
	Connected         bool    `json:"connected" jsonschema:"whether SQL Server is reachable and authenticated"`
	Host              string  `json:"host,omitempty" jsonschema:"SQL Server host"`
	Database          string  `json:"database,omitempty" jsonschema:"monitored Protheus database"`
	LatencyMS         float64 `json:"latency_ms,omitempty" jsonschema:"database ping latency in milliseconds"`
	ActiveRequests    int     `json:"active_requests,omitempty" jsonschema:"number of requests currently executing in this database"`
	BlockedRequests   int     `json:"blocked_requests,omitempty" jsonschema:"currently blocked requests in this database"`
	LongRunning       int     `json:"long_running_queries,omitempty" jsonschema:"requests running for at least ten seconds"`
	MonitoringMessage string  `json:"message,omitempty" jsonschema:"configuration or monitoring information"`
}

func GetHealth(ctx context.Context, cfg config.Config) (Health, error) {
	if !cfg.DatabaseConfigured() {
		return Health{Configured: false, Connected: false, MonitoringMessage: "SQL Server monitoring is not configured. Set DB_HOST, DB_NAME and DB_USER."}, nil
	}
	db, err := open(cfg); if err != nil { return Health{}, err }; defer db.Close()
	pingCtx, cancel := context.WithTimeout(ctx, cfg.QueryTimeout); defer cancel()
	start := time.Now()
	if err := db.PingContext(pingCtx); err != nil { return Health{Configured:true, Connected:false, Host:cfg.DBHost, Database:cfg.DBName}, fmt.Errorf("database ping failed: %w", err) }
	latency := time.Since(start)
	queryCtx, queryCancel := context.WithTimeout(ctx, cfg.QueryTimeout); defer queryCancel()
	const query = `
SELECT
    (SELECT COUNT(*) FROM sys.dm_exec_requests r WHERE r.database_id = DB_ID() AND r.session_id <> @@SPID) AS active_requests,
    (SELECT COUNT(*) FROM sys.dm_exec_requests r WHERE r.database_id = DB_ID() AND r.blocking_session_id > 0) AS blocked_requests,
    (SELECT COUNT(*) FROM sys.dm_exec_requests r WHERE r.database_id = DB_ID() AND r.start_time < DATEADD(SECOND, -10, SYSDATETIME())) AS long_running;`
	var active, blocked, longRunning sql.NullInt64
	if err := db.QueryRowContext(queryCtx, query).Scan(&active, &blocked, &longRunning); err != nil { return Health{}, fmt.Errorf("health query failed; verify monitoring permissions: %w", err) }
	return Health{Configured:true, Connected:true, Host:cfg.DBHost, Database:cfg.DBName, LatencyMS:round(float64(latency.Microseconds())/1000,2), ActiveRequests:int(active.Int64), BlockedRequests:int(blocked.Int64), LongRunning:int(longRunning.Int64)}, nil
}

func open(cfg config.Config) (*sql.DB, error) {
	query := url.Values{}
	query.Add("database", cfg.DBName)
	query.Add("encrypt", cfg.DBEncrypt)
	query.Add("TrustServerCertificate", cfg.DBTrustServerCertificate)
	query.Add("connection timeout", strconv.Itoa(int(cfg.QueryTimeout.Seconds())))
	u := &url.URL{Scheme:"sqlserver", User:url.UserPassword(cfg.DBUser,cfg.DBPassword), Host:fmt.Sprintf("%s:%d",cfg.DBHost,cfg.DBPort), RawQuery:query.Encode()}
	db, err := sql.Open("sqlserver", u.String()); if err != nil { return nil, err }
	db.SetMaxOpenConns(2); db.SetMaxIdleConns(1); db.SetConnMaxLifetime(2*time.Minute)
	return db,nil
}

func round(value float64, places int) float64 { factor:=1.0; for i:=0;i<places;i++ { factor*=10 }; return float64(int(value*factor+0.5))/factor }
