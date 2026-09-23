// Package openrecord installs and detects openrecord, gentle-ai's durable
// record store component: the binary, its qmd prerequisite, and the skill
// fan-out to every selected agent that exposes a skills directory.
package openrecord

import (
	"errors"
	"fmt"
	"os"

	"github.com/franwerner/gentle-ai/v3/internal/components/communitytool"
	"github.com/franwerner/gentle-ai/v3/internal/model"
)

const goImportPath = "github.com/franwerner/open-record/cmd/openrecord"

// installScriptURL points at the script's documented curl|bash invocation
// (copia/open-record/scripts/install.sh header comment), always the current
// default branch: the script itself resolves the release to install.
const installScriptURL = "https://raw.githubusercontent.com/franwerner/open-record/master/scripts/install.sh"

// InstallResult reports the one non-fatal outcome an openrecord install or
// sync can have: how many selected agents actually received the fanned-out
// skills. Zero is not an error — it means no selected agent exposes a skills
// directory — so the caller warns instead of failing the run.
type InstallResult struct {
	FannedOut int
}

// Install runs the ordered component sequence: install the binary only when
// missing, install qmd unconditionally, emit skills with qmd into a staging
// directory, then fan them out to every selected agent that exposes a skills
// directory. A step does not run when the step before it failed, so skills are
// never emitted describing a tool whose qmd half is missing.
func Install(homeDir string, selectedAgents []model.AgentID, runner communitytool.Runner, detector communitytool.Detector) (InstallResult, error) {
	if runner == nil {
		return InstallResult{}, fmt.Errorf("openrecord runner is not configured")
	}
	if detector == nil {
		detector = defaultDetector()
	}

	if DetectStatus(homeDir, detector, runner).CLI != communitytool.AvailabilityAvailable {
		command, err := binaryInstallCommand(detector)
		if err != nil {
			return InstallResult{}, err
		}
		if err := runner.Run(command[0], command[1:]...); err != nil {
			return InstallResult{}, fmt.Errorf("install openrecord binary: %w", err)
		}
	}

	if err := runner.Run("openrecord", "qmd", "install"); err != nil {
		return InstallResult{}, fmt.Errorf("openrecord qmd install: %w", err)
	}

	return EmitAndFanOut(homeDir, selectedAgents, runner)
}

// ErrBinaryUnavailable reports that the openrecord binary is not on PATH.
//
// Sync degrades on it rather than failing, the way the engram arm does: engram
// never shells out at all, because gentle-ai embeds its protocol asset, and it
// ignores a version lookup that fails. openrecord's skills are produced by the
// openrecord binary and deliberately not embedded — a private copy would drift
// from the tool in silence — so a sync without the binary has nothing to copy.
// Refreshing what it can is sync's job; reporting a broken install is the
// doctor's, and it already lists openrecord in coreTools.
var ErrBinaryUnavailable = errors.New("the openrecord binary is not installed")

// SyncWithDetector is Sync with the binary lookup injected. The availability
// probe lives here rather than in the caller so that substituting this arm
// substitutes the whole of it: a probe sitting in front of the seam leaves
// every caller reaching for the machine even when the arm itself is stubbed.
//
// An absent binary is a hard failure, matching the doctor, which lists
// openrecord in coreTools and FAILs without it. Sync provisions nothing, so
// finding no binary here means `gentle-ai install` never ran or did not
// finish; the error names the command that fixes it.
func SyncWithDetector(homeDir string, selectedAgents []model.AgentID, runner communitytool.Runner, detector communitytool.Detector) (InstallResult, error) {
	if runner == nil {
		return InstallResult{}, fmt.Errorf("openrecord runner is not configured")
	}
	if DetectStatus(homeDir, detector, runner).CLI != communitytool.AvailabilityAvailable {
		return InstallResult{}, ErrBinaryUnavailable
	}
	return EmitAndFanOut(homeDir, selectedAgents, runner)
}

// EmitAndFanOut is the step pair install, sync and post-upgrade all share:
// `openrecord skills --emit <staging> --with-qmd` followed by the fan-out into
// every selected agent's skills directory.
func EmitAndFanOut(homeDir string, selectedAgents []model.AgentID, runner communitytool.Runner) (InstallResult, error) {
	if runner == nil {
		return InstallResult{}, fmt.Errorf("openrecord runner is not configured")
	}

	staging, err := os.MkdirTemp("", "gentle-ai-openrecord-emit-*")
	if err != nil {
		return InstallResult{}, fmt.Errorf("stage openrecord skills: %w", err)
	}
	defer os.RemoveAll(staging)

	if err := runner.Run("openrecord", "skills", "--emit", staging, "--with-qmd"); err != nil {
		return InstallResult{}, fmt.Errorf("openrecord skills --emit: %w", err)
	}

	fanned, err := fanOut(staging, homeDir, selectedAgents)
	return InstallResult{FannedOut: fanned}, err
}

// binaryInstallCommand mirrors effectiveMethod's own predicate (Go on PATH
// preferred over the install script) so the first-time install and the
// upgrade path never disagree about which copy is authoritative.
func binaryInstallCommand(detector communitytool.Detector) ([]string, error) {
	if detector == nil {
		return nil, fmt.Errorf("openrecord install: no detector configured")
	}
	if _, err := detector.LookPath("go"); err == nil {
		return []string{"go", "install", goImportPath + "@latest"}, nil
	}
	return []string{"sh", "-c", fmt.Sprintf("curl -fsSL %s | env WITH_QMD=no bash", installScriptURL)}, nil
}
