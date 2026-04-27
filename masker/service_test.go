package masker

import (
	"errors"
	"testing"
)

// MockProducer — ручной мок (простая структура)
type mockProducer struct {
	data []string
	err  error
}

func (m *mockProducer) Produce() ([]string, error) {
	return m.data, m.err
}

// MockPresenter — ручной мок
type mockPresenter struct {
	receivedData []string
	err          error
}

func (m *mockPresenter) Present(data []string) error {
	m.receivedData = data
	return m.err
}

// тест 1: успешный сценарий
func TestRunSuccess(t *testing.T) {
	prod := &mockProducer{data: []string{"a1", "b2"}, err: nil}
	pres := &mockPresenter{err: nil}
	svc := NewService(prod, pres)

	err := svc.Run()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	expected := []string{"a*", "b*"}
	for i := range expected {
		if pres.receivedData[i] != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], pres.receivedData[i])
		}
	}
}

// тест 2: ошибка Producer
func TestRunProducerError(t *testing.T) {
	prod := &mockProducer{data: nil, err: errors.New("produce error")}
	pres := &mockPresenter{}
	svc := NewService(prod, pres)

	err := svc.Run()

	if err == nil {
		t.Error("expected error, got nil")
	}
	if pres.receivedData != nil {
		t.Error("present should not be called on producer error")
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
