package project

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	CodeGraphLayoutInvalidCommand = "graph_layout_invalid_command"
	CodeGraphLayoutSaveCommit     = "graph_layout_save_commit_failed"
)

type SaveGraphLayoutCommand struct {
	Root                 string                   `json:"root"`
	ExpectedGraphVersion int                      `json:"expectedGraphVersion"`
	Viewport             CanvasViewportDTO        `json:"viewport"`
	Theme                string                   `json:"theme"`
	Grid                 CanvasGridDTO            `json:"grid"`
	Nodes                []GraphNodeLayoutCommand `json:"nodes"`
	CorrelationID        string                   `json:"correlationId"`
}

type GraphNodeLayoutCommand struct {
	ID        string            `json:"id"`
	Position  CanvasPositionDTO `json:"position"`
	Size      CanvasSizeDTO     `json:"size,omitempty"`
	Collapsed bool              `json:"collapsed"`
}

func (s *Store) SaveGraphLayout(command SaveGraphLayoutCommand) GraphViewResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.graphLayoutFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before saving Canvas state.")
	}

	if command.Viewport.Zoom <= 0 {
		return s.graphLayoutFailure(CodeGraphLayoutInvalidCommand, SeverityBlocking, false, correlationID, "Canvas viewport zoom must be greater than zero.", "Reload the project graph before saving Canvas state.")
	}
	if command.ExpectedGraphVersion <= 0 {
		return s.graphLayoutFailure(CodeGraphLayoutInvalidCommand, SeverityBlocking, false, correlationID, "Canvas state save requires an expected graph version.", "Reload the project graph before saving Canvas state.")
	}

	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		return s.graphLayoutFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "Project is open in another active session.", "Open the project read-only or confirm takeover before saving Canvas state.")
	}
	if lockInfo.State == LockStateStale {
		return s.graphLayoutFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "Project has a stale lock that needs takeover before saving Canvas state.", "Open the project with takeover, then retry Canvas state save.")
	}

	current, report, err := s.readManifest(root)
	if err != nil {
		health := s.HealthReport(root)
		operationError := s.errorFromReadFailure(report, err, correlationID)
		return GraphViewResult{
			OK:     false,
			Health: &health,
			Error:  &operationError,
			Errors: []OperationError{operationError},
			Events: []ProjectEvent{s.event("project.graph_layout_save", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	if command.ExpectedGraphVersion != current.Graph.Version {
		operationError := s.operationError(
			CodeGraphViewStaleVersion,
			SeverityBlocking,
			true,
			correlationID,
			"Canvas state save used a stale graph version.",
			fmt.Sprintf("expected=%d actual=%d", command.ExpectedGraphVersion, current.Graph.Version),
			[]string{"Reload the project graph before saving Canvas state again."},
		)
		health := s.HealthReport(root)
		return GraphViewResult{
			OK:     false,
			Health: &health,
			Error:  &operationError,
			Errors: []OperationError{operationError},
			Events: []ProjectEvent{s.event("project.graph_layout_save", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	next := current
	layouts := layoutCommandByNodeID(command.Nodes)
	updatedNodes := 0
	for index := range next.Graph.Nodes {
		nodeID := strings.TrimSpace(next.Graph.Nodes[index].ID)
		layout, ok := layouts[nodeID]
		if !ok {
			continue
		}
		next.Graph.Nodes[index].Position = &GraphPosition{X: layout.Position.X, Y: layout.Position.Y}
		if layout.Size.Width > 0 && layout.Size.Height > 0 {
			next.Graph.Nodes[index].Size = &GraphSize{Width: layout.Size.Width, Height: layout.Size.Height}
		}
		collapsed := layout.Collapsed
		next.Graph.Nodes[index].Collapsed = &collapsed
		updatedNodes++
	}

	next.Graph.Viewport = Viewport{X: command.Viewport.X, Y: command.Viewport.Y, Zoom: command.Viewport.Zoom}
	next.Graph.Theme = canvasTheme(command.Theme)
	next.Graph.Grid = graphGrid(command.Grid)
	next.Project.UpdatedAt = s.now().UTC().Format(time.RFC3339)
	if next.Graph.Version <= 0 {
		next.Graph.Version = 1
	}
	next.Graph.Version++
	next.Integrity.LastCleanShutdown = true
	next.Integrity.LastGraphVersion = next.Graph.Version

	data, err := EncodeManifest(next)
	if err != nil {
		operationError := s.operationError(CodeProjectSaveValidation, SeverityBlocking, false, correlationID, "Canvas state was not saved because the manifest failed validation.", err.Error(), []string{"Fix the manifest errors and retry Canvas state save."})
		health := s.HealthReport(root)
		return GraphViewResult{
			OK:     false,
			Health: &health,
			Error:  &operationError,
			Errors: []OperationError{operationError},
			Events: []ProjectEvent{s.event("project.graph_layout_save", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	manifestPath := filepath.Join(root, ManifestFileName)
	if err := s.atomicWrite(manifestPath, data, 0o644, s.instanceID); err != nil {
		userMessage := "Canvas state save failed before replacing the previous manifest."
		recoveryActions := []string{"Retry Canvas state save. The previous manifest remains the active version."}
		var atomicErr atomicWriteError
		if errors.As(err, &atomicErr) && atomicErr.replaced {
			userMessage = "Canvas state save replaced the manifest, but durability could not be confirmed."
			recoveryActions = []string{"Run health check before continuing Canvas work, then retry save if the project reports recovery items."}
		}
		operationError := s.operationError(CodeGraphLayoutSaveCommit, SeverityBlocking, true, correlationID, userMessage, err.Error(), recoveryActions)
		health := s.HealthReport(root)
		return GraphViewResult{
			OK:     false,
			Health: &health,
			Error:  &operationError,
			Errors: []OperationError{operationError},
			Events: []ProjectEvent{s.event("project.graph_layout_save", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("project.graph_layout_saved", correlationID),
		EventType:     "project.graph_layout_saved",
		ProjectID:     next.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     s.now().UTC().Format(time.RFC3339),
		Summary:       "Project Canvas layout saved.",
		Details: map[string]any{
			"graphVersion": next.Graph.Version,
			"updatedNodes": updatedNodes,
			"theme":        next.Graph.Theme,
			"grid":         next.Graph.Grid,
			"viewport":     next.Graph.Viewport,
		},
	})

	result := s.GraphView(GraphViewCommand{
		Root:                 root,
		ExpectedGraphVersion: next.Graph.Version,
		CorrelationID:        correlationID,
	})
	saveEvent := s.event("project.graph_layout_saved", "completed", "Project Canvas layout saved with atomic replace.", correlationID, nil)
	result.Events = append([]ProjectEvent{saveEvent}, result.Events...)
	return result
}

func (s *Store) graphLayoutFailure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string) GraphViewResult {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, []string{technicalDetail})
	if code == CodeProjectRootRequired {
		err.RecoveryActions = []string{technicalDetail}
	}
	return GraphViewResult{
		OK:     false,
		Error:  &err,
		Errors: []OperationError{err},
		Events: []ProjectEvent{s.event("project.graph_layout_save", "blocked", userMessage, correlationID, &err)},
	}
}

func layoutCommandByNodeID(nodes []GraphNodeLayoutCommand) map[string]GraphNodeLayoutCommand {
	layouts := make(map[string]GraphNodeLayoutCommand, len(nodes))
	for _, node := range nodes {
		nodeID := strings.TrimSpace(node.ID)
		if nodeID == "" {
			continue
		}
		layouts[nodeID] = node
	}
	return layouts
}

func graphGrid(grid CanvasGridDTO) *GraphGrid {
	size := grid.Size
	if size <= 0 {
		size = 24
	}
	opacity := grid.Opacity
	if opacity < 0 || opacity > 1 {
		opacity = 0.24
	}
	return &GraphGrid{
		Visible: grid.Visible,
		Size:    size,
		Opacity: opacity,
	}
}
