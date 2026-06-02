package masker
 
import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sort"
	"testing"
)
 
// nopLogger — тихий логгер для тестов
func nopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // в тестах показываем только ошибки
	}))
}
 
type mockProducer struct {
	data []string
	err  error
}
 
func (m *mockProducer) Produce() ([]string, error) {
	return m.data, m.err
}
 
type mockPresenter struct {
	receivedData []string
	err          error
}
 
func (m *mockPresenter) Present(data []string) error {
	m.receivedData = data
	return m.err
}
 
func TestRunSuccess(t *testing.T) {
	prod := &mockProducer{data: []string{"a1", "b2"}, err: nil}
	pres := &mockPresenter{err: nil}
	svc := NewService(prod, pres, DigitsMasker{}, nopLogger())
 
	err := svc.Run(context.Background())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
 
	expected := []string{"a*", "b*"}
	got := make([]string, len(pres.receivedData))
	copy(got, pres.receivedData)
	sort.Strings(got)
	sort.Strings(expected)
 
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], got[i])
		}
	}
}
 
func TestRunProducerError(t *testing.T) {
	prod := &mockProducer{data: nil, err: errors.New("produce error")}
	pres := &mockPresenter{}
	svc := NewService(prod, pres, DigitsMasker{}, nopLogger())
 
	err := svc.Run(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
	if pres.receivedData != nil {
		t.Error("present should not be called on producer error")
	}
}
 
func TestRunEmptyInput(t *testing.T) {
	prod := &mockProducer{data: []string{}, err: nil}
	pres := &mockPresenter{}
	svc := NewService(prod, pres, DigitsMasker{}, nopLogger())
 
	err := svc.Run(context.Background())
	if err != nil {
		t.Errorf("expected no error on empty input, got %v", err)
	}
}
 
func TestRunContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем сразу
 
	prod := &mockProducer{data: []string{"line1", "line2", "line3"}, err: nil}
	pres := &mockPresenter{}
	svc := NewService(prod, pres, DigitsMasker{}, nopLogger())
 
	err := svc.Run(ctx)
	// при отмене контекста ожидаем либо nil, либо context.Canceled
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("expected nil or context.Canceled, got %v", err)
	}
}
 
func TestDigitsMasker(t *testing.T) {
	m := DigitsMasker{}
	tests := []struct {
		input    string
		expected string
	}{
		{"abc123", "abc***"},
		{"no digits", "no digits"},
		{"123", "***"},
	}
	for _, tt := range tests {
		result := m.Mask(tt.input)
		if result != tt.expected {
			t.Errorf("Mask(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
