package masker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFilePresenter(t *testing.T) {
	p := NewFilePresenter("out.txt")
	assert.NotNil(t, p)
	assert.Equal(t, "out.txt", p.path)
}

func TestFilePresenterPresentSuccess(t *testing.T) {
	tmpFile := createTempFile(t, []string{})
	deleteTempFile(tmpFile)

	p := NewFilePresenter(tmpFile)
	data := []string{"hello", "world"}
	err := p.Present(data)
	assert.NoError(t, err)

	result, err := readFileLines(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, data, result)

	os.Remove(tmpFile)
}

func TestFilePresenterPresentCreateError(t *testing.T) {
	p := NewFilePresenter("/nonexistent_dir_12345/out.txt")
	err := p.Present([]string{"test"})
	assert.Error(t, err)
}
