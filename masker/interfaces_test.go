package masker

import (
    "context"
    "log/slog"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
)

// TestInterfaces проверяет, что все интерфейсы реализуются соответствующими типами
func TestInterfaces_ProducerImplemented(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    var _ Producer = (*FileProducer)(nil)
    producer := NewFileProducer("test.txt", logger)
    assert.NotNil(t, producer)
    
    ctx := context.Background()
    _, err := producer.Produce(ctx)
    // Ожидаем ошибку, так как файла нет, но это проверяет, что метод существует
    _ = err
}

func TestInterfaces_PresenterImplemented(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    var _ Presenter = (*FilePresenter)(nil)
    presenter := NewFilePresenter("test.txt", logger)
    assert.NotNil(t, presenter)
}

func TestInterfaces_MaskerImplemented(t *testing.T) {
    var _ Masker = (*DigitsMasker)(nil)
    masker := DigitsMasker{}
    assert.NotNil(t, masker)
}