package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ernestkhasanzhinov/mcp-simulator/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	numServers := flag.Int("servers", 1000, "number of virtual MCP servers")
	minTools := flag.Int("min-tools", 3, "minimum tools per server")
	maxTools := flag.Int("max-tools", 10, "maximum tools per server")
	minMutationInterval := flag.Duration("min-mutation-interval", 5*time.Second, "minimum interval between tool mutations")
	maxMutationInterval := flag.Duration("max-mutation-interval", 60*time.Second, "maximum interval between tool mutations")
	flag.Parse()

	cfg := server.Config{
		NumServers:          *numServers,
		MinTools:            *minTools,
		MaxTools:            *maxTools,
		MinMutationInterval: *minMutationInterval,
		MaxMutationInterval: *maxMutationInterval,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registry := server.NewRegistry(cfg)
	registry.StartMutations(ctx)

	mux := http.NewServeMux()
	mux.Handle("/server/", registry)

	addr := fmt.Sprintf(":%d", *port)
	httpServer := &http.Server{Addr: addr, Handler: mux}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		httpServer.Shutdown(shutdownCtx)
	}()

	log.Printf("Starting %d virtual MCP servers on %s", cfg.NumServers, addr)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
