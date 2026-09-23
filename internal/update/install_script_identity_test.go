package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// installScriptContent reads the real install script off disk, exactly as its
// sibling install_script_test.go does — never a fixture copy, so a repoint
// applied to one and not the other is caught.
func installScriptContent(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "scripts", name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(content)
}

func goPackageAssignmentLine(t *testing.T, script, marker string) string {
	t.Helper()
	for _, line := range strings.Split(script, "\n") {
		if strings.Contains(line, marker) {
			return line
		}
	}
	t.Fatalf("no line contains %q", marker)
	return ""
}

// TestInstallScriptsSeparateReleaseAndModuleIdentity guards the split itself:
// the go install target must derive only from GO_MODULE_PATH, and the retired
// GITHUB_OWNER/GITHUB_REPO names must not exist to be re-conflated into.
func TestInstallScriptsSeparateReleaseAndModuleIdentity(t *testing.T) {
	bash := installScriptContent(t, "install.sh")
	powershell := installScriptContent(t, "install.ps1")

	bashLine := goPackageAssignmentLine(t, bash, "go_package=")
	if !strings.Contains(bashLine, "GO_MODULE_PATH") {
		t.Fatalf("install.sh go_package assignment does not reference GO_MODULE_PATH: %s", bashLine)
	}
	for _, forbidden := range []string{"RELEASE_OWNER", "RELEASE_REPO", "GITHUB_OWNER", "GITHUB_REPO"} {
		if strings.Contains(bashLine, forbidden) {
			t.Fatalf("install.sh go_package assignment references %s, re-coupling the release and module axes: %s", forbidden, bashLine)
		}
	}

	psLine := goPackageAssignmentLine(t, powershell, "$goPackage =")
	if !strings.Contains(psLine, "GO_MODULE_PATH") {
		t.Fatalf("install.ps1 goPackage assignment does not reference GO_MODULE_PATH: %s", psLine)
	}
	for _, forbidden := range []string{"RELEASE_OWNER", "RELEASE_REPO", "GITHUB_OWNER", "GITHUB_REPO"} {
		if strings.Contains(psLine, forbidden) {
			t.Fatalf("install.ps1 goPackage assignment references %s, re-coupling the release and module axes: %s", forbidden, psLine)
		}
	}

	for _, retired := range []string{"GITHUB_OWNER", "GITHUB_REPO"} {
		if strings.Contains(bash, retired) {
			t.Fatalf("install.sh still contains the retired name %s; the release/module conflation can be re-expressed under it", retired)
		}
		if strings.Contains(powershell, retired) {
			t.Fatalf("install.ps1 still contains the retired name %s; the release/module conflation can be re-expressed under it", retired)
		}
	}
}

// TestInstallScriptsResolveReleasesFromTheFork guards against a stale release
// literal: every remaining upstream-org occurrence must be on the explicit
// allow-list of module-axis literals this change deliberately leaves alone.
func TestInstallScriptsResolveReleasesFromTheFork(t *testing.T) {
	bash := installScriptContent(t, "install.sh")
	powershell := installScriptContent(t, "install.ps1")

	for _, want := range []string{
		`RELEASE_OWNER="franwerner"`,
		`RELEASE_REPO="gentle-ai"`,
		`BREW_TAP="franwerner/homebrew-tap"`,
		`BREW_FORMULA_REF="franwerner/tap/${BINARY_NAME}"`,
	} {
		if !strings.Contains(bash, want) {
			t.Fatalf("install.sh is missing %q", want)
		}
	}
	for _, want := range []string{
		`$RELEASE_OWNER = "franwerner"`,
		`$RELEASE_REPO = "gentle-ai"`,
	} {
		if !strings.Contains(powershell, want) {
			t.Fatalf("install.ps1 is missing %q", want)
		}
	}

	bashAllowed := []string{
		`GO_MODULE_PATH="github.com/franwerner/gentle-ai/v3"`,
		"prepend_go_env_pattern GONOSUMDB github.com/franwerner/gentle-ai/v3",
		"prepend_go_env_pattern GOPRIVATE github.com/franwerner/gentle-ai/v3",
		"prepend_go_env_pattern GONOPROXY github.com/franwerner/gentle-ai/v3",
	}
	assertOnlyAllowedOccurrences(t, "install.sh", bash, bashAllowed)

	psAllowed := []string{
		`$GO_MODULE_PATH = "github.com/franwerner/gentle-ai/v3"`,
		`$STABLE_SOURCE_COMMAND = "go install github.com/franwerner/gentle-ai/v3/cmd/gentle-ai@latest"`,
		`Add-GoEnvPattern -Name "GONOSUMDB" -Pattern "github.com/franwerner/gentle-ai/v3"`,
		`Add-GoEnvPattern -Name "GOPRIVATE" -Pattern "github.com/franwerner/gentle-ai/v3"`,
		`Add-GoEnvPattern -Name "GONOPROXY" -Pattern "github.com/franwerner/gentle-ai/v3"`,
	}
	assertOnlyAllowedOccurrences(t, "install.ps1", powershell, psAllowed)
}

// assertOnlyAllowedOccurrences fails, naming the line, on any mention of the
// upstream org outside the given allow-list of module-axis literals.
func assertOnlyAllowedOccurrences(t *testing.T, scriptName, script string, allowed []string) {
	t.Helper()
	for lineNumber, line := range strings.Split(script, "\n") {
		if !strings.Contains(strings.ToLower(line), "gentleman-programming") {
			continue
		}
		isAllowed := false
		for _, allow := range allowed {
			if strings.Contains(line, allow) {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			t.Fatalf("%s:%d references the upstream org outside the module-axis allow-list: %s", scriptName, lineNumber+1, line)
		}
	}
}

// TestInstallScriptsGoInstallTargetIsTheUnchangedModulePath guards the module
// axis itself: it must stay the fixed module path on both channels, whatever
// the release axis is set to.
func TestInstallScriptsGoInstallTargetIsTheUnchangedModulePath(t *testing.T) {
	const wantModulePath = "github.com/franwerner/gentle-ai/v3"

	bash := installScriptContent(t, "install.sh")
	if !strings.Contains(bash, `GO_MODULE_PATH="`+wantModulePath+`"`) {
		t.Fatalf("install.sh GO_MODULE_PATH is not %q", wantModulePath)
	}
	if !strings.Contains(bash, `go_package="${GO_MODULE_PATH}/cmd/${BINARY_NAME}@${version}"`) {
		t.Fatal("install.sh go_package is not built from GO_MODULE_PATH for both channels")
	}

	powershell := installScriptContent(t, "install.ps1")
	if !strings.Contains(powershell, `$GO_MODULE_PATH = "`+wantModulePath+`"`) {
		t.Fatalf("install.ps1 GO_MODULE_PATH is not %q", wantModulePath)
	}
	if !strings.Contains(powershell, `$goPackage = "$GO_MODULE_PATH/cmd/$BINARY_NAME@$version"`) {
		t.Fatal("install.ps1 goPackage is not built from GO_MODULE_PATH for both channels")
	}
}
