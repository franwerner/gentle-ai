package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/franwerner/gentle-ai/v3/internal/backup"
	"github.com/franwerner/gentle-ai/v3/internal/components/communitytool"
	"github.com/franwerner/gentle-ai/v3/internal/components/openrecord"
	"github.com/franwerner/gentle-ai/v3/internal/model"
	"github.com/franwerner/gentle-ai/v3/internal/planner"
	"github.com/franwerner/gentle-ai/v3/internal/state"
	"github.com/franwerner/gentle-ai/v3/internal/verify"
)

// stageOpenRecordEmit replaces runCommand and cmdLookPath for the duration of
// a test: every invocation is recorded, and `openrecord skills --emit <dir>`
// writes the one skill plus the manifest the real binary would, so the real
// fan-out runs against it.
func stageOpenRecordEmit(t *testing.T, binaryOnPath bool) *[][]string {
	t.Helper()
	previousRun, previousLookPath := runCommand, cmdLookPath
	t.Cleanup(func() { runCommand, cmdLookPath = previousRun, previousLookPath })

	calls := &[][]string{}
	cmdLookPath = func(name string) (string, error) {
		if name == "openrecord" && !binaryOnPath {
			return "", os.ErrNotExist
		}
		return filepath.Join("/usr/bin", name), nil
	}
	runCommand = func(name string, args ...string) error {
		*calls = append(*calls, append([]string{name}, args...))
		if name == "openrecord" && len(args) >= 3 && args[0] == "skills" && args[1] == "--emit" {
			staging := args[2]
			if err := os.MkdirAll(filepath.Join(staging, "openrecord-consult"), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(staging, "openrecord-consult", "SKILL.md"), []byte("consult skill body"), 0o644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(staging, ".openrecord-emitted.json"),
				[]byte(`{"version":1,"emitter":"openrecord@test","with_qmd":true,"entries":[{"path":"openrecord-consult/SKILL.md","hash":"h"}]}`), 0o644)
		}
		return nil
	}
	return calls
}

// neutralizeOpenRecordArms makes the openrecord component a no-op and returns
// the restore func. openrecord now ships in every non-custom preset, so any
// test that runs an install or a sync reaches its arms even when the tool is
// nothing to do with what the test is about: the sync arm probes the machine
// through cmdLookPath/runCommand and then emits and fans five skills out into
// the test's home, and the install arm goes to the network when the binary is
// missing. Substituting the availability probe alongside the two seams is what
// takes the count to zero — the seams alone leave `openrecord version` behind,
// because the arm runs that probe before it reaches them.
//
// Nothing else in this package's sync path reaches runCommand or cmdLookPath,
// so the substitution is invisible to whatever the calling test asserts.
func neutralizeOpenRecordArms() func() {
	// Only the two arms. The availability probe lives behind the seam now, so
	// substituting them neutralizes openrecord completely — touching
	// runCommand or cmdLookPath here would instead overwrite whatever the
	// calling test set up for its own subject.
	previousInstall, previousSync := installOpenRecordWithHome, syncOpenRecordWithHome

	installOpenRecordWithHome = func(_ string, agents []model.AgentID, _ communitytool.Runner, _ communitytool.Detector) (openrecord.InstallResult, error) {
		return openrecord.InstallResult{FannedOut: len(agents)}, nil
	}
	syncOpenRecordWithHome = func(_ string, agents []model.AgentID, _ communitytool.Runner, _ communitytool.Detector) (openrecord.InstallResult, error) {
		return openrecord.InstallResult{FannedOut: len(agents)}, nil
	}

	return func() {
		installOpenRecordWithHome, syncOpenRecordWithHome = previousInstall, previousSync
	}
}

// stubOpenRecordArms is neutralizeOpenRecordArms scoped to one test.
func stubOpenRecordArms(t *testing.T) {
	t.Helper()
	t.Cleanup(neutralizeOpenRecordArms())
}

func ranCommand(calls [][]string, prefix ...string) bool {
	for _, call := range calls {
		if len(call) >= len(prefix) && slices.Equal(call[:len(prefix)], prefix) {
			return true
		}
	}
	return false
}

