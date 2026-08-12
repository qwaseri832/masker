package mask_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qwaseri832/masker/internal/mask"
)

func TestDigitsMask(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"пустая строка", "", ""},
		{"без цифр", "просто текст", "просто текст"},
		{"только цифры", "0123456789", "**********"},
		{"цифры в середине", "Привет 123", "Привет ***"},
		{"телефон", "Мой телефон 89123456789", "Мой телефон ***********"},
		{"цифры в начале", "2025 год", "**** год"},
		{"разделители сохраняются", "a1-b2_c3", "a*-b*_c*"},
		{"кириллица не ломается", "дом №5, кв. 12", "дом №*, кв. **"},
		{"не-ASCII цифры не трогаются", "٤٢ и 42", "٤٢ и **"},
		{"эмодзи проходит насквозь", "🎉 7 🎉", "🎉 * 🎉"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, mask.Digits{}.Mask(tt.in))
		})
	}
}

func TestDigitsMaskAvoidsAllocWithoutDigits(t *testing.T) {
	const line = "строка совсем без цифр"
	allocs := testingAllocs(func() { _ = mask.Digits{}.Mask(line) })
	assert.Zero(t, allocs, "маскирование строки без цифр не должно выделять память")
}

func TestMaskerFunc(t *testing.T) {
	var m mask.Masker = mask.MaskerFunc(strings.ToUpper)
	assert.Equal(t, "ABC", m.Mask("abc"))
}

func testingAllocs(fn func()) float64 {
	return testing.AllocsPerRun(100, fn)
}

func BenchmarkDigitsMask(b *testing.B) {
	line := strings.Repeat("заказ 12345 от 2026-08-12; ", 20)
	b.ReportAllocs()
	for b.Loop() {
		_ = mask.Digits{}.Mask(line)
	}
}
