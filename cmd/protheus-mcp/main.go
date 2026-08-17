package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ftvernier/protheus-mcp/internal/config"
	"github.com/ftvernier/protheus-mcp/internal/tools"
)

const version = "0.1.0-alpha"

func main() {
	// STDIO is reserved for MCP JSON-RPC. All application logs MUST go to stderr.
	logger := log.New(os.Stderr, "protheus-mcp: ", log.LstdFlags)

	cfg := config.Load()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "protheus-mcp",
		Version: version,
	}, nil)

	tools.NewRegistry(cfg).Register(server)

	logger.Printf("starting %s using MCP stdio transport", version)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Fatalf("server stopped: %v", err)
	}
}
