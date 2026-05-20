package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
)

const (
	LockFileRelativePath      = "locks/session.lock"
	AuditEventsRelativePath   = "audit/project-events.jsonl"
	RecoveryAuditRelativePath = "audit/repairs.jsonl"
	DigestIndexRelativePath   = "audit/digests.json"

	DefaultProjectType = "series"
)

const (
	HealthStatusClean    = "clean"
	HealthStatusWarning  = "warning"
	HealthStatusBlocking = "blocking"
)

const (
	LockStateUnlocked = "unlocked"
	LockStateOwned    = "owned"
	LockStateStale    = "stale"
	LockStateActive   = "active"
	LockStateTakeover = "takeover"
)

const (
	OpenModeReadWrite         = "read_write"
	OpenModeReadOnly          = "read_only"
	OpenModeTakeoverAvailable = "takeover_available"
)

const (
	CodeProjectRootRequired     = "project_root_required"
	CodeProjectManifestMissing  = "project_manifest_missing"
	CodeProjectManifestInvalid  = "project_manifest_invalid"
	CodeProjectSaveValidation   = "project_save_validation_failed"
	CodeProjectSaveCommit       = "project_save_commit_failed"
	CodeProjectCreateFailed     = "project_create_failed"
	CodeProjectLockWriteFailed  = "project_lock_write_failed"
	CodeProjectActiveLock       = "project_active_lock"
	CodeProjectStaleLock        = "project_stale_lock"
	CodeProjectDirtyShutdown    = "project_dirty_shutdown"
	CodeProjectDirectoryMissing = "project_directory_missing"
	CodeAssetMissing            = "asset_missing"
	CodeDigestMismatch          = "digest_mismatch"
	CodeDigestIndexInvalid      = "digest_index_invalid"
	CodeGraphReferenceMissing   = "graph_reference_missing"
	CodePackageManifestInvalid  = "package_manifest_invalid"
	CodePackageReferenceMissing = "package_reference_missing"
	CodeHealthCheckIncomplete   = "health_check_incomplete"
)

type AtomicWriteFunc func(filename string, data []byte, perm fs.FileMode, instanceID string) error

type StoreOptions struct {
	Now         func() time.Time
	InstanceID  string
	PID         int
	Hostname    string
	StaleAfter  time.Duration
	AtomicWrite AtomicWriteFunc
}

type Store struct {
	now         func() time.Time
	instanceID  string
	pid         int
	hostname    string
	staleAfter  time.Duration
	atomicWrite AtomicWriteFunc
}

