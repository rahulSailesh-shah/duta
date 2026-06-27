package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rahulSailesh-shah/duta/internal/config"
	"github.com/rahulSailesh-shah/duta/internal/database"
	"github.com/rahulSailesh-shah/duta/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.New(ctx, database.Options{
		Region:    cfg.AwsRegion,
		AccessKey: cfg.AwsAccessKey,
		SecretKey: cfg.AwsSecretKey,
	})
	if err != nil {
		return err
	}

	srv := server.New(cfg, db)
	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      srv,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	go func() {
		log.Printf("Server is running on %s\n", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("ListenAndServe error: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(shutDownCtx)
}
