package masker

import (
	"bufio"
	"os"
	"testing"
)

// createTempFile создаёт временный файл с заданным содержимым
func createTempFile(t *testing.T, lines []string) string {
	t.Helper()
	f, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	for _, line := range lines {
		if _, err := f.WriteString(line + "\n"); err != nil {
			t.Fatal(err)
		}
	}
	return f.Name()
}

// deleteTempFile удаляет временный файл
func deleteTempFile(path string) {
	os.Remove(path)
}

// readFileLines читает файл и возвращает строки
func readFileLines(path string) ([]string, error) {
	f, err := os.Open(path)
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
