package engram

import (
	"github.com/franwerner/gentle-ai/v3/internal/installcmd"
	"github.com/franwerner/gentle-ai/v3/internal/model"
	"github.com/franwerner/gentle-ai/v3/internal/system"
)

func InstallCommand(profile system.PlatformProfile) ([][]string, error) {
	return installcmd.NewResolver().ResolveComponentInstall(profile, model.ComponentEngram)
}
