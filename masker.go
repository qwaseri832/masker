package masker

// Masker — интерфейс стратегии маскирования
type Masker interface {
    Mask(line string) string
}

// DigitsMasker — заменяет все цифры на *
type DigitsMasker struct{}

func (m DigitsMasker) Mask(line string) string {
    out := ""
    for _, ch := range line {
        if ch >= '0' && ch <= '9' {
            out += "*"
        } else {
            out += string(ch)
        }
    }
    return out
}
