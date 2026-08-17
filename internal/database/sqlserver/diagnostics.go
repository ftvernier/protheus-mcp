package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ftvernier/protheus-mcp/internal/config"
)

const maxSQLTextLength = 2000

type QueryInfo struct {
	SessionID int `json:"session_id"`
	DurationSeconds float64 `json:"duration_seconds"`
	Status string `json:"status"`
	Command string `json:"command"`
	Database string `json:"database"`
	Host string `json:"host,omitempty"`
	Program string `json:"program,omitempty"`
	WaitType string `json:"wait_type,omitempty"`
	WaitSeconds float64 `json:"wait_seconds,omitempty"`
	BlockingSessionID int `json:"blocking_session_id,omitempty"`
	SQLText string `json:"sql_text,omitempty"`
	SQLTruncated bool `json:"sql_truncated,omitempty"`
}

type LongRunningQueriesResult struct { Database string `json:"database"`; ThresholdSeconds int `json:"threshold_seconds"`; Count int `json:"count"`; Queries []QueryInfo `json:"queries"` }
type BlockingSession struct { BlockedSessionID int `json:"blocked_session_id"`; BlockingSessionID int `json:"blocking_session_id"`; WaitType string `json:"wait_type,omitempty"`; WaitSeconds float64 `json:"wait_seconds,omitempty"`; Database string `json:"database"`; BlockedHost string `json:"blocked_host,omitempty"`; BlockedProgram string `json:"blocked_program,omitempty"`; BlockerHost string `json:"blocker_host,omitempty"`; BlockerProgram string `json:"blocker_program,omitempty"`; BlockedSQLText string `json:"blocked_sql_text,omitempty"`; SQLTruncated bool `json:"sql_truncated,omitempty"` }
type BlockingSessionsResult struct { Database string `json:"database"`; Count int `json:"count"`; BlockingSessions []BlockingSession `json:"blocking_sessions"` }
type SessionDetails struct { Found bool `json:"found"`; SessionID int `json:"session_id"`; Status string `json:"status,omitempty"`; Database string `json:"database,omitempty"`; Host string `json:"host,omitempty"`; Program string `json:"program,omitempty"`; LoginTime string `json:"login_time,omitempty"`; RequestRunning bool `json:"request_running"`; DurationSeconds float64 `json:"duration_seconds,omitempty"`; CPUms int64 `json:"cpu_ms,omitempty"`; WaitType string `json:"wait_type,omitempty"`; WaitSeconds float64 `json:"wait_seconds,omitempty"`; BlockingSessionID int `json:"blocking_session_id,omitempty"`; SQLText string `json:"sql_text,omitempty"`; SQLTruncated bool `json:"sql_truncated,omitempty"` }

func GetLongRunningQueries(ctx context.Context, cfg config.Config, minDurationSeconds int) (LongRunningQueriesResult, error) {
	if !cfg.DatabaseConfigured() { return LongRunningQueriesResult{}, fmt.Errorf("SQL Server monitoring is not configured; set DB_HOST, DB_NAME and DB_USER") }
	if minDurationSeconds <= 0 { minDurationSeconds = 10 }
	db, err := open(cfg); if err != nil { return LongRunningQueriesResult{}, err }; defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, cfg.QueryTimeout); defer cancel()
	const query = `SELECT r.session_id, CAST(r.total_elapsed_time AS float)/1000.0, r.status, r.command, DB_NAME(r.database_id), s.host_name, s.program_name, r.wait_type, CAST(r.wait_time AS float)/1000.0, r.blocking_session_id, txt.text FROM sys.dm_exec_requests r JOIN sys.dm_exec_sessions s ON s.session_id=r.session_id OUTER APPLY sys.dm_exec_sql_text(r.sql_handle) txt WHERE r.database_id=DB_ID() AND r.session_id<>@@SPID AND r.total_elapsed_time>=(@p1*1000) ORDER BY r.total_elapsed_time DESC;`
	rows, err := db.QueryContext(queryCtx, query, minDurationSeconds); if err != nil { return LongRunningQueriesResult{}, fmt.Errorf("long-running query diagnostic failed; verify monitoring permissions: %w",err) }; defer rows.Close()
	result := LongRunningQueriesResult{Database:cfg.DBName, ThresholdSeconds:minDurationSeconds, Queries:[]QueryInfo{}}
	for rows.Next() { var item QueryInfo; var host,program,waitType,sqlText sql.NullString; var blocking sql.NullInt64; if err:=rows.Scan(&item.SessionID,&item.DurationSeconds,&item.Status,&item.Command,&item.Database,&host,&program,&waitType,&item.WaitSeconds,&blocking,&sqlText); err!=nil{return LongRunningQueriesResult{},err}; item.Host=nullString(host); item.Program=nullString(program); item.WaitType=nullString(waitType); if blocking.Valid&&blocking.Int64>0{item.BlockingSessionID=int(blocking.Int64)}; item.SQLText,item.SQLTruncated=truncateSQL(nullString(sqlText)); result.Queries=append(result.Queries,item) }
	if err:=rows.Err();err!=nil{return LongRunningQueriesResult{},err}; result.Count=len(result.Queries); return result,nil
}

