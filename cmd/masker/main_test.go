package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRunMasksFile(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.txt")
	out := filepath.Join(dir, "out.txt")
	require.NoError(t, os.WriteFile(in, []byte("Привет 123\nГод 2025\n"), 0o600))

	err := run(context.Background(), config{input: in, output: out}, discardLogger())
	require.NoError(t, err)

	got, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "Привет ***\nГод ****\n", string(got))
}

func TestRunKeepsPreviousOutputOnCancel(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.txt")
	out := filepath.Join(dir, "out.txt")
	require.NoError(t, os.WriteFile(in, []byte("1\n2\n"), 0o600))
	require.NoError(t, os.WriteFile(out, []byte("прежний результат"), 0o600))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, config{input: in, output: out}, discardLogger())
	assert.ErrorIs(t, err, context.Canceled)

	got, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "прежний результат", string(got))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 2, "лишние файлы: %v", names(entries))
}

func TestRunFailsOnMissingInput(t *testing.T) {
	err := run(context.Background(),
		config{input: filepath.Join(t.TempDir(), "нет.txt"), output: "-"},
		discardLogger())
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestParseFlags(t *testing.T) {
	t.Run("значения по умолчанию", func(t *testing.T) {
		cfg, err := parseFlags("masker", nil, io.Discard)
		require.NoError(t, err)
		assert.Equal(t, "-", cfg.input)
		assert.Equal(t, "-", cfg.output)
		assert.Equal(t, slog.LevelInfo, cfg.level)
	})

	t.Run("явные значения", func(t *testing.T) {
		cfg, err := parseFlags("masker",
			[]string{"-i", "a.txt", "-o", "b.txt", "-log-level", "debug"}, io.Discard)
		require.NoError(t, err)
		assert.Equal(t, "a.txt", cfg.input)
		assert.Equal(t, "b.txt", cfg.output)
		assert.Equal(t, slog.LevelDebug, cfg.level)
	})

	t.Run("неизвестный уровень — ошибка", func(t *testing.T) {
		_, err := parseFlags("masker", []string{"-log-level", "verbose"}, io.Discard)
		assert.Error(t, err)
	})
}

func names(entries []os.DirEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}
