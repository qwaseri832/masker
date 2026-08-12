package mask

import "strings"

type Masker interface {
	Mask(line string) string
}

type MaskerFunc func(line string) string

func (f MaskerFunc) Mask(line string) string { return f(line) }

const Rune = '*'

type Digits struct{}

func (Digits) Mask(line string) string {
	first := strings.IndexFunc(line, isDigit)
	if first < 0 {
		return line
	}

	var b strings.Builder
	b.Grow(len(line))
	b.WriteString(line[:first])
	for _, r := range line[first:] {
		if isDigit(r) {
			b.WriteRune(Rune)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }
