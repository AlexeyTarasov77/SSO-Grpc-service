package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sso.service/internal/app"
	"sso.service/internal/config"
	"sso.service/internal/lib/logger/handlers/slogpretty"
)

func main() {
	var configPath string
	flag.StringVar(&configPath ,"config", "", "path to config file")
	flag.Parse()
	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}
	cfg := config.Load(configPath)
	log := setupLogger(cfg)
	log.Info("Config loaded", "config", cfg)
	storagePath := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	application := app.New(log, cfg, storagePath)
	go application.GRPCApp.Run()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	application.GRPCApp.Stop()
}

func setupLogger(cfg *config.Config) *slog.Logger {
	var log *slog.Logger
	switch cfg.Mode {
	case localMode:
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}
	
		handler := opts.NewPrettyHandler(os.Stdout)
		log = slog.New(handler)

	case prodMode:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		))
	}
	return log
}
