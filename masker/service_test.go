package masker

import (
    "context"
    "errors"
    "log/slog"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
)

type mockProducer struct {
    data []string
    err  error
}

func (m *mockProducer) Produce(ctx context.Context) ([]string, error) {
    return m.data, m.err
}

type mockPresenter struct {
    receivedData []string
    err          error
}

func (m *mockPresenter) Present(ctx context.Context, data []string) error {
    m.receivedData = data
    return m.err
}

type mockMasker struct {
    result string
}

func (m *mockMasker) Mask(ctx context.Context, line string) string {
    return m.result
}

func TestServiceRunSuccess(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

    prod := &mockProducer{data: []string{"a1", "b2"}, err: nil}
    pres := &mockPresenter{err: nil}
    mask := &mockMasker{result: "***"}
    svc := NewService(prod, pres, mask, logger)

    err := svc.Run(context.Background())

    assert.NoError(t, err)
    assert.Equal(t, []string{"***", "***"}, pres.receivedData)
}

func TestServiceRunProducerError(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

    prod := &mockProducer{data: nil, err: errors.New("produce error")}
    pres := &mockPresenter{}
    mask := &mockMasker{result: "***"}
    svc := NewService(prod, pres, mask, logger)

    err := svc.Run(context.Background())

    assert.Error(t, err)
    assert.Nil(t, pres.receivedData)
}

func TestServiceRunContextCancel(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    prod := &mockProducer{data: []string{"a1", "b2"}, err: nil}
    pres := &mockPresenter{err: nil}
    mask := &mockMasker{result: "***"}
    svc := NewService(prod, pres, mask, logger)

    err := svc.Run(ctx)

    assert.Error(t, err)
}

func TestServiceRunPresenterError(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

    prod := &mockProducer{data: []string{"a1"}, err: nil}
    pres := &mockPresenter{err: errors.New("presenter error")}
    mask := &mockMasker{result: "***"}
    svc := NewService(prod, pres, mask, logger)

    err := svc.Run(context.Background())

    assert.Error(t, err)
    assert.EqualError(t, err, "presenter error")
}

func TestServiceRunWithManyWorkers(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

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