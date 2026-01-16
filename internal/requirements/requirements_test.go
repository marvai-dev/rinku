package requirements

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stephan/rinku/internal/progress"
)

func TestSetAndGet(t *testing.T) {
	dir := t.TempDir()

	err := Set(dir, "api/cli", "--port, --config")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	req, err := Get(dir, "api/cli")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if req == nil {
		t.Fatal("expected requirement, got nil")
	}
	if req.Path != "api/cli" {
		t.Errorf("Path = %q, want %q", req.Path, "api/cli")
	}
	if req.Content != "--port, --config" {
		t.Errorf("Content = %q, want %q", req.Content, "--port, --config")
	}
}

func TestGet_NotFound(t *testing.T) {
	dir := t.TempDir()

	req, err := Get(dir, "nonexistent")
	if err != nil {
		t.Fatalf("Get should not error for missing: %v", err)
	}
	if req != nil {
		t.Error("expected nil for missing requirement")
	}
}

func TestSet_PreservesCreatedAt(t *testing.T) {
	dir := t.TempDir()

	// First set
	err := Set(dir, "api/cli", "v1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	req1, _ := Get(dir, "api/cli")
	createdAt := req1.CreatedAt

	// Second set (update)
	err = Set(dir, "api/cli", "v2")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	req2, _ := Get(dir, "api/cli")
	if req2.Content != "v2" {
		t.Errorf("Content = %q, want %q", req2.Content, "v2")
	}
	if !req2.CreatedAt.Equal(createdAt) {
		t.Error("CreatedAt should be preserved on update")
	}
	if !req2.UpdatedAt.After(req2.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt on update")
	}
}

func TestSet_AutoTagsStep(t *testing.T) {
	dir := t.TempDir()

	// Create progress with current step
	m := progress.New(dir, []string{"1", "2a", "2b"})
	_ = m.StartStep("2a")
	_ = m.Save(dir)

	// Set requirement
	err := Set(dir, "api/cli", "content")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	req, _ := Get(dir, "api/cli")
	if req.Step != "2a" {
		t.Errorf("Step = %q, want %q", req.Step, "2a")
	}
}

func TestSet_NestedPath(t *testing.T) {
	dir := t.TempDir()

	err := Set(dir, "api/web/routes/users", "GET /, POST /")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify file exists at expected path
	path := filepath.Join(dir, ".rinku", "requirements", "api", "web", "routes", "users.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file at %s", path)
	}

	req, err := Get(dir, "api/web/routes/users")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if req.Content != "GET /, POST /" {
		t.Errorf("Content = %q", req.Content)
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()

	_ = Set(dir, "api/cli", "cli")
	_ = Set(dir, "api/web/routes", "routes")
	_ = Set(dir, "worker/jobs", "jobs")
	_ = Set(dir, "db/models/user", "user")

	paths, err := List(dir, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	expected := []string{"api/cli", "api/web/routes", "db/models/user", "worker/jobs"}
	if len(paths) != len(expected) {
		t.Fatalf("len(paths) = %d, want %d: %v", len(paths), len(expected), paths)
	}
	for i, p := range expected {
		if paths[i] != p {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], p)
		}
	}
}

func TestList_WithPrefix(t *testing.T) {
	dir := t.TempDir()

	_ = Set(dir, "api/cli", "cli")
	_ = Set(dir, "api/web/routes", "routes")
	_ = Set(dir, "worker/jobs", "jobs")

	paths, err := List(dir, "api/")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(paths), paths)
	}
	if paths[0] != "api/cli" || paths[1] != "api/web/routes" {
		t.Errorf("paths = %v", paths)
	}
}

func TestList_Empty(t *testing.T) {
	dir := t.TempDir()

	paths, err := List(dir, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected empty list, got %v", paths)
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()

	_ = Set(dir, "api/cli", "cli")

	err := Delete(dir, "api/cli")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	req, _ := Get(dir, "api/cli")
	if req != nil {
		t.Error("expected nil after delete")
	}
}

func TestDelete_NotFound(t *testing.T) {
	dir := t.TempDir()

	err := Delete(dir, "nonexistent")
	if err != nil {
		t.Errorf("Delete should not error for missing: %v", err)
	}
}

func TestList_WithWildcard(t *testing.T) {
	dir := t.TempDir()

	_ = Set(dir, "api/cli", "cli")
	_ = Set(dir, "api/web/routes", "routes")
	_ = Set(dir, "worker/cli", "worker cli")
	_ = Set(dir, "db/models", "models")

	// Test */cli pattern - should match api/cli and worker/cli
	paths, err := List(dir, "*/cli")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(paths), paths)
	}
	if paths[0] != "api/cli" || paths[1] != "worker/cli" {
		t.Errorf("paths = %v, want [api/cli worker/cli]", paths)
	}
}

func TestList_WildcardMiddle(t *testing.T) {
	dir := t.TempDir()

	_ = Set(dir, "api/v1/users", "v1 users")
	_ = Set(dir, "api/v2/users", "v2 users")
	_ = Set(dir, "api/v1/posts", "v1 posts")

	// Test api/*/users pattern
	paths, err := List(dir, "api/*/users")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(paths), paths)
	}
	if paths[0] != "api/v1/users" || paths[1] != "api/v2/users" {
		t.Errorf("paths = %v, want [api/v1/users api/v2/users]", paths)
	}
}
