package masker

import (
    "context"
    "os"
    "log/slog"
)

type FilePresenter struct {
    path   string
    logger *slog.Logger
}

func NewFilePresenter(path string, logger *slog.Logger) *FilePresenter {
    return &FilePresenter{path: path, logger: logger}
}

func (p *FilePresenter) Present(ctx context.Context, data []string) error {
    p.logger.DebugContext(ctx, "запись в файл", "path", p.path, "lines", len(data))

    f, err := os.Create(p.path)
    if err != nil {
        return err
    }
    defer f.Close()

    for _, line := range data {
        if _, err := f.WriteString(line + "\n"); err != nil {
            return err
        }
    }
    return nil
}