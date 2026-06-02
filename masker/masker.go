package masker

import (
    "context"
    "log/slog"
    "sync"
)

// Producer поставщик данных
type Producer interface {
    Produce(ctx context.Context) ([]string, error)
}

// Presenter обработчик вывода
type Presenter interface {
    Present(ctx context.Context, data []string) error
}

// Masker интерфейс стратегии маскирования
type Masker interface {
    Mask(ctx context.Context, line string) string
}

// Service — основная бизнес-логика
type Service struct {
    prod   Producer
    pres   Presenter
    masker Masker
    logger *slog.Logger
}

// NewService — конструктор сервиса
func NewService(prod Producer, pres Presenter, masker Masker, logger *slog.Logger) *Service {
    return &Service{
        prod:   prod,
        pres:   pres,
        masker: masker,
        logger: logger,
    }
}

// job структура для передачи задачи
type job struct {
    index int
    text  string
}

// result структура для получения замаскированной строки
type result struct {
    index int
    text  string
}

// Run — главный метод сервиса с Worker Pool
func (s *Service) Run(ctx context.Context) error {
    s.logger.DebugContext(ctx, "starting service", "phase", "produce")
    
    raw, err := s.prod.Produce(ctx)
    if err != nil {
        s.logger.ErrorContext(ctx, "produce failed", "error", err)
        return err
    }
    
    s.logger.DebugContext(ctx, "data produced", "lines", len(raw))
    
    if len(raw) == 0 {
        s.logger.DebugContext(ctx, "no data to process", "phase", "present")
        return s.pres.Present(ctx, []string{})
    }
    
    numWorkers := 10
    if len(raw) < numWorkers {
        numWorkers = len(raw)
    }
    
    jobsCh := make(chan job, len(raw))
    resultsCh := make(chan result, len(raw))
    
    // Заполняем канал задачами
    for i, line := range raw {
        select {
        case <-ctx.Done():
            s.logger.WarnContext(ctx, "context cancelled during job creation")
            return ctx.Err()
        case jobsCh <- job{index: i, text: line}:
        }
    }
    close(jobsCh)
    
    s.logger.DebugContext(ctx, "workers pool started", "workers", numWorkers)
    
    // Запускаем воркеров
    var wg sync.WaitGroup
    for w := 0; w < numWorkers; w++ {
        wg.Add(1)
        go s.worker(ctx, &wg, w, jobsCh, resultsCh)
    }
    
    // Ждем завершения воркеров
    go func() {
        wg.Wait()
        close(resultsCh)
        s.logger.DebugContext(ctx, "all workers completed")
    }()
    
    // Сбор результатов
    masked := make([]string, len(raw))
    for r := range resultsCh {
        masked[r.index] = r.text
    }
    
    s.logger.DebugContext(ctx, "masking completed", "phase", "present")
    
    return s.pres.Present(ctx, masked)
}

func (s *Service) worker(ctx context.Context, wg *sync.WaitGroup, id int, jobsCh <-chan job, resultsCh chan<- result) {
    defer wg.Done()
    
    s.logger.DebugContext(ctx, "worker started", "worker_id", id)
    
    for {
        select {
        case <-ctx.Done():
            s.logger.WarnContext(ctx, "worker stopped by context", "worker_id", id)
            return
        case j, ok := <-jobsCh:
            if !ok {
                s.logger.DebugContext(ctx, "worker finished (no more jobs)", "worker_id", id)
                return
            }
            res := s.masker.Mask(ctx, j.text)
            select {
            case <-ctx.Done():
                s.logger.WarnContext(ctx, "worker cancelled while sending result", "worker_id", id)
                return
            case resultsCh <- result{index: j.index, text: res}:
            }
        }
    }
}
