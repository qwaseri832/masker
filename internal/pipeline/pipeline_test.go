package pipeline_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qwaseri832/masker/internal/mask"
	"github.com/qwaseri832/masker/internal/pipeline"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		want      string
		wantLines int64
	}{
		{
			name:      "пустой вход",
			in:        "",
			want:      "",
			wantLines: 0,
		},
		{
			name:      "несколько строк",
			in:        "Привет 123\nГод 2025\n",
			want:      "Привет ***\nГод ****\n",
			wantLines: 2,
		},
		{
			name:      "последняя строка без перевода строки",
			in:        "a1\nb2",
			want:      "a*\nb*",
			wantLines: 2,
		},
		{
			name:      "CRLF сохраняется",
			in:        "a1\r\nb2\r\n",
			want:      "a*\r\nb*\r\n",
			wantLines: 2,
		},
		{
			name:      "пустые строки сохраняются",
			in:        "1\n\n2\n",
			want:      "*\n\n*\n",
			wantLines: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			st, err := pipeline.Run(context.Background(), &out, strings.NewReader(tt.in), mask.Digits{})

			require.NoError(t, err)
			assert.Equal(t, tt.want, out.String())
			assert.Equal(t, tt.wantLines, st.Lines)
		})
	}
}

func TestRunHandlesVeryLongLine(t *testing.T) {
	long := strings.Repeat("9", 1<<20)

	var out bytes.Buffer
	st, err := pipeline.Run(context.Background(), &out, strings.NewReader(long), mask.Digits{})

	require.NoError(t, err)
	assert.Equal(t, int64(1), st.Lines)
	assert.Equal(t, strings.Repeat("*", 1<<20), out.String())
}

func TestRunReturnsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out bytes.Buffer
	st, err := pipeline.Run(ctx, &out, strings.NewReader("1\n2\n"), mask.Digits{})

	assert.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, st.Lines)
	assert.Empty(t, out.String())
}

func TestRunReportsReadError(t *testing.T) {
	want := errors.New("диск отвалился")

	var out bytes.Buffer
	_, err := pipeline.Run(context.Background(), &out, failingReader{err: want}, mask.Digits{})

	assert.ErrorIs(t, err, want)
}

func TestRunReportsWriteError(t *testing.T) {
	want := errors.New("некуда писать")

	_, err := pipeline.Run(context.Background(), failingWriter{err: want},
		strings.NewReader("1\n"), mask.Digits{})

	assert.ErrorIs(t, err, want)
}

func TestRunUsesProvidedMasker(t *testing.T) {
	var out bytes.Buffer
	_, err := pipeline.Run(context.Background(), &out, strings.NewReader("abc\n"),
		mask.MaskerFunc(strings.ToUpper))

	require.NoError(t, err)
	assert.Equal(t, "ABC\n", out.String())
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func BenchmarkRun(b *testing.B) {
	input := strings.Repeat("заказ 12345 от 2026-08-12\n", 10_000)
	b.SetBytes(int64(len(input)))
	b.ReportAllocs()

	for b.Loop() {
		_, err := pipeline.Run(context.Background(), io.Discard,
			strings.NewReader(input), mask.Digits{})
		if err != nil {
			b.Fatal(err)
		}
	}
}
