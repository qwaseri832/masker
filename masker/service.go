package masker

// Producer поставщик данных
type Producer interface {
    Produce() ([]string, error)
}

// Presenter обработчик вывода
type Presenter interface {
    Present([]string) error
}

// Service — основная бизнес-логика
type Service struct {
    prod   Producer
    pres   Presenter
    masker Masker // добавлено поле для стратегии маскирования
}

// NewService — конструктор сервиса
func NewService(prod Producer, pres Presenter, masker Masker) *Service {
    return &Service{
        prod:   prod,
        pres:   pres,
        masker: masker,
    }
}

// Run — главный метод сервиса
func (s *Service) Run() error {
    raw, err := s.prod.Produce()
    if err != nil {
        return err
    }

    // применяем стратегию маскирования к каждой строке
    masked := make([]string, len(raw))
    for i, line := range raw {
        masked[i] = s.masker.Mask(line)
    }

    return s.pres.Present(masked)
}
