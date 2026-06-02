package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"

    "masker/masker"
    "github.com/urfave/cli/v2"
)

var (
    logLevel string
    input    string
    output   string
)

func main() {
    app := &cli.App{
        Name:    "masker",
        Usage:   "Маскирует цифры в текстовых файлах",
        Version: "2.0.0",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:        "input",
                Aliases:     []string{"i"},
                Value:       "input.txt",
                Usage:       "Путь к входному файлу",
                Destination: &input,
            },
            &cli.StringFlag{
                Name:        "output",
                Aliases:     []string{"o"},
                Value:       "output.txt",
                Usage:       "Путь к выходному файлу",
                Destination: &output,
            },
            &cli.StringFlag{
                Name:        "log-level",
                Aliases:     []string{"l"},
                Value:       "info",
                Usage:       "Уровень логирования (debug, info, warn, error)",
                Destination: &logLevel,
            },
        },
        Action: run,
    }

    if err := app.Run(os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
        os.Exit(1)
    }
}

func run(c *cli.Context) error {
    var level slog.Level
    switch logLevel {
    case "debug":
        level = slog.LevelDebug
    case "info":
        level = slog.LevelInfo
    case "warn":
        level = slog.LevelWarn
    case "error":
        level = slog.LevelError
    default:
        level = slog.LevelInfo
    }

    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: level,
    }))

    logger.Info("запуск masker", "input", input, "output", output, "log_level", logLevel)

    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // ✅ ИСПРАВЛЕНО: передаём logger
    producer := masker.NewFileProducer(input, logger)
    presenter := masker.NewFilePresenter(output, logger)
    maskerImpl := masker.DigitsMasker{}
    service := masker.NewService(producer, presenter, maskerImpl, logger)

    done := make(chan error, 1)

    go func() {
        done <- service.Run(ctx)
        close(done)
    }()

    select {
    case err := <-done:
        if err != nil {
            logger.Error("обработка завершена с ошибкой", "error", err)
            return err
        }
        logger.Info("обработка успешно завершена")
        return nil

    case <-ctx.Done():
        logger.Warn("получен сигнал завершения, ожидание остановки...")

        select {
        case err := <-done:
            if err != nil {
                logger.Error("обработка завершена с ошибкой после сигнала", "error", err)
                return err
            }
            logger.Info("обработка корректно остановлена")
            return nil
        case <-time.After(5 * time.Second):
            logger.Error("таймаут ожидания")
            return fmt.Errorf("graceful shutdown timeout")
        }
    }
}