package masker

import (
    "context"
    "log/slog"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestServiceRunWithManyWorkers(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    // Создаём много данных для проверки работы воркеров
    data := make([]string, 50)
    for i := 0; i < 50; i++ {
        data[i] = "test123"
    }
    
    prod := &mockProducer{data: data, err: nil}
    pres := &mockPresenter{err: nil}
    mask := &mockMasker{result: "***"}
    svc := NewService(prod, pres, mask, logger)
    
    err := svc.Run(context.Background())
    
    assert.NoError(t, err)
    assert.Len(t, pres.receivedData, 50)
}

func TestServiceRunWithContextCancelDuringProcessing(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    // Создаём данные, которые будут обрабатываться долго
    data := make([]string, 100)
    for i := 0; i < 100; i++ {
        data[i] = "value"
    }
    
    prod := &mockProducer{data: data, err: nil}
    pres := &mockPresenter{err: nil}
    mask := &mockMasker{result: "masked"}
    svc := NewService(prod, pres, mask, logger)
    
    ctx, cancel := context.WithCancel(context.Background())
    
    // Отменяем контекст через небольшую задержку
    go func() {
        time.Sleep(10 * time.Millisecond)
        cancel()
    }()
    
    err := svc.Run(ctx)
    
    // Ожидаем ошибку отмены контекста
    assert.Error(t, err)
}

func TestServiceRunWithEmptyData(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    prod := &mockProducer{data: []string{}, err: nil}
    pres := &mockPresenter{err: nil}
    mask := &mockMasker{result: ""}
    svc := NewService(prod, pres, mask, logger)
    
    err := svc.Run(context.Background())
    
    assert.NoError(t, err)
    assert.Empty(t, pres.receivedData)
}

func TestServiceRunWithSingleItem(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    prod := &mockProducer{data: []string{"single"}, err: nil}
    pres := &mockPresenter{err: nil}
    mask := &mockMasker{result: "masked"}
    svc := NewService(prod, pres, mask, logger)
    
    err := svc.Run(context.Background())
    
    assert.NoError(t, err)
    assert.Equal(t, []string{"masked"}, pres.receivedData)
}