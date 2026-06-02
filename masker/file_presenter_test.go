package masker

import (
    "context"
    "log/slog"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFilePresenter_Present_Success(t *testing.T) {
    tmpFile, err := os.CreateTemp("", "test_output_*.txt")
    require.NoError(t, err)
    defer os.Remove(tmpFile.Name())
    tmpFile.Close()

    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    presenter := NewFilePresenter(tmpFile.Name(), logger)
    ctx := context.Background()

    data := []string{"hello", "world", "test"}
    err = presenter.Present(ctx, data)

    assert.NoError(t, err)

    // Проверяем содержимое
    content, err := os.ReadFile(tmpFile.Name())
    require.NoError(t, err)
    assert.Equal(t, "hello\nworld\ntest\n", string(content))
}

func TestFilePresenter_Present_OverwritesFile(t *testing.T) {
    tmpFile, err := os.CreateTemp("", "test_overwrite_*.txt")
    require.NoError(t, err)
    defer os.Remove(tmpFile.Name())

    // Записываем старые данные
    _, err = tmpFile.WriteString("old data\n")
    require.NoError(t, err)
    tmpFile.Close()

    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    presenter := NewFilePresenter(tmpFile.Name(), logger)
    ctx := context.Background()

    data := []string{"new", "data"}
    err = presenter.Present(ctx, data)

    assert.NoError(t, err)

    // Проверяем, что старые данные перезаписаны
    content, err := os.ReadFile(tmpFile.Name())
    require.NoError(t, err)
    assert.Equal(t, "new\ndata\n", string(content))
}

func TestFilePresenter_Present_EmptyData(t *testing.T) {
    tmpFile, err := os.CreateTemp("", "test_empty_*.txt")
    require.NoError(t, err)
    defer os.Remove(tmpFile.Name())
    tmpFile.Close()

    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    presenter := NewFilePresenter(tmpFile.Name(), logger)
    ctx := context.Background()

    err = presenter.Present(ctx, []string{})

    assert.NoError(t, err)

    // Проверяем пустой файл
    content, err := os.ReadFile(tmpFile.Name())
    require.NoError(t, err)
    assert.Equal(t, "", string(content))
}