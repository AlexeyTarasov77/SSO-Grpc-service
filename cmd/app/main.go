package main

import (
	"flag"
	"log/slog"
	"os"

	"sso.service/internal/app"
	"sso.service/internal/config"
	logHandlers "sso.service/pkg/logger/handlers"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()
	if configPath == "" {
		configPath = config.ResolveConfigPath()
	}
	cfg := config.MustLoad(configPath)
	log := setupLogger(cfg)
	log.Info("Config loaded", "config", cfg)
	app.Run(log, cfg)
}

func setupLogger(cfg *config.Config) *slog.Logger {
	var log *slog.Logger
	switch cfg.Mode {
	case "local":
		opts := logHandlers.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}

		handler := opts.NewPrettyHandler(os.Stdout)
		log = slog.New(handler)

	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		))
	}
	return log
}