func GetBlockingSessions(ctx context.Context, cfg config.Config) (BlockingSessionsResult,error) {
	if !cfg.DatabaseConfigured(){return BlockingSessionsResult{},fmt.Errorf("SQL Server monitoring is not configured; set DB_HOST, DB_NAME and DB_USER")}; db,err:=open(cfg);if err!=nil{return BlockingSessionsResult{},err};defer db.Close(); queryCtx,cancel:=context.WithTimeout(ctx,cfg.QueryTimeout);defer cancel()
	const query=`SELECT r.session_id,r.blocking_session_id,r.wait_type,CAST(r.wait_time AS float)/1000.0,DB_NAME(r.database_id),blocked.host_name,blocked.program_name,blocker.host_name,blocker.program_name,txt.text FROM sys.dm_exec_requests r JOIN sys.dm_exec_sessions blocked ON blocked.session_id=r.session_id LEFT JOIN sys.dm_exec_sessions blocker ON blocker.session_id=r.blocking_session_id OUTER APPLY sys.dm_exec_sql_text(r.sql_handle) txt WHERE r.database_id=DB_ID() AND r.blocking_session_id>0 ORDER BY r.wait_time DESC;`
	rows,err:=db.QueryContext(queryCtx,query);if err!=nil{return BlockingSessionsResult{},fmt.Errorf("blocking diagnostic failed; verify monitoring permissions: %w",err)};defer rows.Close();result:=BlockingSessionsResult{Database:cfg.DBName,BlockingSessions:[]BlockingSession{}}
	for rows.Next(){var item BlockingSession;var wt,bh,bp,krh,krp,st sql.NullString;if err:=rows.Scan(&item.BlockedSessionID,&item.BlockingSessionID,&wt,&item.WaitSeconds,&item.Database,&bh,&bp,&krh,&krp,&st);err!=nil{return BlockingSessionsResult{},err};item.WaitType=nullString(wt);item.BlockedHost=nullString(bh);item.BlockedProgram=nullString(bp);item.BlockerHost=nullString(krh);item.BlockerProgram=nullString(krp);item.BlockedSQLText,item.SQLTruncated=truncateSQL(nullString(st));result.BlockingSessions=append(result.BlockingSessions,item)};if err:=rows.Err();err!=nil{return BlockingSessionsResult{},err};result.Count=len(result.BlockingSessions);return result,nil
}

func GetSessionDetails(ctx context.Context,cfg config.Config,sessionID int)(SessionDetails,error){
	if !cfg.DatabaseConfigured(){return SessionDetails{},fmt.Errorf("SQL Server monitoring is not configured; set DB_HOST, DB_NAME and DB_USER")};if sessionID<=0{return SessionDetails{},fmt.Errorf("session_id must be greater than zero")};db,err:=open(cfg);if err!=nil{return SessionDetails{},err};defer db.Close();queryCtx,cancel:=context.WithTimeout(ctx,cfg.QueryTimeout);defer cancel()
	const query=`SELECT s.session_id,COALESCE(r.status,s.status),DB_NAME(r.database_id),s.host_name,s.program_name,s.login_time,CASE WHEN r.session_id IS NULL THEN CAST(0 AS bit) ELSE CAST(1 AS bit) END,CAST(COALESCE(r.total_elapsed_time,0) AS float)/1000.0,COALESCE(r.cpu_time,0),r.wait_type,CAST(COALESCE(r.wait_time,0) AS float)/1000.0,COALESCE(r.blocking_session_id,0),txt.text FROM sys.dm_exec_sessions s LEFT JOIN sys.dm_exec_requests r ON r.session_id=s.session_id LEFT JOIN sys.dm_exec_connections c ON c.session_id=s.session_id OUTER APPLY sys.dm_exec_sql_text(COALESCE(r.sql_handle,c.most_recent_sql_handle)) txt WHERE s.session_id=@p1 AND s.is_user_process=1;`
	var result SessionDetails;var status,databaseName,host,program,waitType,sqlText sql.NullString;var loginTime time.Time;var requestRunning bool;var blocking int;err=db.QueryRowContext(queryCtx,query,sessionID).Scan(&result.SessionID,&status,&databaseName,&host,&program,&loginTime,&requestRunning,&result.DurationSeconds,&result.CPUms,&waitType,&result.WaitSeconds,&blocking,&sqlText);if err==sql.ErrNoRows{return SessionDetails{Found:false,SessionID:sessionID},nil};if err!=nil{return SessionDetails{},fmt.Errorf("session diagnostic failed; verify monitoring permissions: %w",err)};result.Found=true;result.Status=nullString(status);result.Database=nullString(databaseName);result.Host=nullString(host);result.Program=nullString(program);result.LoginTime=loginTime.UTC().Format(time.RFC3339);result.RequestRunning=requestRunning;result.WaitType=nullString(waitType);if blocking>0{result.BlockingSessionID=blocking};result.SQLText,result.SQLTruncated=truncateSQL(nullString(sqlText));return result,nil
}

func nullString(value sql.NullString)string{if !value.Valid{return ""};return strings.TrimSpace(value.String)}
func truncateSQL(value string)(string,bool){value=strings.TrimSpace(value);if len(value)<=maxSQLTextLength{return value,false};return value[:maxSQLTextLength],true}
