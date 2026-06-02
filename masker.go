package masker

type Masker struct {
    // структура для маскирования данных
}

func NewMasker() *Masker {
    return &Masker{}
}

func (m *Masker) Mask(data string) string {
    // простая реализация
    if len(data) <= 4 {
        return "****"
    }
    return data[:2] + "****" + data[len(data)-2:]
}
