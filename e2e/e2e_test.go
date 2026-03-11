package e2e_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var wtwBin string

func TestMain(m *testing.M) {
	bin := os.Getenv("WTW")
	if bin == "" {
		bin = "../wtw"
	}
	abs, err := filepath.Abs(bin)
	if err != nil {
		panic(err)
	}
	wtwBin = abs
	os.Exit(m.Run())
}

func wtw(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(wtwBin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func mustWtw(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := wtw(t, dir, args...)
	if err != nil {
		t.Fatalf("wtw %v failed: %v\n%s", args, err, out)
	}
	return out
}

func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func currentBranch(t *testing.T, dir string) string {
	t.Helper()
	return gitCmd(t, dir, "branch", "--show-current")
}

func writeGlobalConfig(t *testing.T, content string) {
	t.Helper()
	cfgDir := filepath.Join(os.Getenv("HOME"), ".config", "worktree-workflow")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func defaultGlobalConfig(t *testing.T) {
	t.Helper()
	writeGlobalConfig(t, `{"editor":"echo","auto_open_editor":false}`)
}

func initRepo(t *testing.T, dir string, branches ...string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "init")
	gitCmd(t, dir, "config", "user.email", "test@test.com")
	gitCmd(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", "-A")
	gitCmd(t, dir, "commit", "-m", "init")
	for _, b := range branches {
		gitCmd(t, dir, "branch", b)
	}
}

func realPath(t *testing.T, path string) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		t.Errorf("expected directory to exist: %s", path)
	}
}

func assertDirNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected directory to not exist: %s", path)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		t.Errorf("expected file to exist: %s", path)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		t.Errorf("expected file to not exist: %s", path)
	}
}

func assertFileContents(t *testing.T, path, expected string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	got := strings.TrimRight(string(b), "\n")
	if got != expected {
		t.Errorf("%s: expected %q, got %q", path, expected, got)
	}
}

func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	if !strings.Contains(string(b), substr) {
		t.Errorf("%s does not contain %q", path, substr)
	}
}

func assertOutputContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("output does not contain %q\nOutput: %s", substr, output)
	}
}

func writeProjectConfig(t *testing.T, dir string, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".worktree-workflow.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCreate(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "feature-one")

	gitCmd(t, repo, "checkout", "feature-one")
	if err := os.WriteFile(filepath.Join(repo, "feature.js"), []byte("feature one code"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repo, "add", "feature.js")
	gitCmd(t, repo, "commit", "-m", "add feature one")
	gitCmd(t, repo, "checkout", "main")

	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    true,
		"sync_excludes":   []string{"node_modules"},
		"post_copy_hooks": []string{},
	})
	defaultGlobalConfig(t)

	mustWtw(t, repo, "create", "feature-one")

	worktree := filepath.Join(tmp, "test-project--worktrees", "test-project--feature-one")
	assertDirExists(t, worktree)
	if currentBranch(t, worktree) != "feature-one" {
		t.Errorf("expected branch feature-one")
	}
	assertFileExists(t, filepath.Join(worktree, "feature.js"))
}

func TestCreateSecond(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "feature-one", "feature-two")
	defaultGlobalConfig(t)
	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})

	mustWtw(t, repo, "create", "feature-one")
	mustWtw(t, repo, "create", "feature-two")

	worktree := filepath.Join(tmp, "test-project--worktrees", "test-project--feature-two")
	assertDirExists(t, worktree)
	if currentBranch(t, worktree) != "feature-two" {
		t.Errorf("expected branch feature-two")
	}
}

func TestCreateDuplicate(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "feature-one")
	defaultGlobalConfig(t)
	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})

	mustWtw(t, repo, "create", "feature-one")
	_, err := wtw(t, repo, "create", "feature-one")
	if err == nil {
		t.Error("expected duplicate worktree creation to fail")
	}
}

func TestSyncIgnoredFiles(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "feature-one")

	gitignore := ".env\n.env.local\nnode_modules/\ndist/\n"
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(gitignore), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repo, "add", ".gitignore")
	gitCmd(t, repo, "commit", "-m", "add gitignore")

	if err := os.WriteFile(filepath.Join(repo, ".env"), []byte("SECRET=abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".env.local"), []byte("OTHER=xyz"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "node_modules", "fake-pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "node_modules", "fake-pkg", "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "dist", "index.js"), []byte("built"), 0o644); err != nil {
		t.Fatal(err)
	}

	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    true,
		"sync_excludes":   []string{"node_modules"},
		"post_copy_hooks": []string{},
	})
	defaultGlobalConfig(t)

	mustWtw(t, repo, "create", "feature-one")

	worktree := filepath.Join(tmp, "test-project--worktrees", "test-project--feature-one")
	assertFileExists(t, filepath.Join(worktree, ".env"))
	assertFileContents(t, filepath.Join(worktree, ".env"), "SECRET=abc")
	assertFileExists(t, filepath.Join(worktree, ".env.local"))
	assertDirNotExists(t, filepath.Join(worktree, "node_modules"))
	assertFileExists(t, filepath.Join(worktree, "dist", "index.js"))
}

