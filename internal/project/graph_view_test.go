package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestGraphViewDTOGoldenSchema(t *testing.T) {
	snapshot := map[string][]string{
		"canvas":         jsonTagFields(reflect.TypeOf(ProjectCanvasDTO{})),
		"edge":           jsonTagFields(reflect.TypeOf(GraphEdgeDTO{})),
		"frame":          jsonTagFields(reflect.TypeOf(ProductionFrameDTO{})),
		"grid":           jsonTagFields(reflect.TypeOf(CanvasGridDTO{})),
		"history":        jsonTagFields(reflect.TypeOf(FrameHistorySummaryDTO{})),
		"layout":         jsonTagFields(reflect.TypeOf(CanvasLayoutDTO{})),
		"node":           jsonTagFields(reflect.TypeOf(GraphNodeDTO{})),
		"position":       jsonTagFields(reflect.TypeOf(CanvasPositionDTO{})),
		"referenceGroup": jsonTagFields(reflect.TypeOf(ReferenceGroupDTO{})),
		"size":           jsonTagFields(reflect.TypeOf(CanvasSizeDTO{})),
		"viewport":       jsonTagFields(reflect.TypeOf(CanvasViewportDTO{})),
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("marshal schema snapshot: %v", err)
	}
	data = append(data, '\n')

	goldenPath := filepath.Join("testdata", "graph-view-schema.golden.json")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read graph view schema golden: %v", err)
	}
	if string(data) != string(golden) {
		t.Fatalf("Graph View DTO schema snapshot changed.\nGot:\n%s\nWant:\n%s", data, golden)
	}
}

func TestGraphViewProjectionExampleMapping(t *testing.T) {
	root := exampleFixtureRoot(t)
	store := testStore(time.Date(2026, 5, 20, 6, 30, 0, 0, time.UTC), "graph-view")

	result := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "graph-view"})
	if !result.OK {
		t.Fatalf("GraphView() failed: %#v", result.Error)
	}
	if result.Canvas == nil {
		t.Fatal("GraphView() canvas = nil")
	}

	canvas := result.Canvas
	if canvas.ProjectID != "proj_alpha_fixture" {
		t.Fatalf("ProjectID = %q, want proj_alpha_fixture", canvas.ProjectID)
	}
	if canvas.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("SchemaVersion = %q, want %s", canvas.SchemaVersion, CurrentSchemaVersion)
	}
	if canvas.Version != 3 {
		t.Fatalf("Version = %d, want 3", canvas.Version)
	}
	if canvas.Theme != "dark" || canvas.Grid.Size != 24 || !canvas.Grid.Visible {
		t.Fatalf("canvas theme/grid = %#v/%#v, want stable alpha defaults", canvas.Theme, canvas.Grid)
	}

	requireNode(t, canvas.Nodes, "node_script_source", "script", "source", "Script Source")
	requireNode(t, canvas.Nodes, "node_character_mina", "character", "continuity", "Mina")
	requireNode(t, canvas.Nodes, "node_scene_laneway", "scene", "concept", "Laneway Market After Rain")
	requireNode(t, canvas.Nodes, "node_prop_lantern", "prop", "continuity", "Red Lantern")
	requireNode(t, canvas.Nodes, "node_shot_001", "shot", "production", "Mina reaches the stall shelter")
	requireNode(t, canvas.Nodes, "node_package_001", "package", "handoff", "Package pkg_scene001_shot001")
	requireNode(t, canvas.Nodes, "node_result_mock-result-summary", "video_result", "output", "Mock Result Summary")
	requireNode(t, canvas.Nodes, "node_review_001", "note", "review", "Review note review_stub_001")
	requireNoNode(t, canvas.Nodes, "node_run_001")

	if len(canvas.ReferenceGroups) != 1 {
		t.Fatalf("ReferenceGroups length = %d, want 1", len(canvas.ReferenceGroups))
	}
	if len(canvas.ReferenceGroups[0].InputNodeIDs) < 4 {
		t.Fatalf("ReferenceGroup inputs = %#v, want script/context nodes", canvas.ReferenceGroups[0].InputNodeIDs)
	}
	if len(canvas.Frames) != 1 {
		t.Fatalf("Frames length = %d, want 1", len(canvas.Frames))
	}
	if len(canvas.Frames[0].OutputNodeIDs) < 3 {
		t.Fatalf("Frame outputs = %#v, want package/result/review outputs", canvas.Frames[0].OutputNodeIDs)
	}
	if canvas.Frames[0].HistorySummary.CurrentRunID != "prompts/runs/mock-run-001.json" {
		t.Fatalf("CurrentRunID = %q, want prompt run ref path", canvas.Frames[0].HistorySummary.CurrentRunID)
	}

	for _, edge := range canvas.Edges {
		if edge.Validity != GraphEdgeValidityValid {
			t.Fatalf("edge %q validity = %q error=%#v, want valid", edge.ID, edge.Validity, edge.Error)
		}
		if edge.SourceNodeID == "node_run_001" || edge.TargetNodeID == "node_run_001" {
			t.Fatalf("edge %q uses PromptRun node endpoint: %#v", edge.ID, edge)
		}
	}
}

