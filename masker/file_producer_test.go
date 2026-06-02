package masker

import (
    "context"
    "log/slog"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFileProducer_Produce_Success(t *testing.T) {
    // Создаём временный файл
    tmpFile, err := os.CreateTemp("", "test_input_*.txt")
    require.NoError(t, err)
    defer os.Remove(tmpFile.Name())

    content := "line1\nline2\nline3\n"
    _, err = tmpFile.WriteString(content)
    require.NoError(t, err)
    tmpFile.Close()

    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    producer := NewFileProducer(tmpFile.Name(), logger)
    ctx := context.Background()

    result, err := producer.Produce(ctx)

    assert.NoError(t, err)
    assert.Equal(t, []string{"line1", "line2", "line3"}, result)
}

func TestFileProducer_Produce_FileNotFound(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    producer := NewFileProducer("nonexistent_file_12345.txt", logger)
    ctx := context.Background()

    result, err := producer.Produce(ctx)

    assert.Error(t, err)
    assert.Nil(t, result)
}

func TestFileProducer_Produce_EmptyFile(t *testing.T) {
    tmpFile, err := os.CreateTemp("", "test_empty_*.txt")
    require.NoError(t, err)
    defer os.Remove(tmpFile.Name())
    tmpFile.Close()

    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    producer := NewFileProducer(tmpFile.Name(), logger)
    ctx := context.Background()

    result, err := producer.Produce(ctx)

    assert.NoError(t, err)
    assert.Empty(t, result)
}