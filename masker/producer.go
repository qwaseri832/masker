package masker

import (
    "bufio"
    "os"
)

// FileProducer — читает файл и возвращает строки
type FileProducer struct {
    path string
}

// NewFileProducer — конструктор FileProducer
func NewFileProducer(path string) *FileProducer {
    return &FileProducer{path: path}
}

// Produce — реализация интерфейса Producer
func (p *FileProducer) Produce() ([]string, error) {
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