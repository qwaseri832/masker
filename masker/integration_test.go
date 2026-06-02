package masker

import (
    "context"
    "log/slog"
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestIntegration_EndToEnd(t *testing.T) {
    // Создаём временную директорию
    tmpDir := t.TempDir()
    
    inputPath := filepath.Join(tmpDir, "input.txt")
    outputPath := filepath.Join(tmpDir, "output.txt")
    
    // Создаём входной файл
    inputContent := "abc123\ndef456\nhello\nworld789\n"
    err := os.WriteFile(inputPath, []byte(inputContent), 0644)
    require.NoError(t, err)
    
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    producer := NewFileProducer(inputPath, logger)
    presenter := NewFilePresenter(outputPath, logger)
    maskerImpl := DigitsMasker{}
    service := NewService(producer, presenter, maskerImpl, logger)
    
    ctx := context.Background()
    err = service.Run(ctx)
    
    assert.NoError(t, err)
    
    // Проверяем выходной файл
    outputContent, err := os.ReadFile(outputPath)
    require.NoError(t, err)
    
    expected := "abc***\ndef***\nhello\nworld***\n"
    assert.Equal(t, expected, string(outputContent))
}

func TestIntegration_EmptyInputFile(t *testing.T) {
    tmpDir := t.TempDir()
    
    inputPath := filepath.Join(tmpDir, "empty.txt")
    outputPath := filepath.Join(tmpDir, "output.txt")
    
    // Создаём пустой входной файл
    err := os.WriteFile(inputPath, []byte(""), 0644)
    require.NoError(t, err)
    
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    producer := NewFileProducer(inputPath, logger)
    presenter := NewFilePresenter(outputPath, logger)
    maskerImpl := DigitsMasker{}
    service := NewService(producer, presenter, maskerImpl, logger)
    
    ctx := context.Background()
    err = service.Run(ctx)
    
    assert.NoError(t, err)
    
    // Выходной файл должен быть пустым
    outputContent, err := os.ReadFile(outputPath)
    require.NoError(t, err)
    assert.Equal(t, "", string(outputContent))
}