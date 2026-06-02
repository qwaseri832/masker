package masker

import "context"

type DigitsMasker struct{}

func (m DigitsMasker) Mask(ctx context.Context, line string) string {
    out := make([]rune, 0, len(line))
    for _, ch := range line {
        if ch >= '0' && ch <= '9' {
            out = append(out, '*')
        } else {
            out = append(out, ch)
        }
    }
    return string(out)
}