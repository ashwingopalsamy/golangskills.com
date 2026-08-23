package replicareplay

import (
	"context"
	"runtime"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) Post(ctx context.Context, request Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if stored, ok := s.store.lookup(request.Key); ok {
		return stored.result, nil
	}

	runtime.Gosched()
	result := s.store.appendEntry(request)
	runtime.Gosched()
	s.store.save(request.Key, record{request: request, result: result})
	return result, nil
}
