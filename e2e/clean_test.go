package e2e_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClean(t *testing.T) {
	for _, args := range [][]string{{"--clean"}, {"rm", "--clean", "--force"}, {"remove", "--clean"}} {
		t.Run(args[0], func(t *testing.T) {
			tmp := t.TempDir()
			repo, remote := filepath.Join(tmp, "repo"), filepath.Join(tmp, "remote.git")
			initRepo(t, repo)
			gitCmd(t, tmp, "init", "--bare", remote)
			gitCmd(t, repo, "remote", "add", "backup", remote)
			paths := map[string]string{}
			for _, name := range []string{"synced", "dirty", "staged", "untracked", "ignored", "ahead", "behind", "diverged", "deleted", "local", "locked", "current", "renamed"} {
				path := filepath.Join(tmp, name)
				paths[name] = path
				gitCmd(t, repo, "worktree", "add", "-b", name, path)
				if name != "local" {
					gitCmd(t, path, "push", "-u", "backup", name)
				}
			}
			write := func(name, file string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(paths[name], file), []byte("changed"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			write("dirty", "file.txt")
			write("staged", "file.txt")
			gitCmd(t, paths["staged"], "add", "file.txt")
			write("untracked", "new.txt")
			gitCmd(t, paths["ignored"], "config", "core.excludesFile", filepath.Join(tmp, "ignore"))
			if err := os.WriteFile(filepath.Join(tmp, "ignore"), []byte("ignored.txt\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			write("ignored", "ignored.txt")
			for _, name := range []string{"ahead", "behind", "diverged"} {
				gitCmd(t, paths[name], "commit", "--allow-empty", "-m", "extra")
			}
			for _, name := range []string{"behind", "diverged"} {
				gitCmd(t, paths[name], "push")
				gitCmd(t, paths[name], "reset", "--hard", "HEAD~1")
			}
			gitCmd(t, paths["diverged"], "commit", "--allow-empty", "-m", "different")
			// Delete the remote ref directly, leaving a stale local tracking ref.
			gitCmd(t, remote, "update-ref", "-d", "refs/heads/deleted")
			gitCmd(t, repo, "worktree", "lock", paths["locked"])
			gitCmd(t, paths["renamed"], "push", "backup", "HEAD:other-name")
			gitCmd(t, paths["renamed"], "branch", "--set-upstream-to=backup/other-name")
			detached := filepath.Join(tmp, "detached")
			gitCmd(t, repo, "worktree", "add", "--detach", detached)
			stale := filepath.Join(tmp, "stale")
			gitCmd(t, repo, "worktree", "add", "-b", "stale", stale)
			if err := os.RemoveAll(stale); err != nil {
				t.Fatal(err)
			}
			before := gitCmd(t, remote, "show-ref")
			out := mustWtw(t, paths["current"], args...)
			assertOutputContains(t, out, "Removed 2 synced worktree(s)")
			assertOutputContains(t, out, "prunable worktree")
			for name, path := range paths {
				if name == "synced" || name == "renamed" {
					assertDirNotExists(t, path)
					gitCmd(t, repo, "rev-parse", "--verify", "refs/heads/"+name)
				} else {
					assertDirExists(t, path)
				}
			}
			assertDirExists(t, repo)
			assertDirExists(t, detached)
			if after := gitCmd(t, remote, "show-ref"); after != before {
				t.Fatal("clean changed remote refs")
			}
		})
	}
}

func TestCleanRemoteFailure(t *testing.T) {
	tmp := t.TempDir()
	repo, path := filepath.Join(tmp, "repo"), filepath.Join(tmp, "linked")
	initRepo(t, repo)
	gitCmd(t, repo, "worktree", "add", "-b", "linked", path)
	gitCmd(t, repo, "remote", "add", "origin", filepath.Join(tmp, "missing.git"))
	gitCmd(t, repo, "config", "branch.linked.remote", "origin")
	gitCmd(t, repo, "config", "branch.linked.merge", "refs/heads/linked")
	if out, err := wtw(t, repo, "--clean"); err == nil {
		t.Fatalf("expected remote failure: %s", out)
	}
	assertDirExists(t, path)
}

func TestCleanRejectsBranch(t *testing.T) {
	if _, err := wtw(t, t.TempDir(), "rm", "--clean", "branch"); err == nil {
		t.Fatal("expected argument error")
	}
}
