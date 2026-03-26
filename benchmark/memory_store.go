package benchmark

import "sync"

// MemoryStore implements Store with sync.RWMutex-protected maps.
type MemoryStore struct {
	mu     sync.RWMutex
	suites map[string]*Suite
	runs   map[string]*Run
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		suites: make(map[string]*Suite),
		runs:   make(map[string]*Run),
	}
}

func (s *MemoryStore) CreateSuite(suite *Suite) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.suites[suite.ID] = suite
	return nil
}

func (s *MemoryStore) GetSuite(id string) (*Suite, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	suite, ok := s.suites[id]
	if !ok {
		return nil, ErrSuiteNotFound
	}
	return suite, nil
}

func (s *MemoryStore) ListSuites() ([]*Suite, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Suite, 0, len(s.suites))
	for _, suite := range s.suites {
		result = append(result, suite)
	}
	return result, nil
}

func (s *MemoryStore) UpdateSuite(suite *Suite) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.suites[suite.ID]; !ok {
		return ErrSuiteNotFound
	}
	s.suites[suite.ID] = suite
	return nil
}

func (s *MemoryStore) DeleteSuite(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.suites[id]; !ok {
		return ErrSuiteNotFound
	}
	delete(s.suites, id)
	return nil
}

func (s *MemoryStore) CreateRun(run *Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs[run.ID] = run
	return nil
}

func (s *MemoryStore) GetRun(id string) (*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	run, ok := s.runs[id]
	if !ok {
		return nil, ErrRunNotFound
	}
	return run, nil
}

func (s *MemoryStore) ListRuns() ([]*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Run, 0, len(s.runs))
	for _, run := range s.runs {
		result = append(result, run)
	}
	return result, nil
}

func (s *MemoryStore) UpdateRun(run *Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.runs[run.ID]; !ok {
		return ErrRunNotFound
	}
	s.runs[run.ID] = run
	return nil
}
