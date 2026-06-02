package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"

	"masker/masker"
)

func main() {
	app := &cli.App{
		Name:  "masker",
		Usage: "Читает текстовый файл, заменяет все цифры на * и сохраняет результат",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Value:   "input.txt",
				Usage:   "Путь к входному файлу",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "output.txt",
				Usage:   "Путь к выходному файлу",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "info",
				Usage: "Уровень логирования: debug, info, warn, error",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	// --- настройка уровня логирования ---
	level, err := parseLogLevel(c.String("log-level"))
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	logger.Debug("cli: application starting",
		slog.String("input", c.String("input")),
		slog.String("output", c.String("output")),
		slog.String("log_level", c.String("log-level")),
	)

	// --- контекст с отменой по сигналу (Ctrl+C / SIGTERM) ---
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("cli: context created, listening for interrupt signals (Ctrl+C)")

	// --- сборка сервиса ---
	inputPath := c.String("input")
	outputPath := c.String("output")

	prod := masker.NewFileProducer(inputPath)
	pres := masker.NewFilePresenter(outputPath)
	maskerImpl := masker.DigitsMasker{}
	svc := masker.NewService(prod, pres, maskerImpl, logger)

	logger.Debug("cli: service assembled, running")

	// --- запуск в отдельной горутине с чётким жизненным циклом ---
	type runResult struct {
		err error
	}
	done := make(chan runResult, 1)

	go func() {
		logger.Debug("goroutine[main-worker]: started")
		err := svc.Run(ctx)
		logger.Debug("goroutine[main-worker]: finished", slog.Any("error", err))
		done <- runResult{err: err}
	}()

	// --- ожидаем завершения: либо сервис отработал, либо пришёл сигнал ---
	select {
	case res := <-done:
		if res.err != nil {
			logger.Error("cli: service finished with error", slog.Any("error", res.err))
			return res.err
		}
		logger.Info("cli: service finished successfully")
		return nil

	case <-ctx.Done():
		logger.Warn("cli: interrupt received, waiting for goroutine to finish")
		// ждём завершения горутины после отмены контекста
		res := <-done
		if res.err != nil && res.err != context.Canceled {
			logger.Error("cli: service finished with error after interrupt", slog.Any("error", res.err))
			return res.err
		}
		logger.Info("cli: graceful shutdown complete")
		return nil
	}
}

// parseLogLevel преобразует строку в slog.Level
func parseLogLevel(s string) (slog.Level, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level %q, use: debug, info, warn, error", s)
	}
}
