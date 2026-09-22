package sddstatus

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestProjectStatusV2ReportsOpenRecordUnconditionally pins the public v2
// document's record store as a constant: openrecord installs unconditionally,
// so no workspace state can move it.
func TestProjectStatusV2ReportsOpenRecordUnconditionally(t *testing.T) {
	status := baseStatus(ArtifactStoreOpenSpec, t.TempDir(), nil, nil, nil, "apply", nil)

	projected, err := ProjectStatusV2(status)
	if err != nil {
		t.Fatalf("ProjectStatusV2() error = %v", err)
	}
	if projected.RecordStore.Declared != "openrecord" || projected.RecordStore.Resolved != "openrecord" {
		t.Fatalf("projected RecordStore = %+v, want {Declared: openrecord, Resolved: openrecord}", projected.RecordStore)
	}

	raw, err := json.Marshal(projected)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	recordStore, ok := decoded["recordStore"].(map[string]any)
	if !ok {
		t.Fatalf("recordStore field missing or wrong shape in projected JSON: %s", raw)
	}
	if recordStore["declared"] != "openrecord" || recordStore["resolved"] != "openrecord" {
		t.Errorf("recordStore JSON = %+v, want declared/resolved = openrecord", recordStore)
	}
}

// TestProjectStatusV2IgnoresConfigRecordStoreKey proves the declaration axis
// is gone rather than merely defaulted: a workspace still carrying the retired
// key — with any value — projects the same constant.
func TestProjectStatusV2IgnoresConfigRecordStoreKey(t *testing.T) {
	for _, declared := range []string{"openrecord", "not-a-real-store", ""} {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "openspec"), 0o755); err != nil {
			t.Fatal(err)
		}
		content := "strict_tdd: true\n"
		if declared != "" {
			content = "record_store: " + declared + "\n"
		}
		if err := os.WriteFile(filepath.Join(root, "openspec", "config.yaml"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		projected, err := ProjectStatusV2(baseStatus(ArtifactStoreOpenSpec, root, nil, nil, nil, "apply", nil))
		if err != nil {
			t.Fatalf("ProjectStatusV2() with record_store %q error = %v", declared, err)
		}
		if projected.RecordStore.Declared != "openrecord" || projected.RecordStore.Resolved != "openrecord" {
			t.Errorf("record_store %q projected RecordStore = %+v, want the openrecord constant", declared, projected.RecordStore)
		}
	}
}
