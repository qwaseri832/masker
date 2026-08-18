package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/qwaseri832/masker/internal/mask"
	"github.com/qwaseri832/masker/internal/pipeline"
)

const exitInterrupted = 130

func main() {
	cfg, err := parseFlags(filepath.Base(os.Args[0]), os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		os.Exit(2)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.level}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, logger); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Warn("остановлено по сигналу")
			os.Exit(exitInterrupted)
		}
		logger.Error("не удалось обработать файл", "error", err)
		os.Exit(1)
	}
}

type config struct {
	input  string
	output string
	level  slog.Level
}

func parseFlags(prog string, args []string, stderr io.Writer) (config, error) {
	fs := flag.NewFlagSet(prog, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "masker заменяет цифры в тексте на «%c».\n\n", mask.Rune)
		fmt.Fprintf(stderr, "Использование:\n  %s [флаги]\n\nФлаги:\n", prog)
		fs.PrintDefaults()
	}

	var (
		input    = fs.String("i", "-", "входной файл (\"-\" — stdin)")
		output   = fs.String("o", "-", "выходной файл (\"-\" — stdout)")
		levelStr = fs.String("log-level", "info", "уровень логирования: debug, info, warn, error")
	)

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(*levelStr)); err != nil {
		fmt.Fprintf(stderr, "неизвестный уровень логирования %q\n", *levelStr)
		return config{}, fmt.Errorf("уровень логирования: %w", err)
	}

	return config{input: *input, output: *output, level: level}, nil
}

func run(ctx context.Context, cfg config, logger *slog.Logger) error {
	logger.Debug("старт", "input", cfg.input, "output", cfg.output)

	src, closeSrc, err := openInput(cfg.input)
	if err != nil {
		return err
	}
	defer closeSrc()

	out, err := openOutput(cfg.output)
	if err != nil {
		return err
	}

	defer out.discard()

	started := time.Now()
	st, err := pipeline.Run(ctx, out.w, src, mask.Digits{})
	if err != nil {
		return err
	}
	if err := out.commit(); err != nil {
		return err
	}

	logger.Info("готово",
		"lines", st.Lines,
		"duration", time.Since(started).Round(time.Millisecond),
	)
	return nil
}

func openInput(path string) (io.Reader, func(), error) {
	if path == "-" {
		return os.Stdin, func() {}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("открыть входной файл: %w", err)
	}
	return f, func() { _ = f.Close() }, nil
}

type output struct {
	w        io.Writer
	tmp      *os.File
	path     string
	finished bool
}

func openOutput(path string) (*output, error) {
	if path == "-" {
		return &output{w: os.Stdout}, nil
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return nil, fmt.Errorf("создать временный файл: %w", err)
	}

	out := &output{w: tmp, tmp: tmp, path: path}
	if err := tmp.Chmod(outputMode(path)); err != nil {
		out.discard()
		return nil, fmt.Errorf("права на временный файл: %w", err)
	}

	return out, nil
}

func outputMode(path string) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return 0o644
	}
	return info.Mode().Perm()
}

func (o *output) commit() error {
	if o.tmp == nil {
		o.finished = true
		return nil
	}
	if err := o.tmp.Sync(); err != nil {
		return fmt.Errorf("сбросить на диск: %w", err)
	}
	if err := o.tmp.Close(); err != nil {
		return fmt.Errorf("закрыть временный файл: %w", err)
	}
	if err := os.Rename(o.tmp.Name(), o.path); err != nil {
		return fmt.Errorf("переименовать временный файл: %w", err)
	}
	o.finished = true
	return nil
}

func (o *output) discard() {
	if o.tmp == nil || o.finished {
		return
	}
	_ = o.tmp.Close()
	_ = os.Remove(o.tmp.Name())
}
