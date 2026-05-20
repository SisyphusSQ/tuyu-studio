package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	CodeGraphViewStaleVersion    = "graph_view_stale_version"
	CodeGraphViewUnsupportedNode = "graph_view_unsupported_node_kind"
	CodeGraphViewInvalidRelation = "graph_view_invalid_relation"
	CodeGraphViewMissingEndpoint = "graph_view_missing_endpoint"
	CodeGraphViewInvalidRef      = "graph_view_invalid_ref"
	CodeGraphViewPromptRunEdge   = "graph_view_prompt_run_endpoint"
)

const (
	GraphEdgeValidityValid           = "valid"
	GraphEdgeValidityInvalidRelation = "invalid_relation"
	GraphEdgeValidityMissingEndpoint = "missing_endpoint"
)

type GraphViewCommand struct {
	Root                 string `json:"root"`
	ExpectedGraphVersion int    `json:"expectedGraphVersion,omitempty"`
	CorrelationID        string `json:"correlationId"`
}

type GraphViewResult struct {
	OK     bool              `json:"ok"`
	Canvas *ProjectCanvasDTO `json:"canvas,omitempty"`
	Health *HealthReport     `json:"health,omitempty"`
	Error  *OperationError   `json:"error,omitempty"`
	Errors []OperationError  `json:"errors"`
	Events []ProjectEvent    `json:"events"`
}

type ProjectCanvasDTO struct {
	ID              string               `json:"id"`
	ProjectID       string               `json:"projectId"`
	SchemaVersion   string               `json:"schemaVersion"`
	Version         int                  `json:"version"`
	Viewport        CanvasViewportDTO    `json:"viewport"`
	Theme           string               `json:"theme"`
	Grid            CanvasGridDTO        `json:"grid"`
	Nodes           []GraphNodeDTO       `json:"nodes"`
	Edges           []GraphEdgeDTO       `json:"edges"`
	Frames          []ProductionFrameDTO `json:"frames"`
	ReferenceGroups []ReferenceGroupDTO  `json:"referenceGroups"`
	SelectedNodeIDs []string             `json:"selectedNodeIds"`
	UpdatedAt       string               `json:"updatedAt"`
}

type CanvasViewportDTO struct {
	X    int     `json:"x"`
	Y    int     `json:"y"`
	Zoom float64 `json:"zoom"`
}

type CanvasGridDTO struct {
	Visible bool    `json:"visible"`
	Size    int     `json:"size"`
	Opacity float64 `json:"opacity"`
}

type GraphNodeDTO struct {
	ID            string            `json:"id"`
	Kind          string            `json:"kind"`
	Category      string            `json:"category"`
	Title         string            `json:"title"`
	RefID         string            `json:"refId,omitempty"`
	Position      CanvasPositionDTO `json:"position"`
	Size          CanvasSizeDTO     `json:"size"`
	Collapsed     bool              `json:"collapsed"`
	Status        string            `json:"status,omitempty"`
	Badges        []string          `json:"badges"`
	Source        string            `json:"source"`
	SourceEventID string            `json:"sourceEventId,omitempty"`
	Data          map[string]string `json:"data,omitempty"`
}

type CanvasPositionDTO struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type CanvasSizeDTO struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type GraphEdgeDTO struct {
	ID           string          `json:"id"`
	SourceNodeID string          `json:"sourceNodeId"`
	TargetNodeID string          `json:"targetNodeId"`
	Relation     string          `json:"relation"`
	Label        string          `json:"label"`
	CreatedAt    string          `json:"createdAt"`
	Validity     string          `json:"validity"`
	Error        *OperationError `json:"error,omitempty"`
}

type ReferenceGroupDTO struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Role         string          `json:"role"`
	InputNodeIDs []string        `json:"inputNodeIds"`
	Priority     int             `json:"priority"`
	Notes        string          `json:"notes,omitempty"`
	Layout       CanvasLayoutDTO `json:"layout"`
}

type ProductionFrameDTO struct {
	ID                   string                 `json:"id"`
	Title                string                 `json:"title"`
	ReferenceGroupIDs    []string               `json:"referenceGroupIds"`
	OutputNodeIDs        []string               `json:"outputNodeIds"`
	TaskIntent           string                 `json:"taskIntent"`
	RequiredCapabilities []string               `json:"requiredCapabilities"`
	HistorySummary       FrameHistorySummaryDTO `json:"historySummary"`
	Layout               CanvasLayoutDTO        `json:"layout"`
}

type FrameHistorySummaryDTO struct {
	CurrentRunID          string   `json:"currentRunId,omitempty"`
	FavoriteRunIDs        []string `json:"favoriteRunIds"`
	LatestSuccessfulRunID string   `json:"latestSuccessfulRunId,omitempty"`
}

