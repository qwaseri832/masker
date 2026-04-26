package masker

import "os"

// FilePresenter — записывает строки в файл (перезаписывает)
type FilePresenter struct {
    path string
}

// NewFilePresenter — конструктор FilePresenter
func NewFilePresenter(path string) *FilePresenter {
    return &FilePresenter{path: path}
}

// Present — реализация интерфейса Presenter
func (p *FilePresenter) Present(data []string) error {
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