type CreateProjectCommand struct {
	Root          string `json:"root"`
	ProjectID     string `json:"projectId"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	CorrelationID string `json:"correlationId"`
}

type OpenProjectCommand struct {
	Root          string `json:"root"`
	Takeover      bool   `json:"takeover"`
	CorrelationID string `json:"correlationId"`
}

type SaveProjectCommand struct {
	Root          string `json:"root"`
	CorrelationID string `json:"correlationId"`
}

type CheckProjectHealthCommand struct {
	Root          string `json:"root"`
	CorrelationID string `json:"correlationId"`
}

type OperationResult struct {
	OK      bool            `json:"ok"`
	Summary *ProjectSummary `json:"summary,omitempty"`
	Health  *HealthReport   `json:"health,omitempty"`
	Error   *OperationError `json:"error,omitempty"`
	Events  []ProjectEvent  `json:"events"`
}

type ProjectSummary struct {
	ProjectID         string   `json:"projectId"`
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	SchemaVersion     string   `json:"schemaVersion"`
	RootName          string   `json:"rootName"`
	OpenMode          string   `json:"openMode"`
	LockState         string   `json:"lockState"`
	LastCleanShutdown bool     `json:"lastCleanShutdown"`
	GraphVersion      int      `json:"graphVersion"`
	UpdatedAt         string   `json:"updatedAt"`
	Capabilities      []string `json:"capabilities"`
}

type HealthReport struct {
	Status    string       `json:"status"`
	CheckedAt string       `json:"checkedAt"`
	Items     []HealthItem `json:"items"`
}

func (report HealthReport) HasBlocking() bool {
	for _, item := range report.Items {
		if item.Severity == SeverityBlocking {
			return true
		}
	}
	return false
}

type HealthItem struct {
	Severity        string   `json:"severity"`
	Code            string   `json:"code"`
	Path            string   `json:"path,omitempty"`
	AffectedObjects []string `json:"affectedObjects"`
	UserMessage     string   `json:"userMessage"`
	TechnicalDetail string   `json:"technicalDetail,omitempty"`
	RecoveryActions []string `json:"recoveryActions"`
}

type OperationError struct {
	Code            string   `json:"code"`
	Severity        string   `json:"severity"`
	Retryable       bool     `json:"retryable"`
	TargetType      string   `json:"targetType,omitempty"`
	TargetID        string   `json:"targetId,omitempty"`
	UserMessage     string   `json:"userMessage"`
	TechnicalDetail string   `json:"technicalDetail,omitempty"`
	RecoveryActions []string `json:"recoveryActions"`
	CorrelationID   string   `json:"correlationId"`
}

type ProjectEvent struct {
	EventID     string          `json:"eventId"`
	EventType   string          `json:"eventType"`
	State       string          `json:"state"`
	Summary     string          `json:"summary"`
	Error       *OperationError `json:"error,omitempty"`
	NextActions []string        `json:"nextActions"`
	CreatedAt   string          `json:"createdAt"`
}

type LockMetadata struct {
	AppInstanceID string `json:"appInstanceId"`
	OpenedAt      string `json:"openedAt"`
	HeartbeatAt   string `json:"heartbeatAt"`
	PID           int    `json:"pid"`
	Host          string `json:"host"`
}

type LockInfo struct {
	State    string        `json:"state"`
	OpenMode string        `json:"openMode"`
	Message  string        `json:"message"`
	Metadata *LockMetadata `json:"metadata,omitempty"`
}

type DigestIndex struct {
	Files []DigestEntry `json:"files"`
}

type DigestEntry struct {
	Path            string   `json:"path"`
	SHA256          string   `json:"sha256"`
	AffectedObjects []string `json:"affectedObjects"`
}

type auditEntry struct {
	EventID       string         `json:"eventId"`
	EventType     string         `json:"eventType"`
	ProjectID     string         `json:"projectId,omitempty"`
	CorrelationID string         `json:"correlationId,omitempty"`
	CreatedAt     string         `json:"createdAt"`
	Summary       string         `json:"summary"`
	Details       map[string]any `json:"details,omitempty"`
}

type packageManifest struct {
	SchemaVersion           string   `json:"schemaVersion,omitempty"`
	ProjectID               string   `json:"projectId"`
	SceneID                 string   `json:"sceneId"`
	ShotID                  string   `json:"shotId"`
	PackageID               string   `json:"packageId"`
	PackageVersion          int      `json:"packageVersion,omitempty"`
	ProviderProfileID       string   `json:"providerProfileId,omitempty"`
	GenerationPackageStatus string   `json:"generationPackageStatus,omitempty"`
	ContextDigest           string   `json:"contextDigest,omitempty"`
	PromptPath              string   `json:"promptPath"`
	ScriptExcerptPath       string   `json:"scriptExcerptPath,omitempty"`
	ContinuityPath          string   `json:"continuityPath"`
	UploadChecklistPath     string   `json:"uploadChecklistPath"`
	References              []string `json:"references"`
	CreatedAt               string   `json:"createdAt,omitempty"`
	ManifestPath            string   `json:"-"`
}

type atomicWriteError struct {
	stage    string
	replaced bool
	err      error
}

func (err atomicWriteError) Error() string {
	return err.stage + ": " + err.err.Error()
}

func (err atomicWriteError) Unwrap() error {
	return err.err
}

func NewStore(options StoreOptions) *Store {
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	hostname := strings.TrimSpace(options.Hostname)
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	if hostname == "" {
		hostname = "local"
	}

	instanceID := strings.TrimSpace(options.InstanceID)
	if instanceID == "" {
		instanceID = "instance-" + now().UTC().Format("20060102T150405Z")
	}

	pid := options.PID
	if pid == 0 {
		pid = os.Getpid()
	}

	staleAfter := options.StaleAfter
	if staleAfter <= 0 {
		staleAfter = 2 * time.Minute
	}

	atomicWrite := options.AtomicWrite
	if atomicWrite == nil {
		atomicWrite = atomicWriteFile
	}

	return &Store{
		now:         now,
		instanceID:  instanceID,
		pid:         pid,
		hostname:    hostname,
		staleAfter:  staleAfter,
		atomicWrite: atomicWrite,
	}
}

func (s *Store) CreateProject(command CreateProjectCommand) OperationResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.failure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before creating the project.")
	}

	now := s.now().UTC()
	projectID := strings.TrimSpace(command.ProjectID)
	if projectID == "" {
		projectID = "proj_" + now.Format("20060102T150405Z")
	}

	name := strings.TrimSpace(command.Name)
	if name == "" {
		name = "Untitled Tuyu Project"
	}

	projectType := strings.TrimSpace(command.Type)
	if projectType == "" {
		projectType = DefaultProjectType
	}

	manifest, err := NewBaselineManifest(ManifestInput{
		ProjectID: projectID,
		Name:      name,
		Type:      projectType,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return s.failure(CodeProjectCreateFailed, SeverityBlocking, false, correlationID, "Project manifest could not be created.", err.Error())
	}

	if err := CreateBaselineProject(root, manifest); err != nil {
		return s.failure(CodeProjectCreateFailed, SeverityBlocking, true, correlationID, "Project folder could not be created.", err.Error())
	}

	lockInfo, err := s.acquireLock(root, true, correlationID, manifest.Project.ID)
	if err != nil {
		return s.failure(CodeProjectLockWriteFailed, SeverityBlocking, true, correlationID, "Project was created, but the session lock could not be written.", err.Error())
	}

	health := s.HealthReport(root)
	summary := summaryFromManifest(root, manifest, lockInfo)
	events := []ProjectEvent{s.event("project.created", "completed", "Project created on disk.", correlationID, nil)}

	return OperationResult{OK: true, Summary: &summary, Health: &health, Events: events}
}

func (s *Store) OpenProject(command OpenProjectCommand) OperationResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.failure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before opening the project.")
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		health := s.HealthReport(root)
		operationError := s.errorFromReadFailure(report, err, correlationID)
		return OperationResult{
			OK:     false,
			Health: &health,
			Error:  &operationError,
			Events: []ProjectEvent{s.event("project.open", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	lockInfo, err := s.acquireLock(root, command.Takeover, correlationID, manifest.Project.ID)
	if err != nil {
		operationError := s.operationError(CodeProjectLockWriteFailed, SeverityBlocking, true, correlationID, "Project lock could not be written.", err.Error(), []string{"Retry open or inspect the locks directory."})
		return OperationResult{
			OK:     false,
			Error:  &operationError,
			Health: ptr(s.HealthReport(root)),
			Events: []ProjectEvent{s.event("project.open", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	health := s.HealthReport(root)
	summary := summaryFromManifest(root, manifest, lockInfo)
	state := "completed"
	if summary.OpenMode != OpenModeReadWrite || health.HasBlocking() {
		state = "blocked"
	}

	return OperationResult{
		OK:      true,
		Summary: &summary,
		Health:  &health,
		Events:  []ProjectEvent{s.event("project.opened", state, "Project manifest loaded.", correlationID, nil)},
	}
}

func (s *Store) SaveProject(command SaveProjectCommand) OperationResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.failure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before saving the project.")
	}

	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		return s.failure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "Project is open in another active session.", "Open the project read-only or confirm takeover before saving.")
	}
	if lockInfo.State == LockStateStale {
		return s.failure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "Project has a stale lock that needs takeover before saving.", "Open the project with takeover, then retry save.")
	}

	current, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		return OperationResult{
			OK:     false,
			Health: ptr(s.HealthReport(root)),
			Error:  &operationError,
			Events: []ProjectEvent{s.event("project.save", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	next := current
	next.Project.UpdatedAt = s.now().UTC().Format(time.RFC3339)
	if next.Graph.Version <= 0 {
		next.Graph.Version = 1
	}
	next.Graph.Version++
	next.Integrity.LastCleanShutdown = true
	next.Integrity.LastGraphVersion = next.Graph.Version

	data, err := EncodeManifest(next)
	if err != nil {
		operationError := s.operationError(CodeProjectSaveValidation, SeverityBlocking, false, correlationID, "Project was not saved because the manifest failed validation.", err.Error(), []string{"Fix the manifest errors and retry save."})
		return OperationResult{
			OK:     false,
			Health: ptr(s.HealthReport(root)),
			Error:  &operationError,
			Events: []ProjectEvent{s.event("project.save", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	manifestPath := filepath.Join(root, ManifestFileName)
	if err := s.atomicWrite(manifestPath, data, 0o644, s.instanceID); err != nil {
		userMessage := "Project save failed before replacing the previous manifest."
		recoveryActions := []string{"Retry save. The previous manifest remains the active version."}
		var atomicErr atomicWriteError
		if errors.As(err, &atomicErr) && atomicErr.replaced {
			userMessage = "Project save replaced the manifest, but durability could not be confirmed."
			recoveryActions = []string{"Run health check before continuing production work, then retry save if the project reports recovery items."}
		}
		operationError := s.operationError(CodeProjectSaveCommit, SeverityBlocking, true, correlationID, userMessage, err.Error(), recoveryActions)
		return OperationResult{
			OK:     false,
			Health: ptr(s.HealthReport(root)),
			Error:  &operationError,
			Events: []ProjectEvent{s.event("project.save", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("project.saved", correlationID),
		EventType:     "project.saved",
		ProjectID:     next.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     s.now().UTC().Format(time.RFC3339),
		Summary:       "Project manifest saved.",
		Details: map[string]any{
			"graphVersion": next.Graph.Version,
			"schema":       next.SchemaVersion,
		},
	})

	health := s.HealthReport(root)
	summary := summaryFromManifest(root, next, lockInfo)

	return OperationResult{
		OK:      true,
		Summary: &summary,
		Health:  &health,
		Events:  []ProjectEvent{s.event("project.saved", "completed", "Project manifest saved with atomic replace.", correlationID, nil)},
	}
}

func (s *Store) CheckHealth(command CheckProjectHealthCommand) OperationResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.failure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before checking health.")
	}

	health := s.HealthReport(root)
	state := "completed"
	if health.HasBlocking() {
		state = "blocked"
	}

	return OperationResult{
		OK:     true,
		Health: &health,
		Events: []ProjectEvent{s.event("project.health_checked", state, "Project health check completed.", correlationID, nil)},
	}
}

func (s *Store) HealthReport(root string) HealthReport {
	checkedAt := s.now().UTC().Format(time.RFC3339)
	var items []HealthItem

	manifest, report, err := s.readManifest(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			items = append(items, HealthItem{
				Severity:        SeverityBlocking,
				Code:            CodeProjectManifestMissing,
				Path:            ManifestFileName,
				AffectedObjects: []string{},
				UserMessage:     "Project manifest is missing.",
				RecoveryActions: []string{"Restore project.tuyu.json from backup or recreate the project shell."},
			})
		} else if len(report.Items) == 0 {
			items = append(items, HealthItem{
				Severity:        SeverityBlocking,
				Code:            CodeHealthCheckIncomplete,
				Path:            ManifestFileName,
				AffectedObjects: []string{},
				UserMessage:     "Project manifest could not be read.",
				TechnicalDetail: err.Error(),
				RecoveryActions: []string{"Verify project permissions and ensure project.tuyu.json is a readable file."},
			})
		}
		items = append(items, healthItemsFromValidation(report)...)
		return HealthReport{Status: healthStatus(items), CheckedAt: checkedAt, Items: items}
	}

	items = append(items, healthItemsFromValidation(report)...)
	items = append(items, s.checkRequiredDirectories(root)...)
	items = append(items, healthItemsFromLock(s.inspectLock(root))...)
	items = append(items, s.checkGraphReferences(root, manifest)...)
	items = append(items, s.checkPackageReferences(root, manifest)...)
	items = append(items, s.checkDigestIndex(root)...)
	items = append(items, s.checkAssetIndex(root, manifest)...)

	if !manifest.Integrity.LastCleanShutdown {
		items = append(items, HealthItem{
			Severity:        SeverityWarning,
			Code:            CodeProjectDirtyShutdown,
			Path:            "integrity.lastCleanShutdown",
			AffectedObjects: []string{manifest.Project.ID},
			UserMessage:     "Project was not closed cleanly last time.",
			RecoveryActions: []string{"Run health check and inspect recovery candidates before continuing production work."},
		})
	}

	return HealthReport{Status: healthStatus(items), CheckedAt: checkedAt, Items: items}
}

func (s *Store) readManifest(root string) (Manifest, ValidationReport, error) {
	data, err := os.ReadFile(filepath.Join(root, ManifestFileName))
	if err != nil {
		return Manifest{}, ValidationReport{}, err
	}

	manifest, report := DecodeManifest(data)
	if report.HasBlocking() {
		return manifest, report, ValidationError{Report: report}
	}

	return manifest, report, nil
}

func (s *Store) acquireLock(root string, takeover bool, correlationID string, projectID string) (LockInfo, error) {
	current := s.inspectLock(root)
	if current.State == LockStateActive && !takeover {
		return current, nil
	}
	if current.State == LockStateStale && !takeover {
		return current, nil
	}

	state := LockStateOwned
	if current.State == LockStateActive || current.State == LockStateStale {
		state = LockStateTakeover
	}

	now := s.now().UTC().Format(time.RFC3339)
	metadata := LockMetadata{
		AppInstanceID: s.instanceID,
		OpenedAt:      now,
		HeartbeatAt:   now,
		PID:           s.pid,
		Host:          s.hostname,
	}

	if state == LockStateTakeover {
		if err := s.appendAudit(root, auditEntry{
			EventID:       s.eventID("project.lock_takeover", correlationID),
			EventType:     "project.lock_takeover",
			ProjectID:     projectID,
			CorrelationID: correlationID,
			CreatedAt:     now,
			Summary:       "Project lock takeover approved before replacing the session lock.",
			Details: map[string]any{
				"previousState": current.State,
				"host":          current.safeHost(),
			},
		}); err != nil {
			return LockInfo{}, fmt.Errorf("write takeover audit before lock replace: %w", err)
		}
		if err := s.appendRecoveryAudit(root, auditEntry{
			EventID:       s.eventID("project.lock_takeover.recovery", correlationID),
			EventType:     "project.lock_takeover",
			ProjectID:     projectID,
			CorrelationID: correlationID,
			CreatedAt:     now,
			Summary:       "Project lock takeover recorded before replacing the session lock.",
			Details: map[string]any{
				"previousState": current.State,
				"action":        "replace_session_lock",
			},
		}); err != nil {
			return LockInfo{}, fmt.Errorf("write recovery audit before lock replace: %w", err)
		}
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return LockInfo{}, fmt.Errorf("marshal session lock: %w", err)
	}
	data = append(data, '\n')

	lockPath := filepath.Join(root, filepath.FromSlash(LockFileRelativePath))
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return LockInfo{}, fmt.Errorf("create locks directory: %w", err)
	}
	if err := s.atomicWrite(lockPath, data, 0o644, s.instanceID); err != nil {
		return LockInfo{}, err
	}

	return LockInfo{
		State:    state,
		OpenMode: OpenModeReadWrite,
		Message:  "Project is open for editing in this session.",
		Metadata: &metadata,
	}, nil
}

func (s *Store) inspectLock(root string) LockInfo {
	lockPath := filepath.Join(root, filepath.FromSlash(LockFileRelativePath))
	data, err := os.ReadFile(lockPath)
	if errors.Is(err, os.ErrNotExist) {
		return LockInfo{State: LockStateUnlocked, OpenMode: OpenModeReadWrite, Message: "No active project lock."}
	}
	if err != nil {
		return LockInfo{State: LockStateStale, OpenMode: OpenModeTakeoverAvailable, Message: "Project lock could not be read."}
	}

	var metadata LockMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return LockInfo{State: LockStateStale, OpenMode: OpenModeTakeoverAvailable, Message: "Project lock is invalid."}
	}

	if metadata.AppInstanceID == s.instanceID {
		return LockInfo{State: LockStateOwned, OpenMode: OpenModeReadWrite, Message: "Project lock belongs to this session.", Metadata: &metadata}
	}

	if s.lockIsStale(metadata) {
		return LockInfo{State: LockStateStale, OpenMode: OpenModeTakeoverAvailable, Message: "Project lock is stale and can be taken over.", Metadata: &metadata}
	}

	return LockInfo{State: LockStateActive, OpenMode: OpenModeReadOnly, Message: "Project is locked by another active session.", Metadata: &metadata}
}

func (s *Store) lockIsStale(metadata LockMetadata) bool {
	heartbeat, err := time.Parse(time.RFC3339, metadata.HeartbeatAt)
	if err != nil {
		return true
	}
	if metadata.Host == s.hostname {
		return !processExists(metadata.PID)
	}
	return s.now().UTC().Sub(heartbeat.UTC()) > s.staleAfter
}

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

func (s *Store) checkRequiredDirectories(root string) []HealthItem {
	var items []HealthItem
	for _, dir := range RequiredProjectDirectories() {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir)))
		if err == nil && info.IsDir() {
			continue
		}
		items = append(items, HealthItem{
			Severity:        SeverityBlocking,
			Code:            CodeProjectDirectoryMissing,
			Path:            dir,
			AffectedObjects: []string{},
			UserMessage:     "Required project directory is missing.",
			RecoveryActions: []string{"Recreate the missing project directory before saving or exporting."},
		})
	}
	return items
}

func (s *Store) checkGraphReferences(root string, manifest Manifest) []HealthItem {
	var items []HealthItem
	nodeIDs := make(map[string]struct{}, len(manifest.Graph.Nodes))

	for index, node := range manifest.Graph.Nodes {
		nodePath := fmt.Sprintf("graph.nodes[%d]", index)
		affected := []string{}
		if strings.TrimSpace(node.ID) != "" {
			affected = []string{node.ID}
		}

		if strings.TrimSpace(node.ID) == "" {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeGraphReferenceMissing,
				Path:            nodePath + ".id",
				AffectedObjects: []string{},
				UserMessage:     "Graph node is missing an id.",
				RecoveryActions: []string{"Restore the graph node id from a valid project backup or remove the incomplete node after review."},
			})
			continue
		}
		if _, exists := nodeIDs[node.ID]; exists {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeGraphReferenceMissing,
				Path:            nodePath + ".id",
				AffectedObjects: affected,
				UserMessage:     "Graph node id is duplicated.",
				RecoveryActions: []string{"Restore unique graph node ids from backup or rebuild the affected graph partition."},
			})
		}
		nodeIDs[node.ID] = struct{}{}

		if strings.TrimSpace(node.RefID) == "" {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeGraphReferenceMissing,
				Path:            nodePath + ".refId",
				AffectedObjects: affected,
				UserMessage:     "Graph node has no referenced project object.",
				RecoveryActions: []string{"Reconnect the node to a project object or mark it as a broken placeholder before export."},
			})
			continue
		}

		if invalid := healthItemsForProjectPath("graph.refId", node.RefID, affected); len(invalid) > 0 {
			items = append(items, invalid...)
			continue
		}

		target := filepath.Join(root, filepath.FromSlash(node.RefID))
		info, err := os.Stat(target)
		if errors.Is(err, os.ErrNotExist) {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeGraphReferenceMissing,
				Path:            node.RefID,
				AffectedObjects: affected,
				UserMessage:     "Graph node references a missing project object.",
				RecoveryActions: []string{"Restore the missing object file, relink the node, or create a broken reference placeholder."},
			})
			continue
		}
		if err != nil {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeHealthCheckIncomplete,
				Path:            node.RefID,
				AffectedObjects: affected,
				UserMessage:     "Graph node reference could not be read.",
				TechnicalDetail: err.Error(),
				RecoveryActions: []string{"Retry health check after verifying project permissions."},
			})
			continue
		}
		if info.IsDir() {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeGraphReferenceMissing,
				Path:            node.RefID,
				AffectedObjects: affected,
				UserMessage:     "Graph node reference points to a directory, not a file.",
				RecoveryActions: []string{"Relink the node to the intended project object file."},
			})
		}
	}

	for index, edge := range manifest.Graph.Edges {
		edgeID := strings.TrimSpace(edge.ID)
		edgePath := fmt.Sprintf("graph.edges[%d]", index)
		affected := []string{}
		if edgeID != "" {
			affected = []string{edgeID}
			edgePath = "graph.edges." + edgeID
		}

		if edgeID == "" {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeGraphReferenceMissing,
				Path:            edgePath + ".id",
				AffectedObjects: []string{},
				UserMessage:     "Graph edge is missing an id.",
				RecoveryActions: []string{"Restore the graph edge id from backup or remove the incomplete edge after review."},
			})
		}
		if _, exists := nodeIDs[edge.Source]; !exists {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeGraphReferenceMissing,
				Path:            edgePath + ".source",
				AffectedObjects: affected,
				UserMessage:     "Graph edge source node is missing.",
				RecoveryActions: []string{"Restore the source node or remove the invalid edge after review."},
			})
		}
		if _, exists := nodeIDs[edge.Target]; !exists {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeGraphReferenceMissing,
				Path:            edgePath + ".target",
				AffectedObjects: affected,
				UserMessage:     "Graph edge target node is missing.",
				RecoveryActions: []string{"Restore the target node or remove the invalid edge after review."},
			})
		}
	}

	return items
}

func (s *Store) checkPackageReferences(root string, manifest Manifest) []HealthItem {
	manifestPaths := make(map[string]struct{})
	for _, node := range manifest.Graph.Nodes {
		if node.Kind != "package" || strings.TrimSpace(node.RefID) == "" {
			continue
		}
		if len(validateProjectPath("package.manifest", node.RefID)) > 0 {
			continue
		}
		manifestPaths[path.Clean(strings.ReplaceAll(node.RefID, "\\", "/"))] = struct{}{}
	}

	packagesRoot := filepath.Join(root, filepath.FromSlash(manifest.Paths.Packages))
	if err := filepath.WalkDir(packagesRoot, func(filename string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "manifest.json" {
			return nil
		}
		relative, relErr := filepath.Rel(root, filename)
		if relErr == nil {
			manifestPaths[filepath.ToSlash(relative)] = struct{}{}
		}
		return nil
	}); err != nil && !errors.Is(err, os.ErrNotExist) {
		return []HealthItem{{
			Severity:        SeverityWarning,
			Code:            CodeHealthCheckIncomplete,
			Path:            manifest.Paths.Packages,
			AffectedObjects: []string{},
			UserMessage:     "Package directory could not be scanned.",
			TechnicalDetail: err.Error(),
			RecoveryActions: []string{"Retry health check after verifying project permissions."},
		}}
	}

	ordered := make([]string, 0, len(manifestPaths))
	for relative := range manifestPaths {
		ordered = append(ordered, relative)
	}
	sort.Strings(ordered)

	var items []HealthItem
	for _, relative := range ordered {
		items = append(items, s.checkPackageManifest(root, relative)...)
	}
	return items
}

func (s *Store) checkPackageManifest(root string, manifestRelativePath string) []HealthItem {
	if invalid := healthItemsForProjectPath("package.manifest", manifestRelativePath, nil); len(invalid) > 0 {
		return invalid
	}

	manifestPath := filepath.Join(root, filepath.FromSlash(manifestRelativePath))
	data, err := os.ReadFile(manifestPath)
	if errors.Is(err, os.ErrNotExist) {
		return []HealthItem{{
			Severity:        SeverityWarning,
			Code:            CodePackageReferenceMissing,
			Path:            manifestRelativePath,
			AffectedObjects: []string{},
			UserMessage:     "Package manifest referenced by the graph is missing.",
			RecoveryActions: []string{"Re-export the package or restore the package manifest from backup."},
		}}
	}
	if err != nil {
		return []HealthItem{{
			Severity:        SeverityWarning,
			Code:            CodeHealthCheckIncomplete,
			Path:            manifestRelativePath,
			AffectedObjects: []string{},
			UserMessage:     "Package manifest could not be read.",
			TechnicalDetail: err.Error(),
			RecoveryActions: []string{"Retry health check after verifying project permissions."},
		}}
	}

	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return []HealthItem{{
			Severity:        SeverityBlocking,
			Code:            CodePackageManifestInvalid,
			Path:            manifestRelativePath,
			AffectedObjects: []string{},
			UserMessage:     "Package manifest is invalid JSON.",
			TechnicalDetail: err.Error(),
			RecoveryActions: []string{"Re-export the package or restore the package manifest from backup."},
		}}
	}

	var manifest packageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return []HealthItem{{
			Severity:        SeverityBlocking,
			Code:            CodePackageManifestInvalid,
			Path:            manifestRelativePath,
			AffectedObjects: []string{},
			UserMessage:     "Package manifest shape is invalid.",
			TechnicalDetail: err.Error(),
			RecoveryActions: []string{"Re-export the package or restore the package manifest from backup."},
		}}
	}

	affected := cleanAffectedObjects([]string{manifest.PackageID, manifest.ShotID, manifest.SceneID})
	items := scanPackageValues(manifestRelativePath, raw, affected)
	packageRoot := filepath.Dir(manifestPath)
	relativeRoot := filepath.ToSlash(filepath.Dir(manifestRelativePath))
	references := []struct {
		field string
		value string
	}{
		{field: "promptPath", value: manifest.PromptPath},
		{field: "continuityPath", value: manifest.ContinuityPath},
		{field: "uploadChecklistPath", value: manifest.UploadChecklistPath},
	}
	if strings.TrimSpace(manifest.ScriptExcerptPath) != "" {
		references = append(references, struct {
			field string
			value string
		}{field: "scriptExcerptPath", value: manifest.ScriptExcerptPath})
	}
	for index, reference := range manifest.References {
		references = append(references, struct {
			field string
			value string
		}{field: fmt.Sprintf("references[%d]", index), value: reference})
	}

	for _, reference := range references {
		fieldPath := manifestRelativePath + ":" + reference.field
		value := strings.TrimSpace(reference.value)
		if value == "" {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodePackageReferenceMissing,
				Path:            fieldPath,
				AffectedObjects: affected,
				UserMessage:     "Package manifest has an empty local reference path.",
				RecoveryActions: []string{"Re-export the package or restore the missing package-local reference."},
			})
			continue
		}

		if invalid := healthItemsForProjectPath(fieldPath, value, affected); len(invalid) > 0 {
			items = append(items, invalid...)
			continue
		}

		target := filepath.Join(packageRoot, filepath.FromSlash(value))
		if info, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodePackageReferenceMissing,
				Path:            path.Clean(relativeRoot + "/" + strings.ReplaceAll(value, "\\", "/")),
				AffectedObjects: affected,
				UserMessage:     "Package manifest references a missing file.",
				RecoveryActions: []string{"Re-export the package or restore the referenced package file before handoff."},
			})
		} else if err != nil {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeHealthCheckIncomplete,
				Path:            fieldPath,
				AffectedObjects: affected,
				UserMessage:     "Package manifest reference could not be read.",
				TechnicalDetail: err.Error(),
				RecoveryActions: []string{"Retry health check after verifying project permissions."},
			})
		} else if info.IsDir() {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodePackageReferenceMissing,
				Path:            path.Clean(relativeRoot + "/" + strings.ReplaceAll(value, "\\", "/")),
				AffectedObjects: affected,
				UserMessage:     "Package manifest references a directory instead of a file.",
				RecoveryActions: []string{"Relink the package reference to the intended file or re-export the package."},
			})
		}
	}

	return dedupeHealthItems(items)
}

func (s *Store) checkDigestIndex(root string) []HealthItem {
	indexPath := filepath.Join(root, filepath.FromSlash(DigestIndexRelativePath))
	data, err := os.ReadFile(indexPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []HealthItem{{
			Severity:        SeverityWarning,
			Code:            CodeHealthCheckIncomplete,
			Path:            DigestIndexRelativePath,
			AffectedObjects: []string{},
			UserMessage:     "Digest index could not be read.",
			RecoveryActions: []string{"Retry health check after verifying project permissions."},
		}}
	}

	var index DigestIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return []HealthItem{{
			Severity:        SeverityBlocking,
			Code:            CodeDigestIndexInvalid,
			Path:            DigestIndexRelativePath,
			AffectedObjects: []string{},
			UserMessage:     "Digest index is invalid JSON.",
			RecoveryActions: []string{"Restore the digest index from backup or remove it from the alpha project."},
		}}
	}

	var items []HealthItem
	for _, entry := range index.Files {
		invalidPath := false
		for _, issue := range validateProjectPath("digest.path", entry.Path) {
			items = append(items, healthItemFromValidation(issue))
			invalidPath = true
		}
		if invalidPath {
			continue
		}

		fileData, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(entry.Path)))
		if errors.Is(err, os.ErrNotExist) {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeAssetMissing,
				Path:            entry.Path,
				AffectedObjects: cleanAffectedObjects(entry.AffectedObjects),
				UserMessage:     "Referenced project file is missing.",
				RecoveryActions: []string{"Restore or relink the missing file before export."},
			})
			continue
		}
		if err != nil {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeHealthCheckIncomplete,
				Path:            entry.Path,
				AffectedObjects: cleanAffectedObjects(entry.AffectedObjects),
				UserMessage:     "Referenced project file could not be read.",
				RecoveryActions: []string{"Retry health check after verifying project permissions."},
			})
			continue
		}

		sum := sha256.Sum256(fileData)
		got := hex.EncodeToString(sum[:])
		if !strings.EqualFold(got, strings.TrimSpace(entry.SHA256)) {
			items = append(items, HealthItem{
				Severity:        SeverityBlocking,
				Code:            CodeDigestMismatch,
				Path:            entry.Path,
				AffectedObjects: cleanAffectedObjects(entry.AffectedObjects),
				UserMessage:     "Project file digest does not match its recorded value.",
				RecoveryActions: []string{"Reimport or verify the file before it is used in a package."},
			})
		}
	}

	return items
}

func healthItemsFromValidation(report ValidationReport) []HealthItem {
	items := make([]HealthItem, 0, len(report.Items))
	for _, issue := range report.Items {
		items = append(items, healthItemFromValidation(issue))
	}
	return items
}

func healthItemFromValidation(issue ValidationIssue) HealthItem {
	return HealthItem{
		Severity:        issue.Severity,
		Code:            issue.Code,
		Path:            issue.Path,
		AffectedObjects: []string{},
		UserMessage:     issue.UserMessage,
		TechnicalDetail: issue.TechnicalDetail,
		RecoveryActions: issue.RecoveryActions,
	}
}

func healthItemsForProjectPath(field string, value string, affected []string) []HealthItem {
	issues := validateProjectPath(field, value)
	items := make([]HealthItem, 0, len(issues))
	for _, issue := range issues {
		item := healthItemFromValidation(issue)
		item.Path = field
		item.AffectedObjects = cleanAffectedObjects(affected)
		items = append(items, item)
	}
	return items
}

func scanPackageValues(manifestRelativePath string, value any, affected []string) []HealthItem {
	var items []HealthItem
	walkStringValues("", value, func(path string, text string) {
		if !looksSensitive(text) {
			return
		}
		items = append(items, HealthItem{
			Severity:        SeverityBlocking,
			Code:            CodeSensitiveValue,
			Path:            manifestRelativePath + ":" + path,
			AffectedObjects: cleanAffectedObjects(affected),
			UserMessage:     "Package manifest contains a value that looks like private machine data or a credential.",
			TechnicalDetail: redactSensitiveDetail(text),
			RecoveryActions: []string{"Replace the value with a package-local or project-relative non-sensitive placeholder."},
		})
	})
	return items
}

func healthItemsFromLock(info LockInfo) []HealthItem {
	switch info.State {
	case LockStateActive:
		return []HealthItem{{
			Severity:        SeverityWarning,
			Code:            CodeProjectActiveLock,
			Path:            LockFileRelativePath,
			AffectedObjects: []string{},
			UserMessage:     "Project is locked by another active session.",
			RecoveryActions: []string{"Open read-only or confirm takeover before saving."},
		}}
	case LockStateStale:
		return []HealthItem{{
			Severity:        SeverityWarning,
			Code:            CodeProjectStaleLock,
			Path:            LockFileRelativePath,
			AffectedObjects: []string{},
			UserMessage:     "Project lock is stale.",
			RecoveryActions: []string{"Confirm takeover to replace the stale session lock."},
		}}
	default:
		return nil
	}
}

func healthStatus(items []HealthItem) string {
	status := HealthStatusClean
	for _, item := range items {
		if item.Severity == SeverityBlocking {
			return HealthStatusBlocking
		}
		if item.Severity == SeverityWarning {
			status = HealthStatusWarning
		}
	}
	return status
}

func dedupeHealthItems(items []HealthItem) []HealthItem {
	seen := make(map[string]bool, len(items))
	deduped := make([]HealthItem, 0, len(items))
	for _, item := range items {
		key := item.Code + "\x00" + item.Path + "\x00" + strings.Join(item.AffectedObjects, ",")
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, item)
	}
	return deduped
}

func summaryFromManifest(root string, manifest Manifest, lockInfo LockInfo) ProjectSummary {
	return ProjectSummary{
		ProjectID:         manifest.Project.ID,
		Name:              manifest.Project.Name,
		Type:              manifest.Project.Type,
		SchemaVersion:     manifest.SchemaVersion,
		RootName:          filepath.Base(filepath.Clean(root)),
		OpenMode:          lockInfo.OpenMode,
		LockState:         lockInfo.State,
		LastCleanShutdown: manifest.Integrity.LastCleanShutdown,
		GraphVersion:      manifest.Graph.Version,
		UpdatedAt:         manifest.Project.UpdatedAt,
		Capabilities: []string{
			"project_create",
			"project_open",
			"project_save",
			"project_lock",
			"project_health_check",
			"project_graph_view",
			"project_graph_layout_save",
			"project_asset_import",
			"project_asset_list",
		},
	}
}

func (s *Store) errorFromReadFailure(report ValidationReport, err error, correlationID string) OperationError {
	if errors.Is(err, os.ErrNotExist) {
		return s.operationError(CodeProjectManifestMissing, SeverityBlocking, true, correlationID, "Project manifest is missing.", ManifestFileName, []string{"Restore project.tuyu.json from backup or recreate the project shell."})
	}

	if hasValidationCode(report, CodeSchemaUnsupported) {
		return s.operationError(CodeSchemaUnsupported, SeverityBlocking, false, correlationID, "Project schema is not supported by this Alpha shell.", report.Codes()[0], []string{"Open the project with a compatible app version or export a diagnostic copy."})
	}

	return s.operationError(CodeProjectManifestInvalid, SeverityBlocking, false, correlationID, "Project manifest is invalid.", err.Error(), []string{"Fix project.tuyu.json or restore it from backup."})
}

func hasValidationCode(report ValidationReport, code string) bool {
	for _, got := range report.Codes() {
		if got == code {
			return true
		}
	}
	return false
}

func (s *Store) failure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string) OperationResult {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, []string{technicalDetail})
	if code == CodeProjectRootRequired {
		err.RecoveryActions = []string{technicalDetail}
	}
	return OperationResult{
		OK:     false,
		Error:  &err,
		Events: []ProjectEvent{s.event("project.operation", "blocked", userMessage, correlationID, &err)},
	}
}

func (s *Store) operationError(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string, recoveryActions []string) OperationError {
	return OperationError{
		Code:            code,
		Severity:        severity,
		Retryable:       retryable,
		TargetType:      "project",
		UserMessage:     userMessage,
		TechnicalDetail: technicalDetail,
		RecoveryActions: recoveryActions,
		CorrelationID:   correlationID,
	}
}

func (s *Store) event(eventType string, state string, summary string, correlationID string, err *OperationError) ProjectEvent {
	nextActions := []string{}
	if err != nil {
		nextActions = err.RecoveryActions
	}
	return ProjectEvent{
		EventID:     s.eventID(eventType, correlationID),
		EventType:   eventType,
		State:       state,
		Summary:     summary,
		Error:       err,
		NextActions: nextActions,
		CreatedAt:   s.now().UTC().Format(time.RFC3339),
	}
}

func (s *Store) eventID(eventType string, correlationID string) string {
	safeType := regexp.MustCompile(`[^a-zA-Z0-9_-]+`).ReplaceAllString(eventType, "_")
	return "evt-" + safeType + "-" + s.correlationID(correlationID)
}

func (s *Store) correlationID(value string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return "project-" + s.now().UTC().Format("20060102T150405Z")
}

func (s *Store) appendAudit(root string, entry auditEntry) error {
	auditPath := filepath.Join(root, filepath.FromSlash(AuditEventsRelativePath))
	return appendJSONL(auditPath, entry)
}

func (s *Store) appendRecoveryAudit(root string, entry auditEntry) error {
	auditPath := filepath.Join(root, filepath.FromSlash(RecoveryAuditRelativePath))
	return appendJSONL(auditPath, entry)
}

func appendJSONL(auditPath string, entry auditEntry) error {
	if err := os.MkdirAll(filepath.Dir(auditPath), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	file, err := os.OpenFile(auditPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Sync()
}

func atomicWriteFile(filename string, data []byte, perm fs.FileMode, instanceID string) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}

	tmpName := filepath.Join(filepath.Dir(filename), "."+filepath.Base(filename)+".tmp."+safeFileToken(instanceID))
	file, err := os.OpenFile(tmpName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return atomicWriteError{stage: "create_tmp", err: err}
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return atomicWriteError{stage: "write_tmp", err: err}
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return atomicWriteError{stage: "fsync_tmp", err: err}
	}
	if err := file.Close(); err != nil {
		return atomicWriteError{stage: "close_tmp", err: err}
	}
	if err := os.Rename(tmpName, filename); err != nil {
		return atomicWriteError{stage: "rename", err: err}
	}
	cleanup = false

	if err := fsyncDirectory(filepath.Dir(filename)); err != nil {
		return atomicWriteError{stage: "fsync_dir", replaced: true, err: err}
	}
	return nil
}

func fsyncDirectory(dir string) error {
	file, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer file.Close()
	return file.Sync()
}

func safeFileToken(value string) string {
	value = regexp.MustCompile(`[^a-zA-Z0-9_-]+`).ReplaceAllString(value, "_")
	if value == "" {
		return "local"
	}
	return value
}

func cleanAffectedObjects(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}

func (info LockInfo) safeHost() string {
	if info.Metadata == nil {
		return ""
	}
	return info.Metadata.Host
}

func ptr[T any](value T) *T {
	return &value
}
