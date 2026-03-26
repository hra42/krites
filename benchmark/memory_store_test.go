package benchmark

import (
	"testing"
	"time"
)

func newTestSuite(id, name string) *Suite {
	return &Suite{
		ID:          id,
		Name:        name,
		Description: "Test suite",
		Prompts: []Prompt{
			{ID: "p1", Name: "test prompt", UserMessage: "Hello"},
		},
		Models:    []string{"openai/gpt-4o"},
		Config:    DefaultRunConfig(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestMemoryStore_CreateAndGetSuite(t *testing.T) {
	store := NewMemoryStore()
	suite := newTestSuite("s1", "Test Suite")

	if err := store.CreateSuite(suite); err != nil {
		t.Fatalf("CreateSuite failed: %v", err)
	}

	got, err := store.GetSuite("s1")
	if err != nil {
		t.Fatalf("GetSuite failed: %v", err)
	}
	if got.Name != "Test Suite" {
		t.Errorf("expected name 'Test Suite', got %q", got.Name)
	}
}

func TestMemoryStore_GetSuite_NotFound(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.GetSuite("nonexistent")
	if err != ErrSuiteNotFound {
		t.Errorf("expected ErrSuiteNotFound, got %v", err)
	}
}

func TestMemoryStore_ListSuites(t *testing.T) {
	store := NewMemoryStore()

	suites, err := store.ListSuites()
	if err != nil {
		t.Fatalf("ListSuites failed: %v", err)
	}
	if len(suites) != 0 {
		t.Errorf("expected 0 suites, got %d", len(suites))
	}

	store.CreateSuite(newTestSuite("s1", "Suite 1"))
	store.CreateSuite(newTestSuite("s2", "Suite 2"))

	suites, err = store.ListSuites()
	if err != nil {
		t.Fatalf("ListSuites failed: %v", err)
	}
	if len(suites) != 2 {
		t.Errorf("expected 2 suites, got %d", len(suites))
	}
}

func TestMemoryStore_UpdateSuite(t *testing.T) {
	store := NewMemoryStore()
	suite := newTestSuite("s1", "Original")
	store.CreateSuite(suite)

	suite.Name = "Updated"
	if err := store.UpdateSuite(suite); err != nil {
		t.Fatalf("UpdateSuite failed: %v", err)
	}

	got, _ := store.GetSuite("s1")
	if got.Name != "Updated" {
		t.Errorf("expected name 'Updated', got %q", got.Name)
	}
}

func TestMemoryStore_UpdateSuite_NotFound(t *testing.T) {
	store := NewMemoryStore()
	suite := newTestSuite("nonexistent", "Ghost")

	err := store.UpdateSuite(suite)
	if err != ErrSuiteNotFound {
		t.Errorf("expected ErrSuiteNotFound, got %v", err)
	}
}

func TestMemoryStore_DeleteSuite(t *testing.T) {
	store := NewMemoryStore()
	store.CreateSuite(newTestSuite("s1", "To Delete"))

	if err := store.DeleteSuite("s1"); err != nil {
		t.Fatalf("DeleteSuite failed: %v", err)
	}

	_, err := store.GetSuite("s1")
	if err != ErrSuiteNotFound {
		t.Errorf("expected ErrSuiteNotFound after delete, got %v", err)
	}
}

func TestMemoryStore_DeleteSuite_NotFound(t *testing.T) {
	store := NewMemoryStore()

	err := store.DeleteSuite("nonexistent")
	if err != ErrSuiteNotFound {
		t.Errorf("expected ErrSuiteNotFound, got %v", err)
	}
}

func TestMemoryStore_CreateAndGetRun(t *testing.T) {
	store := NewMemoryStore()
	run := &Run{
		ID:        "r1",
		SuiteID:   "s1",
		SuiteName: "Test Suite",
		Status:    RunStatusPending,
		Config:    DefaultRunConfig(),
		CreatedAt: time.Now(),
	}

	if err := store.CreateRun(run); err != nil {
		t.Fatalf("CreateRun failed: %v", err)
	}

	got, err := store.GetRun("r1")
	if err != nil {
		t.Fatalf("GetRun failed: %v", err)
	}
	if got.Status != RunStatusPending {
		t.Errorf("expected status 'pending', got %q", got.Status)
	}
}

func TestMemoryStore_GetRun_NotFound(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.GetRun("nonexistent")
	if err != ErrRunNotFound {
		t.Errorf("expected ErrRunNotFound, got %v", err)
	}
}
