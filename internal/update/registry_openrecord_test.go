package update

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/components/openrecord"
)

// TestOpenRecordRegistryEntryIsComponentCoherent pins the upgrade half of
// openrecord being a component rather than a community tool: `gentle-ai
// upgrade` must know the binary, find it wherever either install route put it,
// and re-emit the skills afterwards.
//
// Detection is by `openrecord version`, never by selection, which is exactly
// why PostUpgrade carries its own component gate (see
// openrecord.TestPostUpgradeSelectionGate): a user with the binary on PATH for
// their own reasons must not have skills written into their agents.
func TestOpenRecordRegistryEntryIsComponentCoherent(t *testing.T) {
	index := slices.IndexFunc(Tools, func(tool ToolInfo) bool { return tool.Name == "openrecord" })
	if index < 0 {
		t.Fatal("update.Tools has no openrecord entry; `gentle-ai upgrade` would not know the binary")
	}
	entry := Tools[index]

	if !slices.Equal(entry.DetectCmd, []string{"openrecord", "version"}) {
		t.Errorf("DetectCmd = %v, want [openrecord version]", entry.DetectCmd)
	}
	if entry.InstallMethod != InstallScript {
		t.Errorf("InstallMethod = %v, want InstallScript — effectiveMethod routes a machine with Go on PATH to go-install via GoImportPath", entry.InstallMethod)
	}
	if entry.GoImportPath == "" {
		t.Error("GoImportPath is empty; the go-install route cannot be selected")
	}
	if entry.PostUpgrade == nil {
		t.Error("PostUpgrade is nil; an upgraded binary would ship stale skills until the next install or sync")
	}

	// FallbackPaths must be the component's own, so upgrade detection and the
	// component's resolveBinaryPath never disagree about where to look.
	if entry.FallbackPaths == nil {
		t.Fatal("FallbackPaths is nil")
	}
	home := filepath.FromSlash("/home/tester")
	if !slices.Equal(entry.FallbackPaths(home, ""), openrecord.FallbackPaths(home, "")) {
		t.Errorf("FallbackPaths = %v, want openrecord.FallbackPaths = %v",
			entry.FallbackPaths(home, ""), openrecord.FallbackPaths(home, ""))
	}
}
