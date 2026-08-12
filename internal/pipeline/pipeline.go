package pipeline

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/qwaseri832/masker/internal/mask"
)

const ctxCheckEvery = 1024

type Stats struct {
	Lines int64
}

func Run(ctx context.Context, dst io.Writer, src io.Reader, m mask.Masker) (Stats, error) {
	var st Stats

	r := bufio.NewReader(src)
	w := bufio.NewWriter(dst)

	for {
		if st.Lines%ctxCheckEvery == 0 {
			if err := ctx.Err(); err != nil {
				return st, err
			}
		}

		line, readErr := r.ReadString('\n')
		if line != "" {
			body, eol := splitEOL(line)
			if _, err := w.WriteString(m.Mask(body)); err != nil {
				return st, fmt.Errorf("запись: %w", err)
			}
			if _, err := w.WriteString(eol); err != nil {
				return st, fmt.Errorf("запись: %w", err)
			}
			st.Lines++
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return st, fmt.Errorf("чтение: %w", readErr)
		}
	}

	if err := w.Flush(); err != nil {
		return st, fmt.Errorf("запись: %w", err)
	}
	return st, nil
}

func splitEOL(line string) (body, eol string) {
	if n := len(line); n > 0 && line[n-1] == '\n' {
		if n > 1 && line[n-2] == '\r' {
			return line[:n-2], "\r\n"
		}
		return line[:n-1], "\n"
	}
	return line, ""
}
