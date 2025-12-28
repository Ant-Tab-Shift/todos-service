package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ant-Tab-Shift/todos-service/internal/domain"
	"github.com/Ant-Tab-Shift/todos-service/internal/infrastructure/storage"
	"github.com/Ant-Tab-Shift/todos-service/internal/transport/http/handlers"
	"github.com/Ant-Tab-Shift/todos-service/internal/transport/http/server"
	"github.com/Ant-Tab-Shift/todos-service/internal/usecases"
)

func main() {
	storage := storage.NewInMemory[domain.TaskSchema]()

	service := usecases.NewTaskService(storage)

	handler := handlers.NewTaskHandler(service)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server := server.New(ctx, ":8080")
	server.RegisterHandlers(handler)

	go func() {
		log.Println("Starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Received signal to start graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
		return
	}

	log.Println("Server stopped gracefully")
}