func TestGraphViewBoundaryNoFullDomainLeak(t *testing.T) {
	root := exampleFixtureRoot(t)
	store := testStore(time.Date(2026, 5, 20, 6, 31, 0, 0, time.UTC), "graph-boundary")

	result := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "graph-boundary"})
	if !result.OK || result.Canvas == nil {
		t.Fatalf("GraphView() = %#v, want canvas", result)
	}

	data, err := json.Marshal(result.Canvas)
	if err != nil {
		t.Fatalf("marshal canvas: %v", err)
	}
	text := string(data)
	for _, forbidden := range []string{
		"夜雨刚停",
		"visualRules",
		"referencePaths",
		"continuityRuleIds",
		"Reserved for downstream review flow verification.",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("Graph View DTO leaked full domain content/key %q in %s", forbidden, text)
		}
	}

	for _, node := range result.Canvas.Nodes {
		if len(node.Data) > 3 {
			t.Fatalf("node %q data = %#v, want compact display summary only", node.ID, node.Data)
		}
		if node.RefID == "" && node.Kind != "note" {
			t.Fatalf("node %q kind=%q missing refId", node.ID, node.Kind)
		}
	}
}

func TestGraphViewBoundaryRejectsUnsafeRefID(t *testing.T) {
	root := copyExampleFixture(t)
	outsidePath := filepath.Join(filepath.Dir(root), "outside-character.json")
	if err := os.WriteFile(outsidePath, []byte(`{"displayName":"External Secret Character","shortDescription":"external leak"}`+"\n"), 0o644); err != nil {
		t.Fatalf("write outside fixture: %v", err)
	}

	if err := copyGraphViewManifest(root, func(manifest *Manifest) {
		for index := range manifest.Graph.Nodes {
			if manifest.Graph.Nodes[index].ID == "node_character_mina" {
				manifest.Graph.Nodes[index].RefID = "../outside-character.json"
			}
		}
	}); err != nil {
		t.Fatalf("mutate manifest: %v", err)
	}

	store := testStore(time.Date(2026, 5, 20, 6, 31, 30, 0, time.UTC), "graph-boundary-unsafe")
	result := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "graph-boundary-unsafe"})
	if !result.OK || result.Canvas == nil {
		t.Fatalf("GraphView() = %#v, want degraded canvas", result)
	}
	if result.Health == nil || !result.Health.HasBlocking() {
		t.Fatalf("GraphView() health = %#v, want blocking path issue", result.Health)
	}
	if !hasOperationError(result.Errors, CodeGraphViewInvalidRef) {
		t.Fatalf("GraphView errors = %#v, want %s", result.Errors, CodeGraphViewInvalidRef)
	}

	node := requireNode(t, result.Canvas.Nodes, "node_character_mina", "character", "continuity", "node_character_mina")
	if node.RefID != "" {
		t.Fatalf("unsafe ref node RefID = %q, want empty", node.RefID)
	}
	if !containsString(node.Badges, "missing_ref") {
		t.Fatalf("unsafe ref node badges = %#v, want missing_ref", node.Badges)
	}

	data, err := json.Marshal(result.Canvas)
	if err != nil {
		t.Fatalf("marshal canvas: %v", err)
	}
	for _, forbidden := range []string{"External Secret Character", "external leak", "../outside-character.json"} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("Graph View DTO leaked unsafe ref content/path %q in %s", forbidden, data)
		}
	}
}

func TestGraphRelationProjectionValidity(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 6, 32, 0, 0, time.UTC), "graph-relations")

	if err := copyGraphViewManifest(root, func(manifest *Manifest) {
		manifest.Graph.Edges = append(manifest.Graph.Edges,
			Edge{
				ID:     "edge_invalid_relation",
				Source: "node_character_mina",
				Target: "node_scene_laneway",
				Kind:   "result_of",
			},
			Edge{
				ID:     "edge_missing_endpoint",
				Source: "node_shot_001",
				Target: "node_missing",
				Kind:   "uses",
			},
			Edge{
				ID:     "edge_prompt_run_to_shot",
				Source: "node_run_001",
				Target: "node_shot_001",
				Kind:   "uses",
			},
		)
	}); err != nil {
		t.Fatalf("mutate manifest: %v", err)
	}

	result := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "graph-relations"})
	if !result.OK || result.Canvas == nil {
		t.Fatalf("GraphView() = %#v, want degraded canvas", result)
	}

	invalid := requireEdge(t, result.Canvas.Edges, "edge_invalid_relation")
	if invalid.Validity != GraphEdgeValidityInvalidRelation {
		t.Fatalf("invalid relation edge validity = %q, want %s", invalid.Validity, GraphEdgeValidityInvalidRelation)
	}
	if invalid.Error == nil || invalid.Error.Code != CodeGraphViewInvalidRelation {
		t.Fatalf("invalid relation error = %#v, want %s", invalid.Error, CodeGraphViewInvalidRelation)
	}

	missing := requireEdge(t, result.Canvas.Edges, "edge_missing_endpoint")
	if missing.Validity != GraphEdgeValidityMissingEndpoint {
		t.Fatalf("missing endpoint edge validity = %q, want %s", missing.Validity, GraphEdgeValidityMissingEndpoint)
	}
	if missing.Error == nil || missing.Error.Code != CodeGraphViewMissingEndpoint {
		t.Fatalf("missing endpoint error = %#v, want %s", missing.Error, CodeGraphViewMissingEndpoint)
	}
	if !hasOperationError(result.Errors, CodeGraphViewPromptRunEdge) {
		t.Fatalf("GraphView errors = %#v, want %s", result.Errors, CodeGraphViewPromptRunEdge)
	}
	for _, edge := range result.Canvas.Edges {
		if edge.SourceNodeID == "node_run_001" || edge.TargetNodeID == "node_run_001" {
			t.Fatalf("PromptRun endpoint leaked into Canvas edge DTO: %#v", edge)
		}
	}
}