type CanvasLayoutDTO struct {
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Collapsed bool `json:"collapsed"`
}

type graphProjection struct {
	canvas ProjectCanvasDTO
	errors []OperationError
}

type nodeDisplaySummary struct {
	title  string
	data   map[string]string
	source string
}

func (s *Store) GraphView(command GraphViewCommand) GraphViewResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.graphViewFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before loading Graph View.")
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		health := s.HealthReport(root)
		operationError := s.errorFromReadFailure(report, err, correlationID)
		return GraphViewResult{
			OK:     false,
			Health: &health,
			Error:  &operationError,
			Errors: []OperationError{operationError},
			Events: []ProjectEvent{s.event("project.graph_view", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	health := s.HealthReport(root)
	if command.ExpectedGraphVersion > 0 && command.ExpectedGraphVersion != manifest.Graph.Version {
		operationError := s.operationError(
			CodeGraphViewStaleVersion,
			SeverityBlocking,
			true,
			correlationID,
			"Graph View request used a stale graph version.",
			fmt.Sprintf("expected=%d actual=%d", command.ExpectedGraphVersion, manifest.Graph.Version),
			[]string{"Reload the project graph before requesting or saving Canvas state."},
		)
		return GraphViewResult{
			OK:     false,
			Health: &health,
			Error:  &operationError,
			Errors: []OperationError{operationError},
			Events: []ProjectEvent{s.event("project.graph_view", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	projection := s.projectGraphView(root, manifest, health, correlationID)
	state := "completed"
	if health.HasBlocking() || hasBlockingOperationError(projection.errors) {
		state = "blocked"
	}

	return GraphViewResult{
		OK:     true,
		Canvas: &projection.canvas,
		Health: &health,
		Errors: projection.errors,
		Events: []ProjectEvent{s.event("project.graph_view.loaded", state, "Project Graph View DTO loaded.", correlationID, nil)},
	}
}

func (s *Store) graphViewFailure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string) GraphViewResult {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, []string{technicalDetail})
	if code == CodeProjectRootRequired {
		err.RecoveryActions = []string{technicalDetail}
	}
	return GraphViewResult{
		OK:     false,
		Error:  &err,
		Errors: []OperationError{err},
		Events: []ProjectEvent{s.event("project.graph_view", "blocked", userMessage, correlationID, &err)},
	}
}

func (s *Store) projectGraphView(root string, manifest Manifest, health HealthReport, correlationID string) graphProjection {
	nodeIssues := healthItemsByAffectedObject(health.Items)
	nodes := make([]GraphNodeDTO, 0, len(manifest.Graph.Nodes)+2)
	nodeKinds := make(map[string]string, len(manifest.Graph.Nodes)+2)
	promptRunNodeIDs := make(map[string]struct{})
	var errors []OperationError

	for index, node := range manifest.Graph.Nodes {
		if strings.TrimSpace(node.Kind) == "prompt_run" {
			if id := strings.TrimSpace(node.ID); id != "" {
				promptRunNodeIDs[id] = struct{}{}
			}
			continue
		}
		dto, err := s.graphNodeDTO(root, manifest, node, index, nodeIssues[node.ID], correlationID)
		if err != nil {
			errors = append(errors, *err)
		}
		nodes = append(nodes, dto)
		nodeKinds[dto.ID] = dto.Kind
	}

	resultNodes, resultEdges := s.syntheticResultNodes(root, manifest, len(nodes), correlationID)
	for _, node := range resultNodes {
		nodes = append(nodes, node)
		nodeKinds[node.ID] = node.Kind
	}

	edges := make([]GraphEdgeDTO, 0, len(manifest.Graph.Edges)+len(resultEdges))
	for _, edge := range manifest.Graph.Edges {
		if err := s.promptRunEndpointError(edge, promptRunNodeIDs, correlationID); err != nil {
			errors = append(errors, *err)
			continue
		}
		dto := s.graphEdgeDTO(edge, nodeKinds, manifest.Project.UpdatedAt, correlationID)
		if dto.Error != nil {
			errors = append(errors, *dto.Error)
		}
		edges = append(edges, dto)
	}
	edges = append(edges, resultEdges...)

	referenceGroups := buildReferenceGroups(nodes)
	frames := buildProductionFrames(manifest, nodes, referenceGroups)

	canvas := ProjectCanvasDTO{
		ID:            manifest.Project.ID + "_canvas",
		ProjectID:     manifest.Project.ID,
		SchemaVersion: manifest.SchemaVersion,
		Version:       manifest.Graph.Version,
		Viewport: CanvasViewportDTO{
			X:    manifest.Graph.Viewport.X,
			Y:    manifest.Graph.Viewport.Y,
			Zoom: manifest.Graph.Viewport.Zoom,
		},
		Theme:           canvasTheme(manifest.Graph.Theme),
		Grid:            canvasGrid(manifest.Graph.Grid),
		Nodes:           nodes,
		Edges:           edges,
		Frames:          frames,
		ReferenceGroups: referenceGroups,
		SelectedNodeIDs: []string{},
		UpdatedAt:       manifest.Project.UpdatedAt,
	}

	return graphProjection{canvas: canvas, errors: errors}
}

func (s *Store) graphNodeDTO(root string, manifest Manifest, node Node, index int, issues []HealthItem, correlationID string) (GraphNodeDTO, *OperationError) {
	kind, category, supported := graphNodeKindCategory(node.Kind)
	refID := strings.TrimSpace(node.RefID)
	filename, canReadRef := safeProjectFilePath(root, refID)
	invalidRef := refID != "" && !canReadRef
	summary := s.nodeDisplaySummary(root, manifest, filename, refID, node, kind, canReadRef)
	if summary.source == "" {
		summary.source = "imported"
	}
	if !supported {
		summary.title = titleFallback(node.RefID)
		summary.data = map[string]string{
			"path":    refID,
			"summary": "Unsupported graph node kind: " + strings.TrimSpace(node.Kind),
		}
	}
	if invalidRef {
		summary.title = safeGraphID(node.ID, "node_"+strconv.Itoa(index+1))
		summary.data = map[string]string{
			"path":    "invalid_ref",
			"summary": "Graph node reference is outside the project-relative file boundary.",
		}
	}

	badges := statusBadges(kind, node.Status)
	if kind == "shot" && canReadRef {
		shotBadges, missingFields := s.shotProjectionBadges(root, manifest, filename)
		badges = append(badges, shotBadges...)
		if len(missingFields) > 0 {
			summary.data["missingFields"] = strings.Join(missingFields, ", ")
		}
	}
	if (kind == BindingTargetCharacter || kind == BindingTargetScene || kind == BindingTargetProp) && canReadRef {
		if strings.TrimSpace(summary.data["mainReferenceAssetId"]) != "" {
			badges = append(badges, "main_reference")
		} else {
			badges = append(badges, "main_reference_missing")
		}
	}
	for _, issue := range issues {
		switch issue.Code {
		case CodeGraphReferenceMissing, CodeProjectManifestMissing:
			badges = append(badges, "missing_ref")
		case CodeAssetMissing:
			badges = append(badges, "missing_asset")
		case CodeDigestMismatch:
			badges = append(badges, "context_dirty")
		}
	}
	if !supported {
		kind = "note"
		category = "review"
		badges = append(badges, "missing_ref")
	}
	if invalidRef {
		badges = append(badges, "missing_ref")
	}
	badges = dedupeSortedStrings(badges)

	dtoRefID := refID
	if invalidRef {
		dtoRefID = ""
	}
	dto := GraphNodeDTO{
		ID:        safeGraphID(node.ID, "node_"+strconv.Itoa(index+1)),
		Kind:      kind,
		Category:  category,
		Title:     summary.title,
		RefID:     dtoRefID,
		Position:  canvasPosition(node.Position, positionForNode(category, lenByCategoryOffset(index, category))),
		Size:      canvasSize(node.Size, sizeForNode(kind)),
		Collapsed: canvasCollapsed(node.Collapsed),
		Status:    strings.TrimSpace(node.Status),
		Badges:    badges,
		Source:    summary.source,
		Data:      summary.data,
	}
	if dto.Title == "" {
		dto.Title = dto.ID
	}

	if invalidRef {
		err := s.operationError(
			CodeGraphViewInvalidRef,
			SeverityBlocking,
			false,
			correlationID,
			"Graph node reference is outside the project boundary.",
			"node="+dto.ID+" refId failed project-relative validation",
			[]string{"Relink the node to a project-relative file before using it in Canvas or export flows."},
		)
		err.TargetID = dto.ID
		return dto, &err
	}

	if supported {
		return dto, nil
	}

	err := s.operationError(
		CodeGraphViewUnsupportedNode,
		SeverityWarning,
		false,
		correlationID,
		"Graph View encountered an unsupported node kind.",
		"node="+dto.ID+" kind="+node.Kind,
		[]string{"Map the node kind to a supported Canvas kind before enabling production actions."},
	)
	err.TargetID = dto.ID
	return dto, &err
}

func canvasTheme(raw string) string {
	switch strings.TrimSpace(raw) {
	case "warm_light":
		return "warm_light"
	default:
		return "dark"
	}
}

func canvasGrid(grid *GraphGrid) CanvasGridDTO {
	dto := CanvasGridDTO{
		Visible: true,
		Size:    24,
		Opacity: 0.24,
	}
	if grid == nil {
		return dto
	}
	dto.Visible = grid.Visible
	if grid.Size > 0 {
		dto.Size = grid.Size
	}
	if grid.Opacity >= 0 && grid.Opacity <= 1 {
		dto.Opacity = grid.Opacity
	}
	return dto
}

func canvasPosition(position *GraphPosition, fallback CanvasPositionDTO) CanvasPositionDTO {
	if position == nil {
		return fallback
	}
	return CanvasPositionDTO{X: position.X, Y: position.Y}
}

func canvasSize(size *GraphSize, fallback CanvasSizeDTO) CanvasSizeDTO {
	if size == nil || size.Width <= 0 || size.Height <= 0 {
		return fallback
	}
	return CanvasSizeDTO{Width: size.Width, Height: size.Height}
}

func canvasCollapsed(collapsed *bool) bool {
	if collapsed == nil {
		return false
	}
	return *collapsed
}

func (s *Store) promptRunEndpointError(edge Edge, promptRunNodeIDs map[string]struct{}, correlationID string) *OperationError {
	source := strings.TrimSpace(edge.Source)
	target := strings.TrimSpace(edge.Target)
	_, sourceIsPromptRun := promptRunNodeIDs[source]
	_, targetIsPromptRun := promptRunNodeIDs[target]
	if !sourceIsPromptRun && !targetIsPromptRun {
		return nil
	}

	edgeID := safeGraphID(edge.ID, "edge_"+safeFileToken(source+"_"+target))
	err := s.operationError(
		CodeGraphViewPromptRunEdge,
		SeverityWarning,
		false,
		correlationID,
		"Graph edge references a PromptRun, which is not a Canvas node endpoint.",
		"edge="+edgeID+" references prompt_run endpoint",
		[]string{"Move run references into frame history or project run records instead of Canvas edges."},
	)
	err.TargetID = edgeID
	return &err
}

func (s *Store) graphEdgeDTO(edge Edge, nodeKinds map[string]string, updatedAt string, correlationID string) GraphEdgeDTO {
	source := strings.TrimSpace(edge.Source)
	target := strings.TrimSpace(edge.Target)
	relation, label, normalizedSource, normalizedTarget := normalizeManifestRelation(edge.Kind, source, target)

	dto := GraphEdgeDTO{
		ID:           safeGraphID(edge.ID, "edge_"+safeFileToken(source+"_"+target)),
		SourceNodeID: normalizedSource,
		TargetNodeID: normalizedTarget,
		Relation:     relation,
		Label:        label,
		CreatedAt:    updatedAt,
		Validity:     GraphEdgeValidityValid,
	}

	sourceKind, sourceOK := nodeKinds[normalizedSource]
	targetKind, targetOK := nodeKinds[normalizedTarget]
	if !sourceOK || !targetOK {
		err := s.operationError(
			CodeGraphViewMissingEndpoint,
			SeverityWarning,
			true,
			correlationID,
			"Graph edge references a missing Canvas node.",
			"edge="+dto.ID+" source="+normalizedSource+" target="+normalizedTarget,
			[]string{"Restore the missing node or remove the edge after reviewing project references."},
		)
		err.TargetID = dto.ID
		dto.Validity = GraphEdgeValidityMissingEndpoint
		dto.Error = &err
		return dto
	}

	if !relationAllowed(sourceKind, targetKind, relation) {
		err := s.operationError(
			CodeGraphViewInvalidRelation,
			SeverityWarning,
			false,
			correlationID,
			"Graph edge relation is not valid for these node kinds.",
			"edge="+dto.ID+" "+sourceKind+"->"+targetKind+" relation="+relation,
			[]string{"Rebuild the edge with a relation allowed by the Creative Graph matrix."},
		)
		err.TargetID = dto.ID
		dto.Validity = GraphEdgeValidityInvalidRelation
		dto.Error = &err
	}

	return dto
}

func (s *Store) syntheticResultNodes(root string, manifest Manifest, offset int, correlationID string) ([]GraphNodeDTO, []GraphEdgeDTO) {
	resultsRoot := filepath.Join(root, filepath.FromSlash(manifest.Paths.AssetResults))
	entries, err := os.ReadDir(resultsRoot)
	if err != nil {
		return nil, nil
	}

	packageNodeID := firstNodeIDByKind(manifest.Graph.Nodes, "package")
	var nodes []GraphNodeDTO
	var edges []GraphEdgeDTO
	resultIndex := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		relative := path.Clean(manifest.Paths.AssetResults + "/" + entry.Name())
		title, summary := readMarkdownTitleAndStatus(filepath.Join(resultsRoot, entry.Name()), "Video Result")
		id := "node_result_" + safeFileToken(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
		node := GraphNodeDTO{
			ID:        id,
			Kind:      "video_result",
			Category:  "output",
			Title:     title,
			RefID:     relative,
			Position:  positionForNode("output", offset+resultIndex),
			Size:      sizeForNode("video_result"),
			Collapsed: false,
			Status:    "pending_review",
			Badges:    []string{"pending_review"},
			Source:    "system",
			Data: map[string]string{
				"path":    relative,
				"summary": summary,
			},
		}
		nodes = append(nodes, node)
		if packageNodeID != "" {
			edgeID := "edge_" + id + "_to_" + packageNodeID
			edge := GraphEdgeDTO{
				ID:           edgeID,
				SourceNodeID: id,
				TargetNodeID: packageNodeID,
				Relation:     "result_of",
				Label:        "result of package",
				CreatedAt:    manifest.Project.UpdatedAt,
				Validity:     GraphEdgeValidityValid,
			}
			if !relationAllowed("video_result", "package", "result_of") {
				err := s.operationError(
					CodeGraphViewInvalidRelation,
					SeverityWarning,
					false,
					correlationID,
					"Graph edge relation is not valid for these node kinds.",
					"edge="+edgeID+" video_result->package relation=result_of",
					[]string{"Rebuild the result relation with a supported endpoint."},
				)
				err.TargetID = edgeID
				edge.Validity = GraphEdgeValidityInvalidRelation
				edge.Error = &err
			}
			edges = append(edges, edge)
		}
		resultIndex++
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	return nodes, edges
}

func (s *Store) nodeDisplaySummary(root string, manifest Manifest, filename string, relative string, node Node, kind string, canRead bool) nodeDisplaySummary {
	data := map[string]string{
		"path": relative,
	}
	if relative == "" {
		return nodeDisplaySummary{title: node.ID, data: data, source: "human"}
	}
	if !canRead {
		data["path"] = "invalid_ref"
		data["summary"] = "Reference is unavailable to the Graph View projection."
		return nodeDisplaySummary{title: safeGraphID(node.ID, titleFallback(relative)), data: data, source: "imported"}
	}

	switch kind {
	case "script":
		title, _ := readMarkdownTitleAndStatus(filename, "Script source")
		data["summary"] = "Script source body is excluded from the Canvas DTO."
		return nodeDisplaySummary{title: title, data: data, source: "imported"}
	case "character", "scene", "prop":
		title, summary := readJSONSummary(filename, "displayName", "shortDescription", titleFallback(relative))
		data["summary"] = summary
		s.addProfileNodeData(root, manifest, filename, relative, kind, data)
		return nodeDisplaySummary{title: title, data: data, source: "imported"}
	case "shot":
		title, summary := readJSONSummary(filename, "title", "description", titleFallback(relative))
		data["summary"] = summary
		data["role"] = "production_unit"
		return nodeDisplaySummary{title: title, data: data, source: "human"}
	case "package":
		title, summary := readPackageSummary(filename)
		data["summary"] = summary
		return nodeDisplaySummary{title: title, data: data, source: "system"}
	case "note":
		title, summary := readReviewSummary(filename)
		data["summary"] = summary
		return nodeDisplaySummary{title: title, data: data, source: "human"}
	default:
		data["summary"] = "Unsupported graph node kind: " + strings.TrimSpace(node.Kind)
		return nodeDisplaySummary{title: titleFallback(relative), data: data, source: "imported"}
	}
}

func (s *Store) addProfileNodeData(root string, manifest Manifest, filename string, relative string, kind string, data map[string]string) {
	record, err := readProfileRecordFile(filename, relative, kind)
	if err != nil {
		return
	}
	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return
	}
	assets := assetMapByID(s.hydrateAssetsForList(root, index.Assets))
	profile := profileDTOWithBindings(record.dto, index, map[string]profileRecord{record.dto.key(): record}, assets)
	data["profileId"] = profile.ID
	data["referenceAssetIds"] = strings.Join(profile.ReferenceAssetIDs, ", ")
	data["bindingCount"] = strconv.Itoa(profile.BindingCount)
	if profile.MainReferenceAssetID != "" {
		data["mainReferenceAssetId"] = profile.MainReferenceAssetID
		data["mainReferencePath"] = profile.MainReferencePath
		data["mainReferenceThumbnailPath"] = profile.MainReferenceThumbnailPath
	}
	if profile.MissingMainReference {
		data["mainReferenceStatus"] = "missing"
	} else if profile.MainReferenceAssetID != "" {
		data["mainReferenceStatus"] = "ready"
	} else {
		data["mainReferenceStatus"] = "empty"
	}
}

func graphNodeKindCategory(raw string) (string, string, bool) {
	switch strings.TrimSpace(raw) {
	case "script_document", "script":
		return "script", "source", true
	case "scene":
		return "scene", "concept", true
	case "character":
		return "character", "continuity", true
	case "prop":
		return "prop", "continuity", true
	case "style":
		return "style", "concept", true
	case "shot":
		return "shot", "production", true
	case "prompt":
		return "prompt", "production", true
	case "fusion":
		return "fusion", "production", true
	case "package":
		return "package", "handoff", true
	case "video_result":
		return "video_result", "output", true
	case "review", "note":
		return "note", "review", true
	case "prompt_run":
		return "note", "review", false
	default:
		return "note", "review", false
	}
}

func normalizeManifestRelation(raw string, source string, target string) (relation string, label string, normalizedSource string, normalizedTarget string) {
	switch strings.TrimSpace(raw) {
	case "candidate_source":
		return "uses", "candidate source", source, target
	case "context":
		return "uses", "uses context", target, source
	case "handoff_candidate":
		return "packaged_as", "packaged as", source, target
	case "review_target":
		return "uses", "review target", source, target
	case "belongs_to", "generated_by", "refines", "packaged_as", "result_of", "depends_on", "uses":
		return raw, raw, source, target
	default:
		if strings.TrimSpace(raw) == "" {
			return "uses", "unspecified relation", source, target
		}
		return raw, raw, source, target
	}
}

func relationAllowed(sourceKind string, targetKind string, relation string) bool {
	allowed := map[string]map[string][]string{
		"script": {
			"scene": {"belongs_to"},
			"shot":  {"uses"},
			"note":  {"uses"},
		},
		"scene": {
			"script": {"uses"},
			"prop":   {"uses"},
			"style":  {"uses"},
			"note":   {"uses"},
		},
		"character": {
			"prop":  {"uses"},
			"style": {"uses"},
			"note":  {"uses"},
		},
		"prop": {
			"style": {"uses"},
			"note":  {"uses"},
		},
		"style": {
			"note": {"uses"},
		},
		"shot": {
			"script":    {"uses"},
			"scene":     {"belongs_to", "uses"},
			"character": {"uses"},
			"prop":      {"uses"},
			"style":     {"uses"},
			"shot":      {"refines"},
			"prompt":    {"uses"},
			"fusion":    {"depends_on"},
			"package":   {"packaged_as"},
			"note":      {"uses"},
		},
		"prompt": {
			"script":    {"uses"},
			"scene":     {"uses"},
			"character": {"uses"},
			"prop":      {"uses"},
			"style":     {"uses"},
			"shot":      {"belongs_to"},
			"prompt":    {"refines"},
			"fusion":    {"generated_by"},
			"note":      {"uses"},
		},
		"fusion": {
			"scene":     {"depends_on"},
			"character": {"depends_on"},
			"prop":      {"depends_on"},
			"style":     {"depends_on"},
			"shot":      {"depends_on"},
			"note":      {"uses"},
		},
		"package": {
			"scene":     {"depends_on"},
			"character": {"depends_on"},
			"prop":      {"depends_on"},
			"style":     {"depends_on"},
			"shot":      {"depends_on"},
			"package":   {"refines"},
			"note":      {"uses"},
		},
		"video_result": {
			"shot":         {"result_of"},
			"package":      {"result_of"},
			"video_result": {"refines"},
			"note":         {"uses"},
		},
		"note": {
			"script":       {"uses"},
			"scene":        {"uses"},
			"character":    {"uses"},
			"prop":         {"uses"},
			"style":        {"uses"},
			"shot":         {"uses"},
			"prompt":       {"uses"},
			"fusion":       {"uses"},
			"package":      {"uses"},
			"video_result": {"uses"},
		},
	}

	targets, ok := allowed[sourceKind]
	if !ok {
		return false
	}
	relations, ok := targets[targetKind]
	if !ok {
		return false
	}
	for _, candidate := range relations {
		if candidate == relation {
			return true
		}
	}
	return false
}

func buildReferenceGroups(nodes []GraphNodeDTO) []ReferenceGroupDTO {
	var inputNodeIDs []string
	for _, node := range nodes {
		switch node.Category {
		case "source", "concept", "continuity":
			inputNodeIDs = append(inputNodeIDs, node.ID)
		}
	}
	sort.Strings(inputNodeIDs)
	if len(inputNodeIDs) == 0 {
		return []ReferenceGroupDTO{}
	}
	return []ReferenceGroupDTO{{
		ID:           "refgrp_alpha_fixture_context",
		Title:        "Alpha fixture context",
		Role:         "shot_context",
		InputNodeIDs: inputNodeIDs,
		Priority:     1,
		Notes:        "Script, scene, character and prop summaries for the alpha fixture.",
		Layout:       CanvasLayoutDTO{X: 40, Y: 420, Width: 520, Height: 180},
	}}
}

func buildProductionFrames(manifest Manifest, nodes []GraphNodeDTO, groups []ReferenceGroupDTO) []ProductionFrameDTO {
	if len(nodes) == 0 {
		return []ProductionFrameDTO{}
	}

	var groupIDs []string
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID)
	}

	var outputNodeIDs []string
	for _, node := range nodes {
		switch node.Category {
		case "handoff", "output", "review":
			outputNodeIDs = append(outputNodeIDs, node.ID)
		}
	}
	sort.Strings(outputNodeIDs)

	return []ProductionFrameDTO{{
		ID:                "frame_alpha_fixture_shot_001",
		Title:             "Alpha fixture shot 001 production",
		ReferenceGroupIDs: groupIDs,
		OutputNodeIDs:     outputNodeIDs,
		TaskIntent:        "Prepare and review the manual handoff package for the first alpha fixture shot.",
		RequiredCapabilities: []string{
			"project_graph_view",
			"manual_handoff",
		},
		HistorySummary: FrameHistorySummaryDTO{
			CurrentRunID:   firstNodeRefByKind(manifest.Graph.Nodes, "prompt_run"),
			FavoriteRunIDs: []string{},
		},
		Layout: CanvasLayoutDTO{X: 580, Y: 360, Width: 620, Height: 260},
	}}
}

func healthItemsByAffectedObject(items []HealthItem) map[string][]HealthItem {
	byObject := make(map[string][]HealthItem)
	for _, item := range items {
		for _, object := range item.AffectedObjects {
			byObject[object] = append(byObject[object], item)
		}
	}
	return byObject
}

func statusBadges(kind string, status string) []string {
	switch strings.TrimSpace(status) {
	case "ready", "candidate_ready", "context_ready":
		if kind == "package" {
			return []string{"package_ready"}
		}
		return []string{"context_ready"}
	case "draft":
		if kind == "package" {
			return []string{"package_ready"}
		}
		return []string{"context_dirty"}
	case "pending", "pending_review":
		return []string{"pending_review"}
	case "pending_confirmation", "not_submitted":
		return []string{"context_dirty"}
	case "approved":
		return []string{"approved"}
	case "needs_revision":
		return []string{"needs_revision"}
	default:
		return []string{}
	}
}

func (s *Store) shotProjectionBadges(root string, manifest Manifest, filename string) ([]string, []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil
	}
	var shot ShotCardDTO
	if err := json.Unmarshal(data, &shot); err != nil {
		return []string{"blocked"}, nil
	}
	report := s.validateShotContext(root, manifest, shot)
	if report.CanEnterContextReady {
		return nil, nil
	}
	return []string{"missing_context", "blocked"}, shotContextBlockingFields(report)
}

func shotContextBlockingFields(report ShotContextReportDTO) []string {
	values := append([]string{}, report.MissingFields...)
	for _, issue := range report.Blocking {
		values = append(values, issue.Field)
	}
	values = cleanStringList(values)
	sort.Strings(values)
	deduped := values[:0]
	for _, value := range values {
		if len(deduped) == 0 || deduped[len(deduped)-1] != value {
			deduped = append(deduped, value)
		}
	}
	return deduped
}

func positionForNode(category string, offset int) CanvasPositionDTO {
	column := map[string]int{
		"source":     80,
		"concept":    320,
		"continuity": 320,
		"production": 620,
		"handoff":    900,
		"output":     1140,
		"review":     1140,
	}
	x := column[category]
	if x == 0 {
		x = 80
	}
	return CanvasPositionDTO{X: x, Y: 96 + offset*108}
}

func lenByCategoryOffset(index int, category string) int {
	if category == "continuity" || category == "concept" {
		return index % 4
	}
	if category == "production" {
		return index % 3
	}
	if category == "handoff" || category == "output" || category == "review" {
		return index % 5
	}
	return index
}

func sizeForNode(kind string) CanvasSizeDTO {
	switch kind {
	case "shot":
		return CanvasSizeDTO{Width: 240, Height: 112}
	case "package", "video_result":
		return CanvasSizeDTO{Width: 248, Height: 104}
	default:
		return CanvasSizeDTO{Width: 220, Height: 96}
	}
}

func readJSONSummary(filename string, titleField string, summaryField string, fallback string) (string, string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fallback, "Reference file could not be read."
	}

	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return fallback, "Reference file could not be decoded."
	}

	title := stringValue(value[titleField])
	if title == "" {
		title = stringValue(value["id"])
	}
	if title == "" {
		title = fallback
	}

	summary := stringValue(value[summaryField])
	if summary == "" {
		summary = stringValue(value["status"])
	}
	return title, summary
}

