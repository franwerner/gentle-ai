package cli

import (
	"github.com/gentleman-programming/gentle-ai/v2/internal/components/communitytool"
	"github.com/gentleman-programming/gentle-ai/v2/internal/components/openrecord"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// NeutralizeOpenRecordForTest substitutes the openrecord install and sync arms
// with no-ops and returns a restore function.
//
// It is exported because internal/app drives the real install and sync
// pipelines in its TUI tests. Those tests are about the TUI, not about
// openrecord, and openrecord is a hard dependency now: without this they reach
// for a binary the test environment does not have, and CI installs no
// ecosystem binary at all.
//
// The package-private seam vars cover internal/cli's own tests; this is the
// same substitution reachable from a sibling package.
func NeutralizeOpenRecordForTest() func() {
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