// TestComponentApplyStepInstallsOpenRecord pins the install arm's whole
// sequence: qmd install, emit, and the fan-out into each selected agent's
// skills directory.
func TestComponentApplyStepInstallsOpenRecord(t *testing.T) {
	home := t.TempDir()
	calls := stageOpenRecordEmit(t, true)
	state := &runtimeState{}

	step := componentApplyStep{
		component: model.ComponentOpenRecord,
		homeDir:   home,
		agents:    []model.AgentID{model.AgentClaudeCode, model.AgentGeminiCLI},
		state:     state,
	}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !ranCommand(*calls, "openrecord", "qmd", "install") {
		t.Fatalf("install arm did not run `openrecord qmd install`: %v", *calls)
	}
	if !ranCommand(*calls, "openrecord", "skills", "--emit") {
		t.Fatalf("install arm did not emit skills: %v", *calls)
	}
	for _, skillsDir := range []string{
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".gemini", "skills"),
	} {
		if _, err := os.Stat(filepath.Join(skillsDir, "openrecord-consult", "SKILL.md")); err != nil {
			t.Errorf("skills were not fanned out to %s: %v", skillsDir, err)
		}
		if _, err := os.Stat(filepath.Join(skillsDir, ".openrecord-emitted.json")); err != nil {
			t.Errorf("manifest was not fanned out to %s: %v", skillsDir, err)
		}
	}
	if !state.openRecordResolved || state.openRecordFannedOut != 2 {
		t.Fatalf("runtimeState openrecord outcome = (%t, %d), want (true, 2)", state.openRecordResolved, state.openRecordFannedOut)
	}
}

// TestComponentApplyStepSkipsOpenRecordBinaryWhenPresent pins the "install the
// binary only when missing" half of the sequence.
func TestComponentApplyStepSkipsOpenRecordBinaryWhenPresent(t *testing.T) {
	for _, tt := range []struct {
		name         string
		binaryOnPath bool
		wantInstall  bool
	}{
		{name: "binary on PATH skips the install", binaryOnPath: true, wantInstall: false},
		{name: "binary missing installs it", binaryOnPath: false, wantInstall: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := stageOpenRecordEmit(t, tt.binaryOnPath)
			step := componentApplyStep{
				component: model.ComponentOpenRecord,
				homeDir:   t.TempDir(),
				agents:    []model.AgentID{model.AgentClaudeCode},
			}
			if err := step.Run(); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if got := ranCommand(*calls, "go", "install"); got != tt.wantInstall {
				t.Fatalf("binary install ran = %t, want %t: %v", got, tt.wantInstall, *calls)
			}
		})
	}
}

// TestComponentApplyStepOpenRecordToleratesNoSkillsDir pins that a zero
// fan-out is a warning, not a failure: Pi exposes no skills directory.
func TestComponentApplyStepOpenRecordToleratesNoSkillsDir(t *testing.T) {
	stageOpenRecordEmit(t, true)
	state := &runtimeState{}
	step := componentApplyStep{
		component: model.ComponentOpenRecord,
		homeDir:   t.TempDir(),
		agents:    []model.AgentID{model.AgentPi},
		state:     state,
	}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error = %v, want nil — a zero fan-out is non-fatal", err)
	}
	if state.openRecordFannedOut != 0 {
		t.Fatalf("openRecordFannedOut = %d, want 0", state.openRecordFannedOut)
	}
}

// TestComponentSyncStepOpenRecordIsInjectOnly pins the engram precedent for
// sync: re-emit and fan out, never install the binary and never run qmd setup.
func TestComponentSyncStepOpenRecordIsInjectOnly(t *testing.T) {
	home := t.TempDir()
	calls := stageOpenRecordEmit(t, true)

	step := componentSyncStep{
		component: model.ComponentOpenRecord,
		homeDir:   home,
		agents:    []model.AgentID{model.AgentClaudeCode},
	}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !ranCommand(*calls, "openrecord", "skills", "--emit") {
		t.Fatalf("sync did not re-emit skills: %v", *calls)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "openrecord-consult", "SKILL.md")); err != nil {
		t.Errorf("sync did not fan out skills: %v", err)
	}
	for _, call := range *calls {
		joined := strings.Join(call, " ")
		if strings.HasPrefix(joined, "go install") || strings.HasPrefix(joined, "sh -c") {
			t.Errorf("sync installed the openrecord binary: %q", joined)
		}
		if strings.HasPrefix(joined, "openrecord qmd") {
			t.Errorf("sync ran qmd setup: %q", joined)
		}
	}
}

