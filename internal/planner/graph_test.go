package planner

import (
	"testing"

	"github.com/franwerner/gentle-ai/v3/internal/model"
)

// TestMVPGraphCarriesOpenRecordWithNoDependencies pins openrecord's edge: it
// is a node of the graph, and it depends on nothing — its skills come from the
// openrecord binary itself and land in a directory its own fan-out creates, so
// no other component has to run first.
func TestMVPGraphCarriesOpenRecordWithNoDependencies(t *testing.T) {
	graph := MVPGraph()
	if !graph.Has(model.ComponentOpenRecord) {
		t.Fatal("MVPGraph() has no openrecord node")
	}
	if deps := graph.DependenciesOf(model.ComponentOpenRecord); len(deps) != 0 {
		t.Fatalf("DependenciesOf(openrecord) = %v, want none", deps)
	}
	for _, pair := range SoftOrderingConstraints() {
		if pair[0] == model.ComponentOpenRecord || pair[1] == model.ComponentOpenRecord {
			t.Errorf("openrecord appears in a soft-ordering pair %v, but shares no output file with another component", pair)
		}
	}
}