func TestNoSync(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "nosync-project")
	initRepo(t, repo, "no-sync-test")

	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(".env\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".env"), []byte("SECRET=123"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repo, "add", ".gitignore")
	gitCmd(t, repo, "commit", "-m", "add gitignore")

	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})
	defaultGlobalConfig(t)

	mustWtw(t, repo, "create", "no-sync-test")

	assertFileNotExists(t, filepath.Join(tmp, "nosync-project--worktrees", "nosync-project--no-sync-test", ".env"))
}

func TestHooks(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "bugfix-123")

	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":  false,
		"sync_excludes": []string{},
		"post_copy_hooks": []string{
			"touch hook-step-1.txt",
			"echo 'hello from hook' > hook-step-2.txt",
			"pwd > hook-step-3.txt",
		},
	})
	defaultGlobalConfig(t)

	mustWtw(t, repo, "create", "bugfix-123")

	worktree := filepath.Join(tmp, "test-project--worktrees", "test-project--bugfix-123")
	assertFileExists(t, filepath.Join(worktree, "hook-step-1.txt"))
	assertFileContents(t, filepath.Join(worktree, "hook-step-2.txt"), "hello from hook")

	realWorktree := realPath(t, worktree)
	assertFileContents(t, filepath.Join(worktree, "hook-step-3.txt"), realWorktree)
}

func TestHooksFail(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "hook-fail-test")

	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":  false,
		"sync_excludes": []string{},
		"post_copy_hooks": []string{
			"touch before-fail.txt",
			"exit 1",
			"touch after-fail.txt",
		},
	})
	defaultGlobalConfig(t)

	wtw(t, repo, "create", "hook-fail-test") //nolint:errcheck

	worktree := filepath.Join(tmp, "test-project--worktrees", "test-project--hook-fail-test")
	assertFileExists(t, filepath.Join(worktree, "before-fail.txt"))
	assertFileNotExists(t, filepath.Join(worktree, "after-fail.txt"))
}

func TestRemoveRm(t *testing.T) {
	testRemove(t, "rm")
}

func TestRemoveAlias(t *testing.T) {
	testRemove(t, "remove")
}

func testRemove(t *testing.T, cmd string) {
	t.Helper()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "feature-to-rm")
	defaultGlobalConfig(t)
	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})

	mustWtw(t, repo, "create", "feature-to-rm")

	worktree := filepath.Join(tmp, "test-project--worktrees", "test-project--feature-to-rm")
	assertDirExists(t, worktree)

	mustWtw(t, repo, cmd, "feature-to-rm")
	assertDirNotExists(t, worktree)

	out := gitCmd(t, repo, "worktree", "list")
	if strings.Contains(out, "feature-to-rm") {
		t.Error("branch still in git worktree list after removal")
	}
}

func TestRemoveForceFlag(t *testing.T) {
	testRemoveForce(t, "-f")
}

func TestRemoveForceLongFlag(t *testing.T) {
	testRemoveForce(t, "--force")
}

func testRemoveForce(t *testing.T, flag string) {
	t.Helper()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	branch := "force-rm" + strings.ReplaceAll(flag, "-", "")
	initRepo(t, repo, branch)
	defaultGlobalConfig(t)
	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})

	mustWtw(t, repo, "create", branch)

	worktree := filepath.Join(tmp, "test-project--worktrees", "test-project--"+branch)
	if err := os.WriteFile(filepath.Join(worktree, "untracked.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := wtw(t, repo, "rm", branch)
	if err == nil {
		t.Error("expected removal of dirty worktree to fail without force")
	}
	assertOutputContains(t, out, "modified/untracked files")

	mustWtw(t, repo, "rm", flag, branch)
	assertDirNotExists(t, worktree)
}

func TestGlobalForceRm(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "feature-one")
	defaultGlobalConfig(t)
	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})

	mustWtw(t, repo, "create", "feature-one")

	worktree := filepath.Join(tmp, "test-project--worktrees", "test-project--feature-one")
	if err := os.WriteFile(filepath.Join(worktree, "untracked.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := mustWtw(t, repo, "--force", "rm", "feature-one")
	if strings.Contains(out, "Force remove") {
		t.Error("should not prompt with --force")
	}
	assertDirNotExists(t, worktree)
}

func TestRemoveNonExistent(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo)
	defaultGlobalConfig(t)

	_, err := wtw(t, repo, "rm", "nonexistent-branch")
	if err == nil {
		t.Error("expected removal of nonexistent worktree to fail")
	}
}