// TestRestorePersistedSelectionReaddsOpenRecord pins the existing-user
// migration: RestorePersistedSelection overwrites the freshly built component
// list with the persisted one, so every state.json written before openrecord
// became a component would otherwise lack it forever — the same class of bug
// as #3430's missing SDD component.
func TestRestorePersistedSelectionReaddsOpenRecord(t *testing.T) {
	selection := model.Selection{Components: []model.ComponentID{model.ComponentEngram, model.ComponentOpenRecord}}
	persisted := state.InstallState{
		SelectionConfigured: true,
		Components:          []model.ComponentID{model.ComponentEngram, model.ComponentSDD},
	}

	RestorePersistedSelection(&selection, persisted, SyncFlags{})

	if !selection.HasComponent(model.ComponentOpenRecord) {
		t.Fatalf("Components = %v, want openrecord re-added to a persisted selection that predates it", selection.Components)
	}
	for _, want := range []model.ComponentID{model.ComponentEngram, model.ComponentSDD} {
		if !selection.HasComponent(want) {
			t.Errorf("Components = %v, want the persisted %q preserved", selection.Components, want)
		}
	}
}

// TestNormalizeComponentsAcceptsOpenRecord pins the allow-list half of the
// catalog entry: `--component openrecord` must be a valid explicit selection.
func TestNormalizeComponentsAcceptsOpenRecord(t *testing.T) {
	got, err := normalizeComponents([]string{"openrecord"}, model.PresetCustom, model.PersonaGentleman)
	if err != nil {
		t.Fatalf("normalizeComponents(openrecord) error = %v", err)
	}
	if !slices.Equal(got, []model.ComponentID{model.ComponentOpenRecord}) {
		t.Fatalf("normalizeComponents(openrecord) = %v, want [openrecord]", got)
	}
}

// TestComponentSyncStepOpenRecordDegradesWhenBinaryMissing pins that sync
// refreshes what it can instead of failing, the way the engram arm does.
// gentle-ai embeds engram's protocol asset, so that arm never shells out;
// openrecord's skills are produced by the openrecord binary and deliberately
// not embedded, so without it there is simply nothing to copy. Reporting a
// broken install is the doctor's job, and openrecord is in its coreTools.
func TestComponentSyncStepOpenRecordDegradesWhenBinaryMissing(t *testing.T) {
	home := t.TempDir()
	calls := stageOpenRecordEmit(t, false)

	step := componentSyncStep{
		component: model.ComponentOpenRecord,
		homeDir:   home,
		agents:    []model.AgentID{model.AgentClaudeCode},
	}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error = %v, want the sync to degrade rather than fail", err)
	}
	if ranCommand(*calls, "openrecord", "skills", "--emit") {
		t.Fatalf("sync emitted skills despite the binary being absent: %v", *calls)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills")); !os.IsNotExist(err) {
		t.Errorf("sync wrote into the skills dir despite the binary being absent: err = %v", err)
	}
}

// TestBuildSyncSelectionCarriesOpenRecord pins that the base sync selection
// reaches the component even with no persisted selection to restore.
func TestBuildSyncSelectionCarriesOpenRecord(t *testing.T) {
	selection := BuildSyncSelection(SyncFlags{}, []model.AgentID{model.AgentClaudeCode})
	if !selection.HasComponent(model.ComponentOpenRecord) {
		t.Fatalf("BuildSyncSelection() components = %v, want openrecord present", selection.Components)
	}
}

// ─── Path table, backup and rollback ───────────────────────────────────────
//
// openrecord writes files no other component owns, so a missing entry in
// componentPathsWithWorkspaceScoped is silent and has three consequences at
// once: backupTargets snapshots nothing (rollback cannot restore),
// runPostApplyVerification asserts nothing, and the uninstall/backup manifests
// omit the files. These tests pin all three.