func TestGraphViewErrorMapping(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 6, 33, 0, 0, time.UTC), "graph-errors")

	stale := store.GraphView(GraphViewCommand{
		Root:                 root,
		ExpectedGraphVersion: 2,
		CorrelationID:        "graph-stale",
	})
	if stale.OK || stale.Error == nil || stale.Error.Code != CodeGraphViewStaleVersion {
		t.Fatalf("stale GraphView() = %#v, want %s", stale, CodeGraphViewStaleVersion)
	}
	if len(stale.Error.RecoveryActions) == 0 {
		t.Fatal("stale graph error must include recovery actions")
	}

	if err := copyGraphViewManifest(root, func(manifest *Manifest) {
		manifest.Graph.Nodes = append(manifest.Graph.Nodes, Node{
			ID:     "node_unknown_kind",
			Kind:   "unsupported_canvas_kind",
			RefID:  "assets/refs/character-mina.txt",
			Status: "ready",
		})
	}); err != nil {
		t.Fatalf("mutate manifest: %v", err)
	}

	result := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "graph-unsupported"})
	if !result.OK {
		t.Fatalf("GraphView() with unsupported node = %#v, want degraded success", result)
	}
	if !hasOperationError(result.Errors, CodeGraphViewUnsupportedNode) {
		t.Fatalf("GraphView errors = %#v, want %s", result.Errors, CodeGraphViewUnsupportedNode)
	}
	node := requireNode(t, result.Canvas.Nodes, "node_unknown_kind", "note", "review", "Character Mina")
	if !containsString(node.Badges, "missing_ref") {
		t.Fatalf("unsupported node badges = %#v, want missing_ref", node.Badges)
	}
}

func requireNode(t *testing.T, nodes []GraphNodeDTO, id string, kind string, category string, title string) GraphNodeDTO {
	t.Helper()
	for _, node := range nodes {
		if node.ID != id {
			continue
		}
		if node.Kind != kind || node.Category != category || node.Title != title {
			t.Fatalf("node %q = kind=%q category=%q title=%q, want %q/%q/%q", id, node.Kind, node.Category, node.Title, kind, category, title)
		}
		return node
	}
	t.Fatalf("node %q missing from %#v", id, nodeIDs(nodes))
	return GraphNodeDTO{}
}

func requireEdge(t *testing.T, edges []GraphEdgeDTO, id string) GraphEdgeDTO {
	t.Helper()
	for _, edge := range edges {
		if edge.ID == id {
			return edge
		}
	}
	t.Fatalf("edge %q missing from %#v", id, edgeIDs(edges))
	return GraphEdgeDTO{}
}

func requireNoNode(t *testing.T, nodes []GraphNodeDTO, id string) {
	t.Helper()
	for _, node := range nodes {
		if node.ID == id {
			t.Fatalf("node %q should not be projected as a GraphNodeDTO", id)
		}
	}
}

func nodeIDs(nodes []GraphNodeDTO) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	sort.Strings(ids)
	return ids
}

func edgeIDs(edges []GraphEdgeDTO) []string {
	ids := make([]string, 0, len(edges))
	for _, edge := range edges {
		ids = append(ids, edge.ID)
	}
	sort.Strings(ids)
	return ids
}

func hasOperationError(errors []OperationError, code string) bool {
	for _, err := range errors {
		if err.Code == code {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func jsonTagFields(value reflect.Type) []string {
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	fields := make([]string, 0, value.NumField())
	for index := 0; index < value.NumField(); index++ {
		tag := value.Field(index).Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		fields = append(fields, name)
	}
	sort.Strings(fields)
	return fields
}

func copyGraphViewManifest(root string, mutator func(*Manifest)) error {
	data, err := os.ReadFile(filepath.Join(root, ManifestFileName))
	if err != nil {
		return err
	}
	manifest, report := DecodeManifest(data)
	if report.HasBlocking() {
		return ValidationError{Report: report}
	}
	mutator(&manifest)
	next, err := EncodeManifest(manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, ManifestFileName), next, 0o644)
}
