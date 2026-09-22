package filemerge

import "testing"

// A hermes config is YAML the user also owns. Install adds mcp_servers blocks to
// it; before RemoveYAMLMCPServerBlock existed the uninstaller had no way to take
// them back out, so gentle-ai's entries outlived the uninstall in that file.
func TestRemoveYAMLMCPServerBlockLeavesForeignKeysIntact(t *testing.T) {
	content := "plugins:\n  enabled:\n    - orca-status\n\nmcp_servers:\n  context7:\n    command: npx\n    args:\n      - -y\n  engram:\n    command: engram\n    args:\n      - mcp\n"

	updated, changed := RemoveYAMLMCPServerBlock(content, "context7")
	if !changed {
		t.Fatal("RemoveYAMLMCPServerBlock() reported no change for a server that is present")
	}
	if got := ReadYAMLMCPServerCommandOrEmpty(updated, "context7"); got != "" {
		t.Errorf("context7 survived removal with command %q", got)
	}
	if got := ReadYAMLMCPServerCommandOrEmpty(updated, "engram"); got != "engram" {
		t.Errorf("engram command = %q, want it untouched by context7's removal", got)
	}

	updated, changed = RemoveYAMLMCPServerBlock(updated, "engram")
	if !changed {
		t.Fatal("removing the last server reported no change")
	}
	want := "plugins:\n  enabled:\n    - orca-status\n"
	if updated != want {
		t.Errorf("after removing every server:\ngot  %q\nwant %q", updated, want)
	}
}

func TestRemoveYAMLMCPServerBlockIsANoOpWhenAbsent(t *testing.T) {
	content := "plugins:\n  enabled:\n    - orca-status\n"
	updated, changed := RemoveYAMLMCPServerBlock(content, "engram")
	if changed {
		t.Error("RemoveYAMLMCPServerBlock() reported a change for a file with no mcp_servers")
	}
	if updated != content {
		t.Errorf("content mutated: got %q, want %q", updated, content)
	}
}

// ReadYAMLMCPServerCommandOrEmpty is a test convenience over the read helper.
func ReadYAMLMCPServerCommandOrEmpty(content, serverID string) string {
	cmd, ok := ReadYAMLMCPServerCommand(content, serverID)
	if !ok {
		return ""
	}
	return cmd
}
