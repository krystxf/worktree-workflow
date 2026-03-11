package git

import "testing"

func TestParseWorktreeList(t *testing.T) {
	input := []byte("worktree /home/user/repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /home/user/repo--worktrees/repo--feature\nHEAD def456\nbranch refs/heads/feature\n\n")

	wts := parseWorktreeList(input)
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(wts))
	}
	if wts[0].Path != "/home/user/repo" || wts[0].Branch != "main" {
		t.Errorf("worktree 0: %+v", wts[0])
	}
	if wts[1].Path != "/home/user/repo--worktrees/repo--feature" || wts[1].Branch != "feature" {
		t.Errorf("worktree 1: %+v", wts[1])
	}
}

func TestParseWorktreeListDetachedHead(t *testing.T) {
	input := []byte("worktree /home/user/repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /tmp/detached\nHEAD def456\ndetached\n\n")

	wts := parseWorktreeList(input)
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(wts))
	}
	if wts[1].Branch != "(detached HEAD)" {
		t.Errorf("expected '(detached HEAD)', got %q", wts[1].Branch)
	}
}

func TestParseWorktreeListEmpty(t *testing.T) {
	wts := parseWorktreeList([]byte{})
	if len(wts) != 0 {
		t.Fatalf("expected 0 worktrees, got %d", len(wts))
	}
}

func TestParseWorktreeListNoTrailingNewline(t *testing.T) {
	input := []byte("worktree /home/user/repo\nHEAD abc123\nbranch refs/heads/main")

	wts := parseWorktreeList(input)
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(wts))
	}
	if wts[0].Branch != "main" {
		t.Errorf("expected branch 'main', got %q", wts[0].Branch)
	}
}

func TestFilterExcluded(t *testing.T) {
	// null-separated file list
	input := []byte(".env\x00node_modules/foo.js\x00dist/main.js\x00.env.local\x00")

	// No excludes
	result := filterExcluded(input, nil)
	if string(result) != string(input) {
		t.Errorf("no excludes: expected unchanged, got %q", result)
	}

	// Exclude node_modules
	result = filterExcluded(input, []string{"node_modules"})
	expected := ".env\x00dist/main.js\x00.env.local\x00"
	if string(result) != expected {
		t.Errorf("exclude node_modules: expected %q, got %q", expected, result)
	}

	// Exclude multiple
	result = filterExcluded(input, []string{"node_modules", "dist"})
	expected = ".env\x00.env.local\x00"
	if string(result) != expected {
		t.Errorf("exclude multiple: expected %q, got %q", expected, result)
	}
}

func TestFilterExcludedEmpty(t *testing.T) {
	result := filterExcluded([]byte{}, []string{"foo"})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %q", result)
	}
}
