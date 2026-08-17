package tools

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ftvernier/protheus-mcp/internal/config"
	"github.com/ftvernier/protheus-mcp/internal/database/sqlserver"
	"github.com/ftvernier/protheus-mcp/internal/protheus"
	syshealth "github.com/ftvernier/protheus-mcp/internal/system"
)

type EmptyInput struct{}
type LongRunningQueriesInput struct { MinDurationSeconds int `json:"min_duration_seconds,omitempty" jsonschema:"minimum elapsed execution time in seconds; defaults to 10"` }
type SessionDetailsInput struct { SessionID int `json:"session_id" jsonschema:"SQL Server session ID to inspect"` }
type SystemHealthOutput struct { Health syshealth.Health `json:"health"` }
type ProtheusProcessesOutput struct { Result protheus.ProcessList `json:"result"` }
type DatabaseHealthOutput struct { Health sqlserver.Health `json:"health"` }
type LongRunningQueriesOutput struct { Result sqlserver.LongRunningQueriesResult `json:"result"` }
type BlockingSessionsOutput struct { Result sqlserver.BlockingSessionsResult `json:"result"` }
type SessionDetailsOutput struct { Result sqlserver.SessionDetails `json:"result"` }
type Registry struct { cfg config.Config }
func NewRegistry(cfg config.Config)*Registry{return &Registry{cfg:cfg}}
func(r *Registry)Register(server *mcp.Server){
	mcp.AddTool(server,&mcp.Tool{Name:"get_system_health",Description:"Read-only. Returns current CPU, memory, uptime and local disk usage from the host running the Protheus MCP server."},r.getSystemHealth)
	mcp.AddTool(server,&mcp.Tool{Name:"get_protheus_processes",Description:"Read-only. Finds Protheus AppServer processes and returns PID, CPU, memory and process uptime."},r.getProtheusProcesses)
	mcp.AddTool(server,&mcp.Tool{Name:"get_database_health",Description:"Read-only. Checks SQL Server connectivity and reports active requests, blocked requests and long-running requests for the configured Protheus database."},r.getDatabaseHealth)
	mcp.AddTool(server,&mcp.Tool{Name:"get_long_running_queries",Description:"Read-only. Lists currently executing SQL Server requests in the configured Protheus database whose elapsed time exceeds a configurable threshold."},r.getLongRunningQueries)
	mcp.AddTool(server,&mcp.Tool{Name:"get_blocking_sessions",Description:"Read-only. Lists current SQL Server blocking relationships in the configured Protheus database."},r.getBlockingSessions)
	mcp.AddTool(server,&mcp.Tool{Name:"get_session_details",Description:"Read-only. Inspects one SQL Server user session by session ID and returns status, client host/program, current wait/blocker, CPU, elapsed time and SQL text."},r.getSessionDetails)
}
func(r *Registry)getSystemHealth(ctx context.Context,_ *mcp.CallToolRequest,_ EmptyInput)(*mcp.CallToolResult,SystemHealthOutput,error){h,e:=syshealth.GetHealth(ctx);if e!=nil{return nil,SystemHealthOutput{},e};return nil,SystemHealthOutput{Health:h},nil}
func(r *Registry)getProtheusProcesses(ctx context.Context,_ *mcp.CallToolRequest,_ EmptyInput)(*mcp.CallToolResult,ProtheusProcessesOutput,error){x,e:=protheus.FindProcesses(ctx,r.cfg.ProtheusProcess);if e!=nil{return nil,ProtheusProcessesOutput{},e};return nil,ProtheusProcessesOutput{Result:x},nil}
func(r *Registry)getDatabaseHealth(ctx context.Context,_ *mcp.CallToolRequest,_ EmptyInput)(*mcp.CallToolResult,DatabaseHealthOutput,error){h,e:=sqlserver.GetHealth(ctx,r.cfg);if e!=nil{return nil,DatabaseHealthOutput{},e};return nil,DatabaseHealthOutput{Health:h},nil}
func(r *Registry)getLongRunningQueries(ctx context.Context,_ *mcp.CallToolRequest,input LongRunningQueriesInput)(*mcp.CallToolResult,LongRunningQueriesOutput,error){x,e:=sqlserver.GetLongRunningQueries(ctx,r.cfg,input.MinDurationSeconds);if e!=nil{return nil,LongRunningQueriesOutput{},e};return nil,LongRunningQueriesOutput{Result:x},nil}
func(r *Registry)getBlockingSessions(ctx context.Context,_ *mcp.CallToolRequest,_ EmptyInput)(*mcp.CallToolResult,BlockingSessionsOutput,error){x,e:=sqlserver.GetBlockingSessions(ctx,r.cfg);if e!=nil{return nil,BlockingSessionsOutput{},e};return nil,BlockingSessionsOutput{Result:x},nil}
func(r *Registry)getSessionDetails(ctx context.Context,_ *mcp.CallToolRequest,input SessionDetailsInput)(*mcp.CallToolResult,SessionDetailsOutput,error){x,e:=sqlserver.GetSessionDetails(ctx,r.cfg,input.SessionID);if e!=nil{return nil,SessionDetailsOutput{},e};return nil,SessionDetailsOutput{Result:x},nil}
