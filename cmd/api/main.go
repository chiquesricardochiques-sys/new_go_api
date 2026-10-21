package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ricardo-platform/salao-usuarios-api/internal/config"
	"github.com/ricardo-platform/salao-usuarios-api/internal/database"
	httpapi "github.com/ricardo-platform/salao-usuarios-api/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erro na configuração: ", err)
	}

	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal("Erro no banco de dados: ", err)
	}
	defer db.Close()

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(db),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stopCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		log.Printf("API de usuários iniciada em http://localhost:%s", cfg.Port)
		log.Printf("Verificação: http://localhost:%s/health", cfg.Port)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Printf("Erro no servidor: %v", err)
			stop()
		}
	}()

	<-stopCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Erro ao encerrar o servidor: %v", err)
	}
}
