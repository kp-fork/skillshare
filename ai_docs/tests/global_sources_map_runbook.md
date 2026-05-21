# CLI E2E Runbook: Global Config — sources map ↔ legacy fields interop

Validates v0.19.16's `sources:` map for global config alongside the legacy
`source:` / `agents_source:` / `extras_source:` top-level fields. Confirms
both formats produce identical CLI behavior, mixed configs prefer the new
format, and fresh `init -g` emits the new shape.

**Origin**: v0.19.16 global config sources map feature.

## Scope

- Fresh `init -g` writes only the new `sources:` map (no legacy fields)
- Legacy-only config (`source:` / `agents_source:`) continues to work
- New-format config (`sources.skills:` / `sources.agents:`) works
- Mixed config: `sources.skills` wins over the corresponding legacy field
- `ss list` / `ss sync` resolve paths via `EffectiveSkillsSource()` /
  `EffectiveAgentsSource()` regardless of format
- `ss sync` validation rejects configs with neither `source` nor `sources.skills`

## Environment

Run inside devcontainer. mdproof setup auto-runs
`ss init -g --force --all-targets --no-git --no-skill` before each step, so
each step starts with a fresh default config in the new format.

## Steps

### 1. Fresh init -g writes the new sources map and omits legacy fields

```bash
config=~/.config/skillshare/config.yaml
echo "sources_map=$(grep -c '^sources:' $config)"
echo "skills_subkey=$(grep -c '^  skills:' $config)"
echo "agents_subkey=$(grep -c '^  agents:' $config)"
echo "legacy_source=$(grep -c '^source:' $config)"
echo "legacy_agents_source=$(grep -c '^agents_source:' $config)"
echo "legacy_extras_source=$(grep -c '^extras_source:' $config)"
```

Expected:
- exit_code: 0
- sources_map=1
- skills_subkey=1
- agents_subkey=1
- legacy_source=0
- legacy_agents_source=0
- legacy_extras_source=0

### 2. Legacy-only config — ss list resolves skills via top-level source

```bash
SRC=/tmp/sources-runbook-legacy-skills
AGENTS=/tmp/sources-runbook-legacy-agents
rm -rf $SRC $AGENTS
mkdir -p $SRC/test-legacy $AGENTS

cat > ~/.config/skillshare/config.yaml <<EOF
# yaml-language-server: \$schema=https://raw.githubusercontent.com/runkids/skillshare/main/schemas/config.schema.json
source: $SRC
agents_source: $AGENTS
mode: merge
targets:
  claude:
    skills:
      path: ~/.claude/skills
EOF

cat > $SRC/test-legacy/SKILL.md <<'SKILL'
---
name: test-legacy
description: Skill placed under legacy source path
---
# Body
SKILL

ss list --json
```

Expected:
- exit_code: 0
- jq: [.[].name] | contains(["test-legacy"])

### 3. New sources map — ss list resolves skills via sources.skills

```bash
SRC=/tmp/sources-runbook-new-skills
AGENTS=/tmp/sources-runbook-new-agents
rm -rf $SRC $AGENTS
mkdir -p $SRC/test-new $AGENTS

cat > ~/.config/skillshare/config.yaml <<EOF
# yaml-language-server: \$schema=https://raw.githubusercontent.com/runkids/skillshare/main/schemas/config.schema.json
sources:
  skills: $SRC
  agents: $AGENTS
mode: merge
targets:
  claude:
    skills:
      path: ~/.claude/skills
EOF

cat > $SRC/test-new/SKILL.md <<'SKILL'
---
name: test-new
description: Skill placed under sources.skills path
---
# Body
SKILL

ss list --json
```

Expected:
- exit_code: 0
- jq: [.[].name] | contains(["test-new"])

### 4. Mixed config — sources.skills wins over legacy source

```bash
WIN=/tmp/sources-runbook-win
LOSE=/tmp/sources-runbook-lose
AGENTS=/tmp/sources-runbook-mixed-agents
rm -rf $WIN $LOSE $AGENTS
mkdir -p $WIN/winning $LOSE/losing $AGENTS

cat > ~/.config/skillshare/config.yaml <<EOF
# yaml-language-server: \$schema=https://raw.githubusercontent.com/runkids/skillshare/main/schemas/config.schema.json
source: $LOSE
agents_source: $AGENTS
sources:
  skills: $WIN
mode: merge
targets:
  claude:
    skills:
      path: ~/.claude/skills
EOF

cat > $WIN/winning/SKILL.md <<'SKILL'
---
name: winning
description: Skill under sources.skills (should appear)
---
# Body
SKILL

cat > $LOSE/losing/SKILL.md <<'SKILL'
---
name: losing
description: Skill under legacy source (should NOT appear)
---
# Body
SKILL

ss list --json
```

Expected:
- exit_code: 0
- jq: [.[].name] | contains(["winning"])
- jq: ([.[].name] | contains(["losing"])) | not

### 5. Sync works with new sources map and a custom target path

```bash
SRC=/tmp/sources-runbook-sync-skills
AGENTS=/tmp/sources-runbook-sync-agents
TARGET=/tmp/sources-runbook-sync-target
rm -rf $SRC $AGENTS $TARGET
mkdir -p $SRC/sync-skill $AGENTS $TARGET

cat > ~/.config/skillshare/config.yaml <<EOF
# yaml-language-server: \$schema=https://raw.githubusercontent.com/runkids/skillshare/main/schemas/config.schema.json
sources:
  skills: $SRC
  agents: $AGENTS
mode: merge
targets:
  claude:
    skills:
      path: $TARGET
EOF

cat > $SRC/sync-skill/SKILL.md <<'SKILL'
---
name: sync-skill
description: Verify sync writes to target when sources.skills is set
---
# Body
SKILL

ss sync --json
```

Expected:
- exit_code: 0
- jq: .linked == 1
- jq: .details[0].name == "claude"

### 6. Sync rejects configs that set neither source nor sources.skills

```bash
AGENTS=/tmp/sources-runbook-empty-agents
rm -rf $AGENTS
mkdir -p $AGENTS

# Config with NEITHER legacy `source:` NOR `sources.skills:`
cat > ~/.config/skillshare/config.yaml <<EOF
# yaml-language-server: \$schema=https://raw.githubusercontent.com/runkids/skillshare/main/schemas/config.schema.json
agents_source: $AGENTS
mode: merge
targets:
  claude:
    skills:
      path: ~/.claude/skills
EOF

ss sync 2>&1
EXIT=$?
echo "exit_status=$EXIT"
```

Expected:
- exit_status=1
- source path is empty

## Pass Criteria

All six steps pass. Together they prove that the global config sources map
shipped in v0.19.16 is wired through the runtime helpers, the legacy
top-level fields still work unchanged, mixed configs prefer the new keys,
and the validation surface still rejects configs that explicitly omit both
shapes.
