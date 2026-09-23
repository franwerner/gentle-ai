package openrecord

import (
	"context"
	"os"
	"os/exec"
	"slices"

	"github.com/franwerner/gentle-ai/v3/internal/components/communitytool"
	"github.com/franwerner/gentle-ai/v3/internal/model"
	"github.com/franwerner/gentle-ai/v3/internal/state"
)

// PostUpgrade is openrecord's update.ToolInfo.PostUpgrade step: re-emit and
// re-fan-out the skills, exactly like install does, after the binary itself
// upgraded successfully.
//
// It gates on the persisted component selection rather than running whenever
// the binary upgraded: update.Tools is checked by detection (`openrecord
// version`), not by selection, so an ungated re-emit would write skills for a
// user who has the binary on PATH for their own reasons and whose install
// never selected the component. The gate is not vacuous now that openrecord
// ships in every non-custom preset — PresetCustom and an explicit
// `--component` list can both leave it out.
func PostUpgrade(ctx context.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	installed, err := state.Read(homeDir)
	if err != nil {
		// No persisted selection to act on — nothing was ever configured.
		return nil
	}
	if !selectedFor(installed.Components) {
		return nil
	}

	selectedAgents := make([]model.AgentID, 0, len(installed.InstalledAgents))
	for _, agentID := range installed.InstalledAgents {
		selectedAgents = append(selectedAgents, model.AgentID(agentID))
	}

	runner := communitytool.RunnerFunc(func(name string, args ...string) error {
		return exec.CommandContext(ctx, name, args...).Run()
	})

	_, err = EmitAndFanOut(homeDir, selectedAgents, runner)
	return err
}

func selectedFor(components []model.ComponentID) bool {
	return slices.Contains(components, model.ComponentOpenRecord)
}
