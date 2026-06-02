//go:build !online

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skillshare/internal/testutil"
)

// setupExtrasOneTarget is a helper that creates a sandbox with a minimal global
// config and one extra ("rules") pointing to dir1. It returns the sandbox and
// the path of the first target directory.
func setupExtrasOneTarget(t *testing.T) (*testutil.Sandbox, string) {
	t.Helper()
	sb := testutil.NewSandbox(t)

	claudeTarget := sb.CreateTarget("claude")
	dir1 := filepath.Join(sb.Home, "extra-target-1")
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatal(err)
	}

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets:
  claude:
    path: ` + claudeTarget + `
extras:
  - name: rules
    targets:
      - path: ` + dir1 + `
`)
	return sb, dir1
}

// TestExtrasTarget_AddTarget_AddsSecondTarget verifies that --add-target appends
// a new entry to the extra's targets list and persists it in the config.
func TestExtrasTarget_AddTarget_AddsSecondTarget(t *testing.T) {
	sb, dir1 := setupExtrasOneTarget(t)
	defer sb.Cleanup()

	dir2 := filepath.Join(sb.Home, "extra-target-2")
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}

	result := sb.RunCLI("extras", "rules", "--add-target", dir2, "--mode", "copy", "-g")
	result.AssertSuccess(t)

	configContent := sb.ReadFile(sb.ConfigPath)
	if !strings.Contains(configContent, dir1) {
		t.Errorf("config should still contain first target %s:\n%s", dir1, configContent)
	}
	if !strings.Contains(configContent, dir2) {
		t.Errorf("config should now contain second target %s:\n%s", dir2, configContent)
	}
}

// TestExtrasTarget_AddTarget_DuplicateErrors verifies that adding a target path
// that already exists on the extra returns a non-zero exit and reports the
// duplicate.
func TestExtrasTarget_AddTarget_DuplicateErrors(t *testing.T) {
	sb, dir1 := setupExtrasOneTarget(t)
	defer sb.Cleanup()

	result := sb.RunCLI("extras", "rules", "--add-target", dir1, "-g")
	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "already exists")
}

// TestExtrasTarget_RemoveTarget_RemovesOne verifies that --remove-target strips
// the specified entry from the config while preserving the remaining targets.
func TestExtrasTarget_RemoveTarget_RemovesOne(t *testing.T) {
	sb, dir1 := setupExtrasOneTarget(t)
	defer sb.Cleanup()

	dir2 := filepath.Join(sb.Home, "extra-target-2")
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}

	// First add a second target so we have two.
	addResult := sb.RunCLI("extras", "rules", "--add-target", dir2, "-g")
	addResult.AssertSuccess(t)

	// Now remove the second target.
	rmResult := sb.RunCLI("extras", "rules", "--remove-target", dir2, "-g")
	rmResult.AssertSuccess(t)

	configContent := sb.ReadFile(sb.ConfigPath)
	if !strings.Contains(configContent, dir1) {
		t.Errorf("config should still contain first target %s:\n%s", dir1, configContent)
	}
	if strings.Contains(configContent, dir2) {
		t.Errorf("config should no longer contain second target %s:\n%s", dir2, configContent)
	}
}

// TestExtrasTarget_RemoveTarget_LastTargetErrors verifies that attempting to
// remove the sole remaining target returns a non-zero exit and reports the
// "last target" constraint.
func TestExtrasTarget_RemoveTarget_LastTargetErrors(t *testing.T) {
	sb, dir1 := setupExtrasOneTarget(t)
	defer sb.Cleanup()

	result := sb.RunCLI("extras", "rules", "--remove-target", dir1, "-g")
	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "last target")
}

// TestExtrasTarget_ModeSubcommandRemoved verifies that "extras mode" is no
// longer a valid subcommand. When called as "extras mode --mode copy", the word
// "mode" is treated as an extra name; since no extra with that name exists, the
// command must fail.
func TestExtrasTarget_ModeSubcommandRemoved(t *testing.T) {
	sb, _ := setupExtrasOneTarget(t)
	defer sb.Cleanup()

	result := sb.RunCLI("extras", "mode", "--mode", "copy", "-g")
	result.AssertFailure(t)
}
