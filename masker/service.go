package masker

import (
    "context"
    "log/slog"
    "sync"
)

type Service struct {
    prod   Producer
    pres   Presenter
    masker Masker
    logger *slog.Logger
}

func NewService(prod Producer, pres Presenter, masker Masker, logger *slog.Logger) *Service {
    return &Service{
        prod:   prod,
        pres:   pres,
        masker: masker,
        logger: logger,
    }
}

type job struct {
    index int
    text  string
}

type result struct {
    index int
    text  string
}

func (s *Service) Run(ctx context.Context) error {
    s.logger.InfoContext(ctx, "начало обработки")

    raw, err := s.prod.Produce(ctx)
    if err != nil {
        s.logger.ErrorContext(ctx, "ошибка чтения", "error", err)
        return err
    }

    s.logger.DebugContext(ctx, "прочитано строк", "count", len(raw))

    if len(raw) == 0 {
        s.logger.WarnContext(ctx, "файл пуст")
        return s.pres.Present(ctx, []string{})
    }

    numWorkers := 10
    if len(raw) < numWorkers {
        numWorkers = len(raw)
    }

    jobsCh := make(chan job, len(raw))
    resultsCh := make(chan result, len(raw))

    for i, line := range raw {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case jobsCh <- job{index: i, text: line}:
        }
    }
    close(jobsCh)

    var wg sync.WaitGroup
    for w := 0; w < numWorkers; w++ {
        wg.Add(1)
        go s.worker(ctx, &wg, w, jobsCh, resultsCh)
    }

    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    masked := make([]string, len(raw))
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case r, ok := <-resultsCh:
            if !ok {
                return s.pres.Present(ctx, masked)
            }
            masked[r.index] = r.text
        }
    }
}

func (s *Service) worker(ctx context.Context, wg *sync.WaitGroup, id int, jobsCh <-chan job, resultsCh chan<- result) {
    defer wg.Done()
    s.logger.DebugContext(ctx, "воркер запущен", "worker_id", id)

    for {
        select {
        case <-ctx.Done():
            s.logger.WarnContext(ctx, "воркер остановлен", "worker_id", id)
            return
        case j, ok := <-jobsCh:
            if !ok {
                s.logger.DebugContext(ctx, "воркер завершён", "worker_id", id)
                return
            }
            res := s.masker.Mask(ctx, j.text)
            select {
            case <-ctx.Done():
                return
            case resultsCh <- result{index: j.index, text: res}:
            }
        }
    }
}