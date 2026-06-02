package masker

import (
    "bufio"
    "context"
    "os"
    "log/slog"
)

type FileProducer struct {
    path   string
    logger *slog.Logger
}

func NewFileProducer(path string, logger *slog.Logger) *FileProducer {
    return &FileProducer{path: path, logger: logger}
}

func (p *FileProducer) Produce(ctx context.Context) ([]string, error) {
    p.logger.DebugContext(ctx, "чтение файла", "path", p.path)

    f, err := os.Open(p.path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    var lines []string
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}