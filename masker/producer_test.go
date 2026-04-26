package masker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileProducer(t *testing.T) {
	p := NewFileProducer("test.txt")
	assert.NotNil(t, p)
	assert.Equal(t, "test.txt", p.path)
}

func TestFileProducerProduceSuccess(t *testing.T) {
	// Создаём временный файл
	content := []string{"line1", "line2", "line3"}
	tmpFile := createTempFile(t, content)
	defer deleteTempFile(tmpFile)

	p := NewFileProducer(tmpFile)
	result, err := p.Produce()

	assert.NoError(t, err)
	assert.Equal(t, content, result)
}

func TestFileProducerProduceFileNotFound(t *testing.T) {
	p := NewFileProducer("nonexistent_file_12345.txt")
	_, err := p.Produce()
	assert.Error(t, err)
}