// seedOpenRecordFanOut writes a previous emit into skillDir exactly as fanOut
// would: the listed skills plus the manifest that records them. It returns the
// manifest bytes so a test can assert the restored copy byte-for-byte.
func seedOpenRecordFanOut(t *testing.T, skillDir string, files map[string]string) []byte {
	t.Helper()
	entries := make([]string, 0, len(files))
	names := make([]string, 0, len(files))
	for relPath := range files {
		names = append(names, relPath)
	}
	slices.Sort(names)
	for _, relPath := range names {
		body := files[relPath]
		full := filepath.Join(skillDir, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", full, err)
		}
		sum := sha256.Sum256([]byte(body))
		entries = append(entries, fmt.Sprintf(`{"path":%q,"hash":%q}`, relPath, hex.EncodeToString(sum[:])))
	}
	manifest := []byte(fmt.Sprintf(`{"version":1,"emitter":"openrecord@previous","with_qmd":true,"entries":[%s]}`,
		strings.Join(entries, ",")))
	if err := os.WriteFile(filepath.Join(skillDir, ".openrecord-emitted.json"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest error = %v", err)
	}
	return manifest
}

func openRecordPlan(agents ...model.AgentID) (model.Selection, planner.ResolvedPlan) {
	selection := model.Selection{Agents: agents, Components: []model.ComponentID{model.ComponentOpenRecord}}
	return selection, planner.ResolvedPlan{Agents: agents, OrderedComponents: selection.Components}
}

// TestComponentPathsOpenRecordDeclaresManifestAndEmittedSkills pins the path
// table entry: every fanned-out file plus the manifest that records it, for
// each selected agent that exposes a skills directory.
func TestComponentPathsOpenRecordDeclaresManifestAndEmittedSkills(t *testing.T) {
	home := t.TempDir()
	claudeSkills := filepath.Join(home, ".claude", "skills")
	geminiSkills := filepath.Join(home, ".gemini", "skills")
	seedOpenRecordFanOut(t, claudeSkills, map[string]string{
		"openrecord-consult/SKILL.md": "consult",
		"openrecord-write/SKILL.md":   "write",
	})
	seedOpenRecordFanOut(t, geminiSkills, map[string]string{"openrecord-consult/SKILL.md": "consult"})

	selection, _ := openRecordPlan(model.AgentClaudeCode, model.AgentGeminiCLI)
	adapters := resolveAdapters(selection.Agents)
	paths := componentPathsWithWorkspace(home, "", selection, adapters, model.ComponentOpenRecord)

	for _, want := range []string{
		filepath.Join(claudeSkills, "openrecord-consult", "SKILL.md"),
		filepath.Join(claudeSkills, "openrecord-write", "SKILL.md"),
		filepath.Join(claudeSkills, ".openrecord-emitted.json"),
		filepath.Join(geminiSkills, "openrecord-consult", "SKILL.md"),
		filepath.Join(geminiSkills, ".openrecord-emitted.json"),
	} {
		if !containsPath(paths, want) {
			t.Errorf("componentPaths(openrecord) missing %q; paths=%v", want, paths)
		}
	}
	// Gemini's manifest lists one skill; the table must not leak Claude's second.
	if containsPath(paths, filepath.Join(geminiSkills, "openrecord-write", "SKILL.md")) {
		t.Errorf("componentPaths(openrecord) declared a path no manifest lists; paths=%v", paths)
	}
}

// TestComponentPathsOpenRecordSkipsAgentsWithoutSkillsDir pins the Pi case:
// the fan-out skips an agent whose SkillsDir is empty, so the path table must
// declare nothing for it rather than inventing a target verification would
// then fail on.
func TestComponentPathsOpenRecordSkipsAgentsWithoutSkillsDir(t *testing.T) {
	home := t.TempDir()
	selection, _ := openRecordPlan(model.AgentPi)
	adapters := resolveAdapters(selection.Agents)

	if paths := componentPathsWithWorkspace(home, "", selection, adapters, model.ComponentOpenRecord); len(paths) != 0 {
		t.Fatalf("componentPaths(openrecord) for Pi = %v, want none — Pi exposes no skills directory", paths)
	}
}

// TestComponentPathsOpenRecordIsEmptyBeforeFirstEmit pins the honest half of
// the manifest-driven contract: with no manifest on disk gentle-ai has not
// seen what openrecord ships, so it declares nothing rather than guessing a
// skill list post-apply verification would then require.
func TestComponentPathsOpenRecordIsEmptyBeforeFirstEmit(t *testing.T) {
	home := t.TempDir()
	selection, _ := openRecordPlan(model.AgentClaudeCode)
	adapters := resolveAdapters(selection.Agents)

	if paths := componentPathsWithWorkspace(home, "", selection, adapters, model.ComponentOpenRecord); len(paths) != 0 {
		t.Fatalf("componentPaths(openrecord) before the first emit = %v, want none", paths)
	}
}

// TestComponentPathsOpenRecordIgnoresWorkspaceScope pins that openrecord stays
// out of componentPathDirScoped's per-adapter injection list. Both the install
// and the sync arm hand openrecord.Install/Sync s.homeDir, and the fan-out
// resolves adapter.SkillsDir(homeDir) itself — a workspace-scoped path table
// would declare files the writer never touches.
func TestComponentPathsOpenRecordIgnoresWorkspaceScope(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	claudeSkills := filepath.Join(home, ".claude", "skills")
	seedOpenRecordFanOut(t, claudeSkills, map[string]string{"openrecord-consult/SKILL.md": "consult"})

	selection, _ := openRecordPlan(model.AgentClaudeCode)
	adapters := resolveAdapters(selection.Agents)

	global := componentPathsWithWorkspaceScoped(home, workspace, ScopeGlobal, selection, adapters, model.ComponentOpenRecord)
	scoped := componentPathsWithWorkspaceScoped(home, workspace, ScopeWorkspace, selection, adapters, model.ComponentOpenRecord)

	if !slices.Equal(global, scoped) {
		t.Fatalf("openrecord paths differ by scope:\nglobal    = %v\nworkspace = %v", global, scoped)
	}
	if !containsPath(scoped, filepath.Join(claudeSkills, ".openrecord-emitted.json")) {
		t.Fatalf("workspace-scoped openrecord paths = %v, want the home-rooted manifest", scoped)
	}
	for _, path := range scoped {
		if strings.HasPrefix(path, workspace) {
			t.Errorf("openrecord declared a workspace-rooted path %q; the fan-out only writes under homeDir", path)
		}
	}
}

// TestSyncComponentPathsOpenRecordMatchesInstall pins that sync does not
// diverge from install the way ComponentPersona does: sync's arm calls
// openrecord.Sync, which runs the same EmitAndFanOut into the same
// directories, so the two contracts must stay identical.
func TestSyncComponentPathsOpenRecordMatchesInstall(t *testing.T) {
	home := t.TempDir()
	seedOpenRecordFanOut(t, filepath.Join(home, ".claude", "skills"),
		map[string]string{"openrecord-consult/SKILL.md": "consult"})

	selection, _ := openRecordPlan(model.AgentClaudeCode)
	adapters := resolveAdapters(selection.Agents)

	install := componentPathsWithWorkspace(home, "", selection, adapters, model.ComponentOpenRecord)
	sync := syncComponentPathsWithWorkspace(home, "", selection, adapters, model.ComponentOpenRecord)

	if !slices.Equal(install, sync) {
		t.Fatalf("sync path table diverged from install:\ninstall = %v\nsync    = %v", install, sync)
	}
	if len(install) == 0 {
		t.Fatal("both tables were empty; the test proves nothing")
	}
}

// TestBackupTargetsOpenRecordIncludeManifestAndEmittedSkills pins that no
// per-component backup extra is needed: the manifest is part of the path table
// itself, so it reaches the snapshot through the ordinary component loop.
// The manifest has to be in there — pruneStaleEntries reads the *previous*
// manifest to decide what to delete, so a rollback that restored the skills
// but not their ownership record would leave the next emit pruning against a
// record that never described what is on disk.
func TestBackupTargetsOpenRecordIncludeManifestAndEmittedSkills(t *testing.T) {
	home := t.TempDir()
	claudeSkills := filepath.Join(home, ".claude", "skills")
	seedOpenRecordFanOut(t, claudeSkills, map[string]string{"openrecord-consult/SKILL.md": "consult"})

	selection, resolved := openRecordPlan(model.AgentClaudeCode)
	targets, err := backupTargets(home, "", ScopeGlobal, selection, resolved)
	if err != nil {
		t.Fatalf("backupTargets() error = %v", err)
	}
	for _, want := range []string{
		filepath.Join(claudeSkills, "openrecord-consult", "SKILL.md"),
		filepath.Join(claudeSkills, ".openrecord-emitted.json"),
	} {
		if !containsPath(targets, want) {
			t.Errorf("backupTargets missing openrecord path %q; targets=%v", want, targets)
		}
	}
}

// TestSyncBackupTargetsOpenRecordIncludeManifestAndEmittedSkills is the same
// contract for sync, which re-emits over the same files.
func TestSyncBackupTargetsOpenRecordIncludeManifestAndEmittedSkills(t *testing.T) {
	home := t.TempDir()
	claudeSkills := filepath.Join(home, ".claude", "skills")
	seedOpenRecordFanOut(t, claudeSkills, map[string]string{"openrecord-consult/SKILL.md": "consult"})

	selection, _ := openRecordPlan(model.AgentClaudeCode)
	targets, err := syncBackupTargets(home, "", selection, resolveAdapters(selection.Agents))
	if err != nil {
		t.Fatalf("syncBackupTargets() error = %v", err)
	}
	for _, want := range []string{
		filepath.Join(claudeSkills, "openrecord-consult", "SKILL.md"),
		filepath.Join(claudeSkills, ".openrecord-emitted.json"),
	} {
		if !containsPath(targets, want) {
			t.Errorf("syncBackupTargets missing openrecord path %q; targets=%v", want, targets)
		}
	}
}

// TestPostApplyVerificationOpenRecordAssertsFannedOutSkills pins the second
// consequence of the path table entry: after a successful fan-out the emitted
// skills and the manifest are required files, and a fan-out that silently
// wrote nothing fails the run instead of reporting success.
func TestPostApplyVerificationOpenRecordAssertsFannedOutSkills(t *testing.T) {
	home := t.TempDir()
	stageOpenRecordEmit(t, true)
	selection, resolved := openRecordPlan(model.AgentClaudeCode)

	step := componentApplyStep{component: model.ComponentOpenRecord, homeDir: home, agents: selection.Agents}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	report := runPostApplyVerification(postApplyVerificationInput{
		HomeDir: home, Scope: ScopeGlobal, Selection: selection, Resolved: resolved,
	})
	if !report.Ready {
		t.Fatalf("verification failed after a successful fan-out: %+v", report.Checks)
	}
	skill := filepath.Join(home, ".claude", "skills", "openrecord-consult", "SKILL.md")
	if !slices.ContainsFunc(report.Checks, func(r verify.CheckResult) bool {
		return string(r.ID) == "verify:file:"+skill
	}) {
		t.Fatalf("verification did not assert the fanned-out skill %q; results=%+v", skill, report.Checks)
	}

	// Remove what the fan-out wrote: verification must now fail rather than
	// report a healthy install over an empty skills directory.
	if err := os.Remove(skill); err != nil {
		t.Fatalf("Remove(%q) error = %v", skill, err)
	}
	if runPostApplyVerification(postApplyVerificationInput{
		HomeDir: home, Scope: ScopeGlobal, Selection: selection, Resolved: resolved,
	}).Ready {
		t.Fatal("verification passed with the fanned-out skill deleted; the path table asserts nothing")
	}
}

// TestOpenRecordInstallRollbackRestoresPreviousFanOut is the point of the path
// table, the backup contract and the manifest decision together: a failed
// install must put back exactly what it replaced.
//
// It drives the real chain — backupTargets → backup.Snapshotter →
// componentApplyStep (the real emit and fan-out) → rollbackRestoreStep — and
// covers the three ways a re-emit can change a target directory: an
// overwritten skill, a skill the new emit no longer ships (pruneStaleEntries
// deletes it), and the manifest itself.
func TestOpenRecordInstallRollbackRestoresPreviousFanOut(t *testing.T) {
	home := t.TempDir()
	claudeSkills := filepath.Join(home, ".claude", "skills")
	const previousConsult = "PREVIOUS consult body"
	const previousRetired = "PREVIOUS retired body"
	previousManifest := seedOpenRecordFanOut(t, claudeSkills, map[string]string{
		"openrecord-consult/SKILL.md": previousConsult,
		"openrecord-retired/SKILL.md": previousRetired,
	})

	selection, resolved := openRecordPlan(model.AgentClaudeCode)

	// 1. Snapshot, exactly as prepareBackupStep does.
	targets, err := backupTargets(home, "", ScopeGlobal, selection, resolved)
	if err != nil {
		t.Fatalf("backupTargets() error = %v", err)
	}
	manifest, err := backup.NewSnapshotter().Create(filepath.Join(t.TempDir(), "snapshot"), targets)
	if err != nil {
		t.Fatalf("Snapshotter.Create() error = %v", err)
	}

	// 2. Install for real. The staged emit ships only openrecord-consult, with
	// new content, so the fan-out overwrites one file, prunes the other and
	// rewrites the manifest.
	stageOpenRecordEmit(t, true)
	step := componentApplyStep{component: model.ComponentOpenRecord, homeDir: home, agents: selection.Agents}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	consult := filepath.Join(claudeSkills, "openrecord-consult", "SKILL.md")
	retired := filepath.Join(claudeSkills, "openrecord-retired", "SKILL.md")
	if body, _ := os.ReadFile(consult); string(body) == previousConsult {
		t.Fatal("the fan-out did not overwrite the previous skill; the rollback proves nothing")
	}
	if _, err := os.Stat(retired); !os.IsNotExist(err) {
		t.Fatalf("the fan-out did not prune the retired skill (err = %v); the rollback proves nothing", err)
	}

	// 3. A later step fails: roll back.
	rollback := rollbackRestoreStep{id: "apply:rollback-restore", state: &runtimeState{manifest: manifest}, homeDir: home}
	if err := rollback.Rollback(); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	for _, want := range []struct {
		path string
		body string
	}{
		{consult, previousConsult},
		{retired, previousRetired},
		{filepath.Join(claudeSkills, ".openrecord-emitted.json"), string(previousManifest)},
	} {
		body, err := os.ReadFile(want.path)
		if err != nil {
			t.Errorf("rollback did not restore %q: %v", want.path, err)
			continue
		}
		if string(body) != want.body {
			t.Errorf("rollback restored %q with the wrong content:\ngot  %q\nwant %q", want.path, body, want.body)
		}
	}
}

// ─── Decisions this batch made by NOT changing something ───────────────────

// TestNeedsCompatibilitySkillsRefreshIgnoresOpenRecord pins that openrecord
// does not gate the ~/.agents/skills refresh. The predicate is not "does this
// component write skills" but "does this component write into the
// compatibility tree": compatibilitySkillsRefreshStep re-injects gentle-ai's
// own embedded Skills and SDD assets there, while openrecord fans out into
// each adapter's SkillsDir — and ~/.agents/skills is no adapter's. Adding it
// would schedule a step whose body does nothing for this component.
func TestNeedsCompatibilitySkillsRefreshIgnoresOpenRecord(t *testing.T) {
	if needsCompatibilitySkillsRefresh([]model.ComponentID{model.ComponentOpenRecord}) {
		t.Fatal("openrecord gated the compatibility-skills refresh; it writes into adapter skills dirs, never ~/.agents/skills")
	}
	// The gate still fires for the components that do write there.
	for _, component := range []model.ComponentID{model.ComponentSkills, model.ComponentSDD} {
		if !needsCompatibilitySkillsRefresh([]model.ComponentID{model.ComponentOpenRecord, component}) {
			t.Errorf("compatibility refresh did not fire for %q alongside openrecord", component)
		}
	}
	paths, err := compatibilitySkillPaths(t.TempDir(), []model.ComponentID{model.ComponentOpenRecord}, model.Selection{})
	if err != nil {
		t.Fatalf("compatibilitySkillPaths() error = %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("compatibilitySkillPaths(openrecord) = %v, want none", paths)
	}
}

// TestPiOnlyComponentsExcludesOpenRecord pins the Pi-only default. Pi exposes
// no skills directory, so openrecord's only delivery channel skips it
// entirely: selecting it by default would install a binary and run qmd setup
// for a Pi-only user and then warn, on every install and every sync, that
// nothing was fanned out. Engram and Persona are in the set because Pi
// actually receives both.
func TestPiOnlyComponentsExcludesOpenRecord(t *testing.T) {
	if slices.Contains(piOnlyComponents(), model.ComponentOpenRecord) {
		t.Fatalf("piOnlyComponents() = %v, want openrecord absent — Pi has no skills directory to fan out to", piOnlyComponents())
	}
}

// TestRequiredDoctorToolsIncludeOpenRecord pins openrecord as an
// ecosystem-level binary: it ships in every non-custom preset, so the doctor
// checks it on PATH unconditionally, like engram and gga.
func TestRequiredDoctorToolsIncludeOpenRecord(t *testing.T) {
	if !slices.Contains(requiredDoctorTools(nil), "openrecord") {
		t.Fatalf("requiredDoctorTools(nil) = %v, want openrecord", requiredDoctorTools(nil))
	}
}

// ─── The install/sync seams ────────────────────────────────────────────────

// TestComponentApplyStepOpenRecordRunsNothingThroughTheInstallSeam pins why
// the seam exists. openrecord ships in every non-custom preset, so any test
// that reaches the install arm without substituting something runs the real
// tool: `openrecord version`, `openrecord qmd install` and an emit into its
// home on a machine that has the binary, and `go install …@latest` over the
// network on one that does not. With installOpenRecordWithHome substituted,
// the arm must execute no command and look nothing up on PATH — the only two
// ways this package reaches the machine — and write nothing under the home.
func TestComponentApplyStepOpenRecordRunsNothingThroughTheInstallSeam(t *testing.T) {
	previousInstall := installOpenRecordWithHome
	previousRun, previousLookPath := runCommand, cmdLookPath
	t.Cleanup(func() {
		installOpenRecordWithHome = previousInstall
		runCommand, cmdLookPath = previousRun, previousLookPath
	})
	runCommand = func(name string, args ...string) error {
		t.Errorf("the install arm executed %q %v; with the seam substituted it must run no command at all", name, args)
		return nil
	}
	cmdLookPath = func(name string) (string, error) {
		t.Errorf("the install arm looked %q up on PATH; with the seam substituted it must touch neither PATH nor the network", name)
		return "", os.ErrNotExist
	}

	var gotHome string
	var gotAgents []model.AgentID
	installOpenRecordWithHome = func(homeDir string, agents []model.AgentID, _ communitytool.Runner, _ communitytool.Detector) (openrecord.InstallResult, error) {
		gotHome, gotAgents = homeDir, agents
		return openrecord.InstallResult{FannedOut: len(agents)}, nil
	}

	home := t.TempDir()
	agents := []model.AgentID{model.AgentClaudeCode, model.AgentGeminiCLI}
	runtime := &runtimeState{}
	step := componentApplyStep{component: model.ComponentOpenRecord, homeDir: home, agents: agents, state: runtime}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if gotHome != home || !slices.Equal(gotAgents, agents) {
		t.Fatalf("the seam received (%q, %v), want (%q, %v)", gotHome, gotAgents, home, agents)
	}
	if !runtime.openRecordResolved || runtime.openRecordFannedOut != len(agents) {
		t.Fatalf("runtimeState openrecord outcome = (%t, %d), want (true, %d)", runtime.openRecordResolved, runtime.openRecordFannedOut, len(agents))
	}
	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("the install arm wrote %v into the home; with the seam substituted nothing may reach the filesystem", entries)
	}
}

// TestComponentSyncStepOpenRecordGoesThroughTheSyncSeam is the sync half of
// the same contract. The arm's own availability probe still runs through
// runCommand/cmdLookPath — that pair is the seam for it — but the emit and
// fan-out must reach the machine only through syncOpenRecordWithHome.
func TestComponentSyncStepOpenRecordGoesThroughTheSyncSeam(t *testing.T) {
	previousSync := syncOpenRecordWithHome
	t.Cleanup(func() { syncOpenRecordWithHome = previousSync })
	calls := stageOpenRecordEmit(t, true)

	var called bool
	syncOpenRecordWithHome = func(string, []model.AgentID, communitytool.Runner, communitytool.Detector) (openrecord.InstallResult, error) {
		called = true
		return openrecord.InstallResult{FannedOut: 1}, nil
	}

	home := t.TempDir()
	step := componentSyncStep{component: model.ComponentOpenRecord, homeDir: home, agents: []model.AgentID{model.AgentClaudeCode}}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !called {
		t.Fatal("the sync arm did not go through syncOpenRecordWithHome")
	}
	if ranCommand(*calls, "openrecord", "skills", "--emit") {
		t.Fatalf("the sync arm emitted skills past the seam: %v", *calls)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills")); !os.IsNotExist(err) {
		t.Errorf("the sync arm wrote into the skills dir past the seam: err = %v", err)
	}
}
