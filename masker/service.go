package masker

// Producer интерфейс поставщика данных
type Producer interface {
	Produce() ([]string, error)
}

// Presenter интерфейс отображения результата
type Presenter interface {
	Present([]string) error
}

// Service структура с бизнес-логикой
type Service struct {
	producer  Producer
	presenter Presenter
}

// NewService конструктор
func NewService(p Producer, pr Presenter) *Service {
	return &Service{producer: p, presenter: pr}
}

// функция маскирования (НЕ метод, просто функция, сигнатура не менялась)
func maskData(data []string) []string {
	result := make([]string, len(data))
	for i, line := range data {
		masked := ""
		for _, ch := range line {
			if ch >= '0' && ch <= '9' {
				masked += "*"
			} else {
				masked += string(ch)
			}
		}
		result[i] = masked
	}
	return result
}

// Run главный метод
func (s *Service) Run() error {
	raw, err := s.producer.Produce()
	if err != nil {
		return err
	}
	masked := maskData(raw)
	return s.presenter.Present(masked)
}
