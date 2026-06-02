package main

import (
	"fmt"
	"os"
	"time"

	"skillshare/internal/config"
	"skillshare/internal/oplog"
	"skillshare/internal/sync"
	"skillshare/internal/ui"
)

// cmdExtrasRemoveTarget handles `skillshare extras <name> --remove-target <path>
// [--prune]`. It removes one target from an existing extra. By default only the
// config entry is removed (synced files are left on disk); --prune additionally
// deletes the skillshare-managed files under that target.
func cmdExtrasRemoveTarget(args []string) error {
	start := time.Now()

	mode, rest, err := parseModeArgs(args)
	if err != nil {
		return err
	}

	cwd, _ := os.Getwd()
	if mode == modeAuto {
		if projectConfigExists(cwd) {
			mode = modeProject
		} else {
			mode = modeGlobal
		}
	}
	applyModeLabel(mode)

	var name, rmPath string
	var prune bool
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--remove-target":
			if i+1 >= len(rest) {
				return fmt.Errorf("--remove-target requires a path argument")
			}
			i++
			rmPath = rest[i]
		case "--prune":
			prune = true
		case "--help", "-h":
			printExtrasRemoveTargetHelp()
			return nil
		default:
			if name == "" {
				name = rest[i]
			} else {
				return fmt.Errorf("unexpected argument: %s", rest[i])
			}
		}
	}

	if name == "" {
		return fmt.Errorf("extras name is required: skillshare extras <name> --remove-target <path>")
	}
	if rmPath == "" {
		return fmt.Errorf("--remove-target requires a path argument")
	}

	var extras []config.ExtraConfig
	var configPath string
	var saveFn func() error
	if mode == modeProject {
		projCfg, loadErr := config.LoadProject(cwd)
		if loadErr != nil {
			return loadErr
		}
		extras = projCfg.Extras
		configPath = config.ProjectConfigPath(cwd)
		saveFn = func() error { return projCfg.Save(cwd) }
	} else {
		cfg, loadErr := config.Load()
		if loadErr != nil {
			return loadErr
		}
		extras = cfg.Extras
		configPath = config.ConfigPath()
		saveFn = cfg.Save
	}

	idx := -1
	for i := range extras {
		if extras[i].Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("extra %q not found", name)
	}

	tIdx := -1
	var targetMode string
	for j, t := range extras[idx].Targets {
		if t.Path == rmPath {
			tIdx = j
			targetMode = t.Mode
			break
		}
	}
	if tIdx == -1 {
		return fmt.Errorf("target %q not found in extra %q", rmPath, name)
	}
	if len(extras[idx].Targets) == 1 {
		return fmt.Errorf("%q is the last target of %q — use 'skillshare extras remove %s' to remove the whole extra", rmPath, name, name)
	}

	// Resolve the on-disk target path for optional pruning before mutating config.
	var resolved string
	if mode == modeProject {
		resolved = resolveProjectPath(cwd, rmPath)
	} else {
		resolved = config.ExpandPath(rmPath)
	}

	extras[idx].Targets = append(extras[idx].Targets[:tIdx], extras[idx].Targets[tIdx+1:]...)

	if err := saveFn(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.Success("Removed target %s from %s", shortenPath(rmPath), name)

	if prune {
		pruned, errs := sync.PruneExtraTarget(resolved, targetMode)
		for _, msg := range errs {
			ui.Warning("%s", msg)
		}
		ui.Info("Pruned %d file(s) from %s", pruned, shortenPath(rmPath))
	} else {
		ui.Info("Synced files left in place. Run 'skillshare sync extras%s' to clean up orphaned links, or re-run with --prune.", projectSuffix(mode))
	}

	e := oplog.NewEntry("extras-target", "ok", time.Since(start))
	e.Args = map[string]any{"name": name, "target": rmPath, "action": "remove", "prune": prune}
	oplog.WriteWithLimit(configPath, oplog.OpsFile, e, logMaxEntries()) //nolint:errcheck

	return nil
}

func printExtrasRemoveTargetHelp() {
	fmt.Println(`Usage: skillshare extras <name> --remove-target <path> [--prune]

Remove one target from an existing extra. By default only the config entry is
removed; synced files are left on disk. Use --prune to also delete the
skillshare-managed files under that target.

Options:
  --remove-target <path>  Target directory to remove (required)
  --prune                 Also delete skillshare-managed files under that target
  --project, -p           Use project mode (.skillshare/)
  --global, -g            Use global mode (~/.config/skillshare/)
  --help, -h              Show this help

Examples:
  skillshare extras rules --remove-target ~/.cursor/rules
  skillshare extras rules --remove-target ~/.cursor/rules --prune`)
}