func TestCustomNaming(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "custom-naming-test")

	writeGlobalConfig(t, `{"editor":"echo","auto_open_editor":false,"naming":{"worktree_dir_suffix":"-wt","branch_separator":"_"}}`)
	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})

	mustWtw(t, repo, "create", "custom-naming-test")

	assertDirExists(t, filepath.Join(tmp, "test-project-wt", "test-project_custom-naming-test"))
}

func TestRemoveCustomNamingRm(t *testing.T) {
	testRemoveCustomNaming(t, "rm", "rm-custom-naming-test")
}

func TestRemoveCustomNamingAlias(t *testing.T) {
	testRemoveCustomNaming(t, "remove", "remove-custom-naming-test")
}

func testRemoveCustomNaming(t *testing.T, cmd, branch string) {
	t.Helper()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, branch)

	writeGlobalConfig(t, `{"editor":"echo","auto_open_editor":false,"naming":{"worktree_dir_suffix":"-wt","branch_separator":"_"}}`)
	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})

	mustWtw(t, repo, "create", branch)

	worktree := filepath.Join(tmp, "test-project-wt", "test-project_"+branch)
	assertDirExists(t, worktree)

	mustWtw(t, repo, cmd, branch)
	assertDirNotExists(t, worktree)

	out := gitCmd(t, repo, "worktree", "list")
	if strings.Contains(out, branch) {
		t.Errorf("branch %s still in git worktree list", branch)
	}
}

func TestRemoveForceCustomNamingShortFlag(t *testing.T) {
	testRemoveForceCustomNaming(t, "-f", "force-rm-custom-f")
}

func TestRemoveForceCustomNamingLongFlag(t *testing.T) {
	testRemoveForceCustomNaming(t, "--force", "force-rm-custom-double")
}

func testRemoveForceCustomNaming(t *testing.T, flag, branch string) {
	t.Helper()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, branch)

	writeGlobalConfig(t, `{"editor":"echo","auto_open_editor":false,"naming":{"worktree_dir_suffix":"-wt","branch_separator":"_"}}`)
	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})

	mustWtw(t, repo, "create", branch)

	worktree := filepath.Join(tmp, "test-project-wt", "test-project_"+branch)
	assertDirExists(t, worktree)

	if err := os.WriteFile(filepath.Join(worktree, "untracked.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := wtw(t, repo, "rm", branch)
	if err == nil {
		t.Error("expected failure removing dirty worktree")
	}
	assertOutputContains(t, out, "modified/untracked files")

	mustWtw(t, repo, "rm", flag, branch)
	assertDirNotExists(t, worktree)
}

func TestListShowsWorktrees(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo, "feature-one", "feature-two")
	defaultGlobalConfig(t)
	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    false,
		"sync_excludes":   []string{},
		"post_copy_hooks": []string{},
	})

	mustWtw(t, repo, "create", "feature-one")
	mustWtw(t, repo, "create", "feature-two")

	out := gitCmd(t, repo, "worktree", "list")
	assertOutputContains(t, out, "feature-one")
	assertOutputContains(t, out, "feature-two")
}

func TestListHelp(t *testing.T) {
	for _, alias := range []string{"ls", "list"} {
		t.Run(alias, func(t *testing.T) {
			tmp := t.TempDir()
			repo := filepath.Join(tmp, "test-project")
			initRepo(t, repo)
			defaultGlobalConfig(t)

			out, err := wtw(t, repo, alias, "--help")
			if err != nil {
				t.Fatalf("wtw %s --help failed: %v\n%s", alias, err, out)
			}
		})
	}
}

func TestNotGitRepo(t *testing.T) {
	tmp := t.TempDir()

	out, err := wtw(t, tmp, "create", "some-branch")
	if err == nil {
		t.Error("expected failure outside git repo")
	}
	assertOutputContains(t, out, "not a git repository")

	out, err = wtw(t, tmp, "ls")
	if err == nil {
		t.Error("expected failure outside git repo")
	}
	assertOutputContains(t, out, "not a git repository")
}

func TestHelp(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo)

	out := mustWtw(t, repo, "--help")
	for _, s := range []string{"create", "ls", "rm", "init"} {
		assertOutputContains(t, out, s)
	}

	out = mustWtw(t, repo, "create", "--help")
	assertOutputContains(t, out, "branch")

	out = mustWtw(t, repo, "rm", "--help")
	assertOutputContains(t, out, "force")

	out = mustWtw(t, repo, "init", "--help")
	assertOutputContains(t, out, "local")
}

func TestHelpRemoveAlias(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "test-project")
	initRepo(t, repo)

	out := mustWtw(t, repo, "remove", "--help")
	assertOutputContains(t, out, "force")
}

