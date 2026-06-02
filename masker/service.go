package masker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Producer — поставщик данных
type Producer interface {
	Produce() ([]string, error)
}

// Presenter — обработчик вывода
type Presenter interface {
	Present([]string) error
}

// job — задача для воркера: индекс строки и её текст
type job struct {
	index int
	text  string
}

// result — результат воркера: индекс строки и замаскированный текст
type result struct {
	index int
	text  string
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

// Run — главный метод сервиса.
// Читает строки, запускает worker pool (до 10 горутин), маскирует параллельно,
// собирает результаты в правильном порядке и записывает в файл.
// Все горутины завершаются при отмене контекста (Ctrl+C).
func (s *Service) Run(ctx context.Context) error {
	s.logger.DebugContext(ctx, "service: starting")

	// --- чтение входных данных ---
	s.logger.DebugContext(ctx, "producer: reading input file")
	raw, err := s.prod.Produce()
	if err != nil {
		s.logger.ErrorContext(ctx, "producer: failed to read input", slog.Any("error", err))
		return fmt.Errorf("producer error: %w", err)
	}
	s.logger.InfoContext(ctx, "producer: input file read", slog.Int("lines", len(raw)))

	if len(raw) == 0 {
		s.logger.WarnContext(ctx, "producer: input file is empty, nothing to mask")
		return s.pres.Present([]string{})
	}

	// --- настройка worker pool ---
	numWorkers := 10
	if len(raw) < numWorkers {
		numWorkers = len(raw)
	}
	s.logger.DebugContext(ctx, "worker pool: configuring",
		slog.Int("workers", numWorkers),
		slog.Int("jobs", len(raw)),
	)

	jobsCh := make(chan job, len(raw))
	resultsCh := make(chan result, len(raw))

	// заполняем канал задачами
	for i, line := range raw {
		jobsCh <- job{index: i, text: line}
	}
	close(jobsCh)
	s.logger.DebugContext(ctx, "worker pool: jobs channel filled and closed")

	// --- запуск воркеров ---
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		workerID := w
		go func() {
			defer wg.Done()
			s.logger.DebugContext(ctx, "worker: started", slog.Int("worker_id", workerID))

			for {
				// проверяем отмену контекста перед каждой задачей
				select {
				case <-ctx.Done():
					s.logger.WarnContext(ctx, "worker: context cancelled, stopping",
						slog.Int("worker_id", workerID),
					)
					return
				default:
				}

				j, ok := <-jobsCh
				if !ok {
					s.logger.DebugContext(ctx, "worker: jobs channel closed, finishing",
						slog.Int("worker_id", workerID),
					)
					return
				}

				s.logger.DebugContext(ctx, "worker: processing line",
					slog.Int("worker_id", workerID),
					slog.Int("line_index", j.index),
				)

				masked := s.masker.Mask(j.text)
				resultsCh <- result{index: j.index, text: masked}
			}
		}()
	}

	// --- горутина-коллектор: ждёт всех воркеров, затем закрывает канал результатов ---
	go func() {
		s.logger.DebugContext(ctx, "collector: waiting for all workers to finish")
		wg.Wait()
		s.logger.DebugContext(ctx, "collector: all workers done, closing results channel")
		close(resultsCh)
	}()

	// --- сбор результатов с проверкой контекста ---
	masked := make([]string, len(raw))
	collected := 0
	for {
		select {
		case <-ctx.Done():
			s.logger.WarnContext(ctx, "service: context cancelled during result collection",
				slog.Int("collected", collected),
				slog.Int("total", len(raw)),
			)
			return ctx.Err()
		case r, ok := <-resultsCh:
			if !ok {
				// канал закрыт — все результаты собраны
				goto done
			}
			masked[r.index] = r.text
			collected++
			s.logger.DebugContext(ctx, "collector: received result",
				slog.Int("line_index", r.index),
				slog.Int("collected", collected),
			)
		}
	}

done:
	s.logger.InfoContext(ctx, "service: masking complete", slog.Int("lines_processed", collected))

	// --- запись результата ---
	s.logger.DebugContext(ctx, "presenter: writing output file")
	if err := s.pres.Present(masked); err != nil {
		s.logger.ErrorContext(ctx, "presenter: failed to write output", slog.Any("error", err))
		return fmt.Errorf("presenter error: %w", err)
	}
	s.logger.InfoContext(ctx, "presenter: output file written successfully")

	s.logger.DebugContext(ctx, "service: finished")
	return nil
}
