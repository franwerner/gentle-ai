package model

import (
	"slices"
	"testing"
)

func TestComponentsForPresetFullGentlemanUsesInstallSafeVisualInventory(t *testing.T) {
	tests := []struct {
		name    string
		persona PersonaID
	}{
		{name: "gentleman persona", persona: PersonaGentleman},
		{name: "custom persona", persona: PersonaCustom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComponentsForPreset(PresetFullGentleman, tt.persona)

			if slices.Contains(got, ComponentTheme) {
				t.Fatalf("ComponentsForPreset() includes generic ComponentTheme: %v", got)
			}
			for _, want := range []ComponentID{ComponentClaudeTheme, ComponentOpenCodeGentleLogo} {
				if !slices.Contains(got, want) {
					t.Errorf("ComponentsForPreset() missing safe visual component %q: %v", want, got)
				}
			}
		})
	}
}

func TestVisualPolishComponentsReturnsCompleteManagedCleanupInventory(t *testing.T) {
	want := []ComponentID{ComponentTheme, ComponentClaudeTheme, ComponentOpenCodeGentleLogo}
	if got := VisualPolishComponents(); !slices.Equal(got, want) {
		t.Fatalf("VisualPolishComponents() = %v, want complete cleanup inventory %v", got, want)
	}
}

// TestComponentsForPresetAlwaysCarriesOpenRecord pins the "installed always,
// never chosen" contract: openrecord ships in every non-custom preset, exactly
// like engram, and PresetCustom stays the only way to opt out.
func TestComponentsForPresetAlwaysCarriesOpenRecord(t *testing.T) {
	for _, preset := range []PresetID{PresetFullGentleman, PresetEcosystemOnly, PresetMinimal} {
		for _, persona := range []PersonaID{PersonaGentleman, PersonaCustom} {
			got := ComponentsForPreset(preset, persona)
			if !slices.Contains(got, ComponentOpenRecord) {
				t.Errorf("ComponentsForPreset(%q, %q) is missing %q: %v", preset, persona, ComponentOpenRecord, got)
			}
			if slices.Contains(got, ComponentEngram) != slices.Contains(got, ComponentOpenRecord) {
				t.Errorf("ComponentsForPreset(%q, %q) does not carry openrecord wherever it carries engram: %v", preset, persona, got)
			}
		}
	}
	if got := ComponentsForPreset(PresetCustom, PersonaGentleman); got != nil {
		t.Errorf("ComponentsForPreset(custom) = %v, want nil", got)
	}
}
