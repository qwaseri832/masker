package masker

import (
    "sync"
)

// Producer поставщик данных
type Producer interface {
    Produce() ([]string, error)
}

// Presenter обработчик вывода
type Presenter interface {
    Present([]string) error
}

// Masker интерфейс стратегии маскирования
type Masker interface {
    Mask(string) string
}

// Service — основная бизнес-логика
type Service struct {
    prod   Producer
    pres   Presenter
    masker Masker
}

// NewService — конструктор сервиса
func NewService(prod Producer, pres Presenter, masker Masker) *Service {
    return &Service{
        prod:   prod,
        pres:   pres,
        masker: masker,
    }
}

// job структура для передачи задачи и сохранения исходного индекса строки
type job struct {
    index int
    text  string
}

// result структура для получения замаскированной строки с её индексом
type result struct {
    index int
    text  string
}

// Run — главный метод сервиса с использованием Worker Pool (ровно 10 рутин)
func (s *Service) Run() error {
    raw, err := s.prod.Produce()
    if err != nil {
        return err
    }

    if len(raw) == 0 {
        return s.pres.Present([]string{})
    }

    numWorkers := 10
    if len(raw) < numWorkers {
        numWorkers = len(raw)
    }

    jobsCh := make(chan job, len(raw))
    resultsCh := make(chan result, len(raw))

    // Заполняем канал задачами
    for i, line := range raw {
        jobsCh <- job{index: i, text: line}
    }
    close(jobsCh)

    // Запускаем ровно 10 воркеров-рутин
    var wg sync.WaitGroup
    for w := 0; w < numWorkers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := range jobsCh {
                // Каждый вызов маскирования выполняется внутри запущенной рутины
                res := s.masker.Mask(j.text)
                resultsCh <- result{index: j.index, text: res}
            }
        }()
    }

    // Ждем завершения всех рутин и закрываем канал результатов (Fan-In)
    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    // Сбор результатов и восстановление правильного порядка строк
    masked := make([]string, len(raw))
    for r := range resultsCh {
        masked[r.index] = r.text
    }

    return s.pres.Present(masked)
}