func TestCreateNewBranchNonInteractiveFails(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "newbranch-project")
	initRepo(t, repo)
	defaultGlobalConfig(t)

	out, err := wtw(t, repo, "create", "nonexistent-branch")
	if err == nil {
		t.Error("expected failure for nonexistent branch in non-interactive mode")
	}
	assertOutputContains(t, out, "does not exist")
}

func TestGlobalForceCreate(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "force-create-project")
	initRepo(t, repo)
	defaultGlobalConfig(t)

	out := mustWtw(t, repo, "--force", "create", "new-branch-from-force")
	if strings.Contains(out, "Create it") {
		t.Error("should not prompt with --force")
	}
	assertOutputContains(t, out, "Done")

	worktree := filepath.Join(tmp, "force-create-project--worktrees", "force-create-project--new-branch-from-force")
	assertDirExists(t, worktree)

	branchList := gitCmd(t, repo, "branch")
	if !strings.Contains(branchList, "new-branch-from-force") {
		t.Error("branch was not created")
	}
}

func TestDefaults(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "bare-project")
	initRepo(t, repo, "test-defaults")

	writeGlobalConfig(t, `{"auto_open_editor":false}`)

	mustWtw(t, repo, "create", "test-defaults")

	assertDirExists(t, filepath.Join(tmp, "bare-project--worktrees", "bare-project--test-defaults"))
}

func TestInitNaming(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "init-naming-project")
	initRepo(t, repo, "test-branch")

	writeGlobalConfig(t, `{"editor":"echo","auto_open_editor":false,"naming":{"worktree_dir_suffix":"-trees","branch_separator":"."}}`)

	mustWtw(t, repo, "create", "test-branch")

	assertDirExists(t, filepath.Join(tmp, "init-naming-project-trees", "init-naming-project.test-branch"))

	cfgPath := filepath.Join(os.Getenv("HOME"), ".config", "worktree-workflow", "config.json")
	assertFileContains(t, cfgPath, "worktree_dir_suffix")
	assertFileContains(t, cfgPath, "branch_separator")

	defaultGlobalConfig(t)
}

func TestInitLocalOutsideRepo(t *testing.T) {
	tmp := t.TempDir()

	for _, args := range [][]string{
		{"init", "--local"},
		{"init", "--local", "-f"},
		{"init", "--local", "--force"},
	} {
		out, err := wtw(t, tmp, args...)
		if err == nil {
			t.Errorf("expected failure for %v outside git repo", args)
		}
		assertOutputContains(t, out, "not a git repository")
	}
}

func TestInitLocalForce(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "initlocal-force-project")
	initRepo(t, repo, "init-local-test")
	defaultGlobalConfig(t)

	cfgFile := filepath.Join(repo, ".worktree-workflow.json")
	if _, err := os.Stat(cfgFile); err == nil {
		t.Fatal("config file should not exist before init")
	}

	mustWtw(t, repo, "init", "--local", "-f")

	assertFileExists(t, cfgFile)
	assertFileContains(t, cfgFile, `"sync_ignored"`)
	assertFileContains(t, cfgFile, `"sync_excludes"`)
	assertFileContains(t, cfgFile, `"post_copy_hooks"`)

	mustWtw(t, repo, "create", "init-local-test")
	assertDirExists(t, filepath.Join(tmp, "initlocal-force-project--worktrees", "initlocal-force-project--init-local-test"))
}

func TestInitLocalConfigSyncAndHooks(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "initlocal-project")
	initRepo(t, repo, "test-branch")
	defaultGlobalConfig(t)

	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(".env\nnode_modules/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repo, "add", ".gitignore")
	gitCmd(t, repo, "commit", "-m", "add gitignore")

	if err := os.WriteFile(filepath.Join(repo, ".env"), []byte("SECRET=123"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "node_modules", "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "node_modules", "pkg", "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	writeProjectConfig(t, repo, map[string]any{
		"sync_ignored":    true,
		"sync_excludes":   []string{"node_modules"},
		"post_copy_hooks": []string{"touch .initialized"},
	})

	mustWtw(t, repo, "create", "test-branch")

	worktree := filepath.Join(tmp, "initlocal-project--worktrees", "initlocal-project--test-branch")
	assertDirExists(t, worktree)
	assertFileExists(t, filepath.Join(worktree, ".env"))
	assertDirNotExists(t, filepath.Join(worktree, "node_modules"))
	assertFileExists(t, filepath.Join(worktree, ".initialized"))
}

func TestWindowsRejected(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	tmp := t.TempDir()
	out, err := wtw(t, tmp, "create", "branch")
	if err == nil {
		t.Error("expected failure on windows")
	}
	assertOutputContains(t, out, "Windows")
}
