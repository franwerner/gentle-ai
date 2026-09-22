package communitytool

import (
	"os"
	"path/filepath"
	"testing"
)

// A leftover ~/.pi directory is not an installed Pi. Reconciling CodeGraph for
// one cannot succeed — the MCP capability probe needs Pi's own adapter
// extension, which only a real install provides — and its failure used to take
// the entire CodeGraph install step down with it.
func TestReconcileDetectedPiCodeGraphSkipsAnOrphanAgentDir(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".pi", "agent", "extensions"), 0o755); err != nil {
		t.Fatal(err)
	}

	restore := piBinaryLookPath
	t.Cleanup(func() { piBinaryLookPath = restore })
	piBinaryLookPath = func(string) (string, error) { return "", os.ErrNotExist }

	result, err := reconcileDetectedPiCodeGraph(home, t.TempDir())
	if err != nil {
		t.Fatalf("reconcileDetectedPiCodeGraph() error = %v, want the orphan directory skipped", err)
	}
	if result != nil {
		t.Fatalf("reconcileDetectedPiCodeGraph() = %+v, want nil for an agent that is not installed", result)
	}
}