func readPackageSummary(filename string) (string, string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "Generation package", "Package manifest could not be read."
	}

	var manifest struct {
		PackageID string `json:"packageId"`
		ShotID    string `json:"shotId"`
		Status    string `json:"generationPackageStatus"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "Generation package", "Package manifest could not be decoded."
	}

	title := "Package"
	if manifest.PackageID != "" {
		title = "Package " + manifest.PackageID
	}
	summary := cleanSummaryParts(manifest.Status, manifest.ShotID)
	return title, summary
}

func readReviewSummary(filename string) (string, string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "Review note", "Review note could not be read."
	}

	var review struct {
		ID       string `json:"id"`
		Decision string `json:"decision"`
		TargetID string `json:"targetId"`
	}
	if err := json.Unmarshal(data, &review); err != nil {
		return "Review note", "Review note could not be decoded."
	}

	title := "Review note"
	if review.ID != "" {
		title = "Review note " + review.ID
	}
	summary := cleanSummaryParts(review.Decision, review.TargetID)
	return title, summary
}

func readMarkdownTitleAndStatus(filename string, fallback string) (string, string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fallback, "Reference file could not be read."
	}

	lines := strings.Split(string(data), "\n")
	title := fallback
	summary := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") && title == fallback {
			title = strings.TrimSpace(strings.TrimLeft(line, "#"))
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "status:") {
			summary = strings.TrimSpace(strings.TrimPrefix(line, "Status:"))
			summary = strings.TrimSpace(strings.TrimPrefix(summary, "status:"))
			break
		}
		if summary == "" && !strings.HasPrefix(line, "#") {
			summary = truncateSummary(line)
		}
	}
	return title, summary
}

func firstNodeIDByKind(nodes []Node, kind string) string {
	for _, node := range nodes {
		if strings.TrimSpace(node.Kind) == kind {
			return strings.TrimSpace(node.ID)
		}
	}
	return ""
}

func firstNodeRefByKind(nodes []Node, kind string) string {
	for _, node := range nodes {
		if strings.TrimSpace(node.Kind) == kind {
			return strings.TrimSpace(node.RefID)
		}
	}
	return ""
}

func titleFallback(relative string) string {
	base := path.Base(strings.TrimSpace(relative))
	base = strings.TrimSuffix(base, path.Ext(base))
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.TrimSpace(base)
	if base == "" || base == "." {
		return "Project object"
	}
	return titleWords(base)
}

func titleWords(value string) string {
	fields := strings.Fields(value)
	for index, field := range fields {
		if field == "" {
			continue
		}
		fields[index] = strings.ToUpper(field[:1]) + field[1:]
	}
	return strings.Join(fields, " ")
}

func safeProjectFilePath(root string, relative string) (string, bool) {
	relative = strings.TrimSpace(relative)
	if relative == "" {
		return "", false
	}
	if issues := validateProjectPath("graph.refId", relative); len(issues) > 0 {
		return "", false
	}

	target := filepath.Join(root, filepath.FromSlash(relative))
	rootPath, ok := evaluatedOrAbsolutePath(root)
	if !ok {
		return "", false
	}

	if targetPath, err := filepath.EvalSymlinks(target); err == nil {
		return target, pathInside(rootPath, targetPath)
	}

	targetPath, err := filepath.Abs(target)
	if err != nil {
		return "", false
	}
	return target, pathInside(rootPath, targetPath)
}

func evaluatedOrAbsolutePath(filename string) (string, bool) {
	if evaluated, err := filepath.EvalSymlinks(filename); err == nil {
		return evaluated, true
	}
	absolute, err := filepath.Abs(filename)
	if err != nil {
		return "", false
	}
	return absolute, true
}

func pathInside(root string, target string) bool {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) && !filepath.IsAbs(relative))
}

func safeGraphID(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return truncateSummary(typed)
	case fmt.Stringer:
		return truncateSummary(typed.String())
	default:
		return ""
	}
}

func cleanSummaryParts(parts ...string) string {
	var cleaned []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	return truncateSummary(strings.Join(cleaned, " · "))
}

func truncateSummary(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= 180 {
		return value
	}
	return value[:177] + "..."
}

func dedupeSortedStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var cleaned []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		cleaned = append(cleaned, value)
	}
	sort.Strings(cleaned)
	return cleaned
}

func hasBlockingOperationError(errors []OperationError) bool {
	for _, err := range errors {
		if err.Severity == SeverityBlocking {
			return true
		}
	}
	return false
}
