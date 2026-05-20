package project

import (
	"reflect"
	"testing"
	"time"
)

func TestGraphLayoutSavePersistsViewportThemeGridAndNodePosition(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 7, 30, 0, 0, time.UTC), "graph-layout-save")

	before, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest before save: %v", err)
	}
	beforeEdges := append([]Edge(nil), before.Graph.Edges...)

	result := store.SaveGraphLayout(SaveGraphLayoutCommand{
		Root:                 root,
		ExpectedGraphVersion: before.Graph.Version,
		Viewport:             CanvasViewportDTO{X: -88, Y: 144, Zoom: 1.25},
		Theme:                "warm_light",
		Grid:                 CanvasGridDTO{Visible: false, Size: 32, Opacity: 0.12},
		Nodes: []GraphNodeLayoutCommand{
			{
				ID:        "node_shot_001",
				Position:  CanvasPositionDTO{X: 777, Y: 333},
				Size:      CanvasSizeDTO{Width: 268, Height: 132},
				Collapsed: true,
			},
		},
		CorrelationID: "graph-layout-save",
	})
	if !result.OK || result.Canvas == nil {
		t.Fatalf("SaveGraphLayout() = %#v, want canvas", result)
	}

	canvas := result.Canvas
	if canvas.Version != before.Graph.Version+1 {
		t.Fatalf("canvas version = %d, want %d", canvas.Version, before.Graph.Version+1)
	}
	if canvas.Viewport.X != -88 || canvas.Viewport.Y != 144 || canvas.Viewport.Zoom != 1.25 {
		t.Fatalf("viewport = %#v, want persisted viewport", canvas.Viewport)
	}
	if canvas.Theme != "warm_light" || canvas.Grid.Visible || canvas.Grid.Size != 32 || canvas.Grid.Opacity != 0.12 {
		t.Fatalf("theme/grid = %#v/%#v, want persisted Canvas settings", canvas.Theme, canvas.Grid)
	}

	shot := requireNode(t, canvas.Nodes, "node_shot_001", "shot", "production", "Mina reaches the stall shelter")
	if shot.RefID != "shots/shot-001.json" {
		t.Fatalf("shot RefID = %q, want domain ref unchanged", shot.RefID)
	}
	if shot.Position.X != 777 || shot.Position.Y != 333 {
		t.Fatalf("shot position = %#v, want persisted position", shot.Position)
	}
	if shot.Size.Width != 268 || shot.Size.Height != 132 || !shot.Collapsed {
		t.Fatalf("shot size/collapsed = %#v/%v, want persisted layout", shot.Size, shot.Collapsed)
	}

	after, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest after save: %v", err)
	}
	if after.Graph.Version != before.Graph.Version+1 {
		t.Fatalf("manifest graph version = %d, want %d", after.Graph.Version, before.Graph.Version+1)
	}
	if after.Integrity.LastGraphVersion != after.Graph.Version {
		t.Fatalf("LastGraphVersion = %d, want graph version %d", after.Integrity.LastGraphVersion, after.Graph.Version)
	}
	if !reflect.DeepEqual(after.Graph.Edges, beforeEdges) {
		t.Fatalf("layout save changed graph edges\nbefore=%#v\nafter=%#v", beforeEdges, after.Graph.Edges)
	}
	beforeShot := requireManifestNode(t, before.Graph.Nodes, "node_shot_001")
	afterShot := requireManifestNode(t, after.Graph.Nodes, "node_shot_001")
	if afterShot.RefID != beforeShot.RefID || afterShot.Kind != beforeShot.Kind {
		t.Fatalf("layout save changed domain node truth before=%#v after=%#v", beforeShot, afterShot)
	}
}

func TestGraphLayoutSaveRejectsStaleVersionWithoutMutation(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 7, 31, 0, 0, time.UTC), "graph-layout-stale")

	before, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest before stale save: %v", err)
	}

	result := store.SaveGraphLayout(SaveGraphLayoutCommand{
		Root:                 root,
		ExpectedGraphVersion: before.Graph.Version - 1,
		Viewport:             CanvasViewportDTO{X: 10, Y: 20, Zoom: 1},
		Theme:                "warm_light",
		Grid:                 CanvasGridDTO{Visible: true, Size: 24, Opacity: 0.24},
		Nodes: []GraphNodeLayoutCommand{
			{ID: "node_shot_001", Position: CanvasPositionDTO{X: 900, Y: 900}},
		},
		CorrelationID: "graph-layout-stale",
	})
	if result.OK || result.Error == nil || result.Error.Code != CodeGraphViewStaleVersion {
		t.Fatalf("stale SaveGraphLayout() = %#v, want %s", result, CodeGraphViewStaleVersion)
	}

	after, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest after stale save: %v", err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("stale layout save mutated manifest\nbefore=%#v\nafter=%#v", before, after)
	}
}

func TestGraphLayoutSaveRequiresExpectedGraphVersion(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 7, 32, 0, 0, time.UTC), "graph-layout-version-required")

	before, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest before missing-version save: %v", err)
	}

	result := store.SaveGraphLayout(SaveGraphLayoutCommand{
		Root:          root,
		Viewport:      CanvasViewportDTO{X: 10, Y: 20, Zoom: 1},
		Theme:         "warm_light",
		Grid:          CanvasGridDTO{Visible: true, Size: 24, Opacity: 0.24},
		Nodes:         []GraphNodeLayoutCommand{{ID: "node_shot_001", Position: CanvasPositionDTO{X: 900, Y: 900}}},
		CorrelationID: "graph-layout-version-required",
	})
	if result.OK || result.Error == nil || result.Error.Code != CodeGraphLayoutInvalidCommand {
		t.Fatalf("missing-version SaveGraphLayout() = %#v, want %s", result, CodeGraphLayoutInvalidCommand)
	}

	after, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest after missing-version save: %v", err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("missing-version layout save mutated manifest\nbefore=%#v\nafter=%#v", before, after)
	}
}

func requireManifestNode(t *testing.T, nodes []Node, id string) Node {
	t.Helper()
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	t.Fatalf("manifest node %q missing", id)
	return Node{}
}
