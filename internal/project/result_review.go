package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const ResultIndexRelativePath = "assets/results/index.json"

const (
	CodeResultSourceMissing      = "result_source_missing"
	CodeResultTargetMissing      = "result_target_missing"
	CodeResultTargetInvalid      = "result_target_invalid"
	CodeResultTargetMismatch     = "result_target_mismatch"
	CodeResultFileMissing        = "result_file_missing"
	CodeResultDuplicateDigest    = "result_duplicate_digest"
	CodeResultPathInvalid        = "result_path_invalid"
	CodeResultWriteFailed        = "result_write_failed"
	CodeResultIndexInvalid       = "result_index_invalid"
	CodeResultReviewNoteMissing  = "result_review_note_missing"
	CodeResultReviewStateInvalid = "result_review_state_invalid"
	CodeResultMissing            = "result_missing"
	CodeResultRebindReasonNeeded = "result_rebind_reason_missing"
)

const (
	ResultDuplicatePolicyCancel  = "cancel"
	ResultDuplicatePolicyReuse   = "reuse"
	ResultDuplicatePolicyNewTake = "new_take"
)

const (
	ResultSourceLocalFile      = "local_file"
	ResultSourceMockRun        = "mock_run"
	ResultSourceHandoffPackage = "handoff_package"
	ResultSourceResultImport   = "result_import"
)

const (
	ResultReviewPending       = "pending"
	ResultReviewApproved      = "approved"
	ResultReviewNeedsRevision = "needs_revision"
	ResultReviewRejected      = "rejected"
)

const (
	ResultStatusImported       = "imported"
	ResultStatusBindingPending = "binding_pending"
	ResultStatusReviewPending  = "review_pending"
	ResultStatusApproved       = ResultReviewApproved
	ResultStatusNeedsRevision  = ResultReviewNeedsRevision
	ResultStatusRejected       = ResultReviewRejected
	ResultStatusMissingFile    = "missing_file"
)

type ImportResultCommand struct {
	Root            string `json:"root"`
	SourcePath      string `json:"sourcePath,omitempty"`
	ShotID          string `json:"shotId,omitempty"`
	PackageID       string `json:"packageId,omitempty"`
	RunID           string `json:"runId,omitempty"`
	DuplicatePolicy string `json:"duplicatePolicy,omitempty"`
	CreatedBy       string `json:"createdBy,omitempty"`
	CorrelationID   string `json:"correlationId"`
}

type ListResultsCommand struct {
	Root          string `json:"root"`
	ResultID      string `json:"resultId,omitempty"`
	ShotID        string `json:"shotId,omitempty"`
	PackageID     string `json:"packageId,omitempty"`
	CorrelationID string `json:"correlationId"`
}

type TraceResultCommand struct {
	Root          string `json:"root"`
	ResultID      string `json:"resultId"`
	CorrelationID string `json:"correlationId"`
}

type UpdateResultReviewCommand struct {
	Root          string `json:"root"`
	ResultID      string `json:"resultId"`
	ReviewStatus  string `json:"reviewStatus"`
	ReviewNotes   string `json:"reviewNotes,omitempty"`
	Reason        string `json:"reason,omitempty"`
	CorrelationID string `json:"correlationId"`
}

type RebindResultCommand struct {
	Root          string `json:"root"`
	ResultID      string `json:"resultId"`
	ShotID        string `json:"shotId,omitempty"`
	PackageID     string `json:"packageId,omitempty"`
	UnbindShot    bool   `json:"unbindShot,omitempty"`
	UnbindPackage bool   `json:"unbindPackage,omitempty"`
	Reason        string `json:"reason"`
	CorrelationID string `json:"correlationId"`
}

type ResultReviewResult struct {
	OK        bool                `json:"ok"`
	Result    *VideoResultDTO     `json:"result,omitempty"`
	Results   []VideoResultDTO    `json:"results"`
	Trace     *ResultTraceDTO     `json:"trace,omitempty"`
	Duplicate *ResultDuplicateDTO `json:"duplicate,omitempty"`
	Health    *HealthReport       `json:"health,omitempty"`
	Error     *OperationError     `json:"error,omitempty"`
	Events    []ProjectEvent      `json:"events"`
}

type VideoResultDTO struct {
	ID           string              `json:"id"`
	ProjectID    string              `json:"projectId"`
	ShotID       string              `json:"shotId,omitempty"`
	PackageID    string              `json:"packageId,omitempty"`
	TakeNumber   int                 `json:"takeNumber"`
	AssetID      string              `json:"assetId"`
	RelativePath string              `json:"relativePath"`
	RecordPath   string              `json:"recordPath"`
	Source       ResultSourceDTO     `json:"source"`
	Status       string              `json:"status"`
	ReviewStatus string              `json:"reviewStatus"`
	ReviewNotes  string              `json:"reviewNotes,omitempty"`
	ReviewReason string              `json:"reviewReason,omitempty"`
	BindingState string              `json:"bindingState"`
	Missing      bool                `json:"missing"`
	TakeHistory  []ResultTakeHistory `json:"takeHistory"`
	CreatedAt    string              `json:"createdAt"`
	UpdatedAt    string              `json:"updatedAt"`
}

type ResultSourceDTO struct {
	Kind         string `json:"kind"`
	SourcePath   string `json:"sourcePath,omitempty"`
	ImportedName string `json:"importedName,omitempty"`
	Digest       string `json:"digest,omitempty"`
	MimeType     string `json:"mimeType,omitempty"`
	RunID        string `json:"runId,omitempty"`
	PackageID    string `json:"packageId,omitempty"`
}

type ResultTakeHistory struct {
	ShotID     string `json:"shotId,omitempty"`
	PackageID  string `json:"packageId,omitempty"`
	TakeNumber int    `json:"takeNumber"`
	Reason     string `json:"reason,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

type ResultDuplicateDTO struct {
	ExistingResultID string `json:"existingResultId"`
	ExistingAssetID  string `json:"existingAssetId"`
	Digest           string `json:"digest"`
	MimeType         string `json:"mimeType"`
	Policy           string `json:"policy"`
}

type ResultTraceDTO struct {
	Result          *VideoResultDTO       `json:"result,omitempty"`
	Asset           *AssetDTO             `json:"asset,omitempty"`
	Shot            *ShotCardDTO          `json:"shot,omitempty"`
	Package         *GenerationPackageDTO `json:"package,omitempty"`
	Run             *MockRunDTO           `json:"run,omitempty"`
	RecoveryActions []string              `json:"recoveryActions"`
}

type resultIndex struct {
	ProjectID string           `json:"projectId"`
	Results   []VideoResultDTO `json:"results"`
	UpdatedAt string           `json:"updatedAt"`
}

type resultTarget struct {
	shotID      string
	packageID   string
	packageData *packageManifest
}

func (s *Store) ImportResult(command ImportResultCommand) ResultReviewResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.resultFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "", "Project root is required.", "Choose a project folder before importing a result.", nil)
	}
	if failure := s.resultWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "video_result"
		return ResultReviewResult{
			OK:      false,
			Results: []VideoResultDTO{},
			Health:  ptr(s.HealthReport(root)),
			Error:   &operationError,
			Events:  []ProjectEvent{s.event("result.import", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	sourcePath := strings.TrimSpace(command.SourcePath)
	source := ResultSourceDTO{Kind: ResultSourceLocalFile}
	target := resultTarget{
		shotID:    strings.TrimSpace(command.ShotID),
		packageID: strings.TrimSpace(command.PackageID),
	}
	if runID := strings.TrimSpace(command.RunID); runID != "" {
		run, ok, err := s.readMockRunByID(root, manifest, runID)
		if err != nil {
			return s.resultFailure(CodeMockRunReadFailed, SeverityBlocking, true, correlationID, runID, "Mock run could not be read for result import.", err.Error(), ptr(s.HealthReport(root)))
		}
		if !ok {
			return s.resultFailure(CodeMockRunTargetMissing, SeverityBlocking, false, correlationID, runID, "Mock run was not found for result import.", runID, ptr(s.HealthReport(root)))
		}
		source.Kind = ResultSourceMockRun
		source.RunID = run.RunID
		if sourcePath == "" && run.Output != nil {
			sourcePath = run.Output.RelativePath
		}
		if target.shotID == "" {
			target.shotID = run.ShotID
		}
		if target.packageID == "" {
			target.packageID = run.PackageID
		}
	}
	if sourcePath == "" {
		return s.resultFailure(CodeResultSourceMissing, SeverityBlocking, false, correlationID, "", "Result source file is required.", "Choose a placeholder result file or provide a mock run with output.", ptr(s.HealthReport(root)))
	}

	resolvedTarget, targetErr := s.resolveResultTarget(root, manifest, target, correlationID)
	if targetErr != nil {
		return *targetErr
	}
	target = resolvedTarget
	if source.Kind == ResultSourceLocalFile && target.packageID != "" {
		source.Kind = ResultSourceHandoffPackage
	}
	source.PackageID = target.packageID

	sourceAbs, err := resolveAssetSourcePath(root, sourcePath)
	if err != nil {
		code := CodeAssetFileUnreadable
		if errors.Is(err, errAssetSourcePathRequired) {
			code = CodeResultSourceMissing
		} else if errors.Is(err, errAssetPathInvalid) {
			code = CodeResultPathInvalid
		}
		return s.resultFailure(code, SeverityBlocking, true, correlationID, "", "Result source file could not be read.", err.Error(), ptr(s.HealthReport(root)))
	}
	facts, err := readAssetSourceFacts(sourceAbs)
	if err != nil {
		return s.resultFailure(assetReadErrorCode(err), SeverityBlocking, true, correlationID, "", "Result source file could not be imported.", err.Error(), ptr(s.HealthReport(root)))
	}
	source.SourcePath = cleanResultSourcePath(root, sourcePath)
	source.ImportedName = facts.filename
	source.Digest = facts.digest
	source.MimeType = facts.mimeType

	results, err := s.loadResultIndex(root, manifest.Project.ID)
	if err != nil {
		return s.resultFailure(CodeResultIndexInvalid, SeverityBlocking, true, correlationID, "", "Result index could not be read.", err.Error(), ptr(s.HealthReport(root)))
	}
	assetIndex, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.resultFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "", "Asset index could not be read before result import.", err.Error(), ptr(s.HealthReport(root)))
	}

	policy := normalizeResultDuplicatePolicy(command.DuplicatePolicy)
	if duplicate := findDuplicateResult(results.Results, facts.digest, facts.mimeType); duplicate != nil && policy != ResultDuplicatePolicyNewTake {
		duplicateDTO := ResultDuplicateDTO{
			ExistingResultID: duplicate.ID,
			ExistingAssetID:  duplicate.AssetID,
			Digest:           facts.digest,
			MimeType:         facts.mimeType,
			Policy:           policy,
		}
		if policy == ResultDuplicatePolicyReuse {
			result := s.hydrateResult(root, *duplicate)
			all := s.hydrateResults(root, results.Results)
			return ResultReviewResult{
				OK:        true,
				Result:    &result,
				Results:   all,
				Duplicate: &duplicateDTO,
				Health:    ptr(s.HealthReport(root)),
				Events:    []ProjectEvent{s.event("result.import.reused", "completed", "Existing result reused for duplicate digest.", correlationID, nil)},
			}
		}
		err := s.resultOperationError(CodeResultDuplicateDigest, SeverityWarning, false, correlationID, "Result with the same digest already exists.", duplicate.ID, []string{"Choose reuse to keep the existing result, or new_take to create another take with the same file digest."})
		return ResultReviewResult{
			OK:        false,
			Results:   s.hydrateResults(root, results.Results),
			Duplicate: &duplicateDTO,
			Health:    ptr(s.HealthReport(root)),
			Error:     &err,
			Events:    []ProjectEvent{s.event("result.import.duplicate", "blocked", err.UserMessage, correlationID, &err)},
		}
	}

	now := s.now().UTC()
	targetKey := resultStorageKey(target.shotID, target.packageID)
	takeNumber := 0
	if target.shotID != "" {
		takeNumber = nextResultTakeNumber(results.Results, target.shotID, "")
	}
	resultID := uniqueResultID(results.Results, targetKey, takeNumber, facts.digest)
	assetID := uniqueAssetID(assetIndex.Assets, facts.kind, facts.digest, AssetDuplicatePolicyCopy)
	relativePath := resultFileRelativePath(manifest.Paths, targetKey, facts, assetID)
	recordPath := resultRecordRelativePath(manifest.Paths, targetKey, resultID)
	for field, relative := range map[string]string{
		"result.relativePath": relativePath,
		"result.recordPath":   recordPath,
	} {
		if issues := validateProjectPath(field, relative); len(issues) > 0 {
			return s.resultFailure(CodeResultPathInvalid, SeverityBlocking, false, correlationID, resultID, "Result destination path was rejected.", issues[0].TechnicalDetail, ptr(s.HealthReport(root)))
		}
	}

	destination := filepath.Join(root, filepath.FromSlash(relativePath))
	copiedPaths := []string{}
	cleanupOnFailure := true
	defer func() {
		if !cleanupOnFailure {
			return
		}
		for _, filename := range copiedPaths {
			_ = os.Remove(filename)
		}
	}()
	if err := copyFileAtomic(sourceAbs, destination, s.instanceID); err != nil {
		return s.resultFailure(CodeAssetFileUnreadable, SeverityBlocking, true, correlationID, resultID, "Result source file could not be copied into the project.", err.Error(), ptr(s.HealthReport(root)))
	}
	copiedPaths = append(copiedPaths, destination)
	destinationDigest, destinationSize, err := digestAssetFile(destination)
	if err != nil {
		return s.resultFailure(CodeAssetDigestFailed, SeverityBlocking, true, correlationID, resultID, "Imported result file could not be verified.", err.Error(), ptr(s.HealthReport(root)))
	}
	if destinationDigest != facts.digest || destinationSize != facts.size {
		return s.resultFailure(CodeAssetDigestFailed, SeverityBlocking, true, correlationID, resultID, "Result source changed during import.", "The copied result did not match the source digest and size captured before copy.", ptr(s.HealthReport(root)))
	}

	status := ResultStatusReviewPending
	bindingState := ResultStatusReviewPending
	if target.shotID == "" && target.packageID == "" {
		status = ResultStatusBindingPending
		bindingState = ResultStatusBindingPending
	}
	history := []ResultTakeHistory{}
	if takeNumber > 0 {
		history = append(history, ResultTakeHistory{
			ShotID:     target.shotID,
			PackageID:  target.packageID,
			TakeNumber: takeNumber,
			Reason:     "import",
			CreatedAt:  now.Format(time.RFC3339),
		})
	}
	result := VideoResultDTO{
		ID:           resultID,
		ProjectID:    manifest.Project.ID,
		ShotID:       target.shotID,
		PackageID:    target.packageID,
		TakeNumber:   takeNumber,
		AssetID:      assetID,
		RelativePath: relativePath,
		RecordPath:   recordPath,
		Source:       source,
		Status:       status,
		ReviewStatus: ResultReviewPending,
		BindingState: bindingState,
		TakeHistory:  history,
		CreatedAt:    now.Format(time.RFC3339),
		UpdatedAt:    now.Format(time.RFC3339),
	}
	asset := AssetDTO{
		ID:              assetID,
		ProjectID:       manifest.Project.ID,
		Type:            facts.kind,
		Role:            "video_result",
		RelativePath:    relativePath,
		OriginalName:    facts.filename,
		MimeType:        facts.mimeType,
		SizeBytes:       facts.size,
		Digest:          facts.digest,
		Source:          AssetSourceDTO{Kind: ResultSourceResultImport, OriginalName: facts.filename, ImportedAt: now.Format(time.RFC3339)},
		Bindings:        []AssetBindingDTO{},
		ThumbnailStatus: AssetThumbnailNone,
		CreatedAt:       now.Format(time.RFC3339),
		UpdatedAt:       now.Format(time.RFC3339),
	}

	if err := s.writeResultRecord(root, result); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, resultID, "Result record could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}
	copiedPaths = append(copiedPaths, filepath.Join(root, filepath.FromSlash(recordPath)))
	results.Results = append(results.Results, result)
	results.UpdatedAt = now.Format(time.RFC3339)
	sortResults(results.Results)
	if err := s.writeResultIndex(root, results); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, resultID, "Result index could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}
	assetIndex.Assets = append(assetIndex.Assets, asset)
	assetIndex.UpdatedAt = now.Format(time.RFC3339)
	sortAssets(assetIndex.Assets)
	if err := s.writeAssetIndex(root, assetIndex); err != nil {
		return s.resultFailure(CodeAssetIndexWriteFailed, SeverityBlocking, true, correlationID, resultID, "Result asset index could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}

	if target.shotID != "" {
		if err := s.setShotResultLink(root, manifest, target.shotID, result.ID, true, now); err != nil {
			return s.resultFailure(CodeShotSaveFailed, SeverityBlocking, true, correlationID, resultID, "Shot result linkage could not be saved.", err.Error(), ptr(s.HealthReport(root)))
		}
	}
	if target.packageID != "" {
		if err := s.markPackageResultReceived(root, target.packageID, now); err != nil {
			return s.resultFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, resultID, "Package result_received state could not be saved.", err.Error(), ptr(s.HealthReport(root)))
		}
	}
	if err := s.updateManifestForResult(root, result, now); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, resultID, "Project manifest could not record the result node.", err.Error(), ptr(s.HealthReport(root)))
	}
	if err := s.appendResultAudit(root, manifest.Project.ID, correlationID, now, "result.import", "VideoResult imported for review.", result, map[string]any{
		"assetId":      assetID,
		"shotId":       target.shotID,
		"packageId":    target.packageID,
		"takeNumber":   takeNumber,
		"relativePath": relativePath,
		"recordPath":   recordPath,
		"sourceKind":   source.Kind,
		"sourceRunId":  source.RunID,
		"sourceDigest": source.Digest,
		"bindingState": bindingState,
		"reviewStatus": result.ReviewStatus,
	}); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, resultID, "Result audit event could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}
	cleanupOnFailure = false

	result = s.hydrateResult(root, result)
	return ResultReviewResult{
		OK:      true,
		Result:  &result,
		Results: s.hydrateResults(root, results.Results),
		Health:  ptr(s.HealthReport(root)),
		Events:  []ProjectEvent{s.event("result.imported", eventStateForResult(result), "VideoResult imported for review.", correlationID, nil)},
	}
}

func (s *Store) ListResults(command ListResultsCommand) ResultReviewResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.resultFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "", "Project root is required.", "Choose a project folder before listing results.", nil)
	}
	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "video_result"
		return ResultReviewResult{
			OK:      false,
			Results: []VideoResultDTO{},
			Health:  ptr(s.HealthReport(root)),
			Error:   &operationError,
			Events:  []ProjectEvent{s.event("result.list", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}
	index, err := s.loadResultIndex(root, manifest.Project.ID)
	if err != nil {
		return s.resultFailure(CodeResultIndexInvalid, SeverityBlocking, true, correlationID, "", "Result index could not be read.", err.Error(), ptr(s.HealthReport(root)))
	}
	results := filterResults(index.Results, command.ResultID, command.ShotID, command.PackageID)
	results = s.hydrateResults(root, results)
	var selected *VideoResultDTO
	if strings.TrimSpace(command.ResultID) != "" && len(results) == 1 {
		selected = &results[0]
	}
	return ResultReviewResult{
		OK:      true,
		Result:  selected,
		Results: results,
		Health:  ptr(s.HealthReport(root)),
		Events:  []ProjectEvent{s.event("result.listed", "completed", "VideoResults listed.", correlationID, nil)},
	}
}

func (s *Store) TraceResult(command TraceResultCommand) ResultReviewResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.resultFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "", "Project root is required.", "Choose a project folder before tracing a result.", nil)
	}
	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "video_result"
		return ResultReviewResult{
			OK:      false,
			Results: []VideoResultDTO{},
			Health:  ptr(s.HealthReport(root)),
			Error:   &operationError,
			Events:  []ProjectEvent{s.event("result.trace", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}
	result, ok, errResult := s.findResult(root, manifest, command.ResultID, correlationID)
	if errResult != nil {
		return *errResult
	}
	if !ok {
		return s.resultFailure(CodeResultMissing, SeverityBlocking, false, correlationID, command.ResultID, "Result was not found.", strings.TrimSpace(command.ResultID), ptr(s.HealthReport(root)))
	}
	trace := s.resultTrace(root, manifest, result, correlationID)
	return ResultReviewResult{
		OK:      true,
		Result:  trace.Result,
		Results: []VideoResultDTO{*trace.Result},
		Trace:   &trace,
		Health:  ptr(s.HealthReport(root)),
		Events:  []ProjectEvent{s.event("result.trace.loaded", eventStateForResult(*trace.Result), "VideoResult trace loaded.", correlationID, nil)},
	}
}

func (s *Store) UpdateResultReview(command UpdateResultReviewCommand) ResultReviewResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.resultFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "", "Project root is required.", "Choose a project folder before updating review.", nil)
	}
	if failure := s.resultWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}
	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "video_result"
		return ResultReviewResult{OK: false, Results: []VideoResultDTO{}, Health: ptr(s.HealthReport(root)), Error: &operationError, Events: []ProjectEvent{s.event("result.review", "blocked", operationError.UserMessage, correlationID, &operationError)}}
	}
	index, err := s.loadResultIndex(root, manifest.Project.ID)
	if err != nil {
		return s.resultFailure(CodeResultIndexInvalid, SeverityBlocking, true, correlationID, "", "Result index could not be read.", err.Error(), ptr(s.HealthReport(root)))
	}
	resultIndex := indexOfResult(index.Results, command.ResultID)
	if resultIndex < 0 {
		return s.resultFailure(CodeResultMissing, SeverityBlocking, false, correlationID, command.ResultID, "Result was not found.", strings.TrimSpace(command.ResultID), ptr(s.HealthReport(root)))
	}
	reviewStatus, ok := normalizeReviewStatus(command.ReviewStatus)
	if !ok {
		return s.resultFailure(CodeResultReviewStateInvalid, SeverityBlocking, false, correlationID, command.ResultID, "Review status is not supported.", strings.TrimSpace(command.ReviewStatus), ptr(s.HealthReport(root)))
	}
	reason := strings.TrimSpace(command.Reason)
	if reviewRequiresReason(reviewStatus) && reason == "" && strings.TrimSpace(command.ReviewNotes) == "" {
		return s.resultFailure(CodeResultReviewNoteMissing, SeverityBlocking, false, correlationID, command.ResultID, "Review reason is required for this state.", reviewStatus, ptr(s.HealthReport(root)))
	}

	now := s.now().UTC()
	result := index.Results[resultIndex]
	if bindingStateForResult(result) == ResultStatusBindingPending && reviewStatus != ResultReviewPending {
		return s.resultFailure(CodeResultTargetMissing, SeverityBlocking, false, correlationID, result.ID, "Result must be bound before review.", "Bind the result to a Shot or Package before setting a review decision.", ptr(s.HealthReport(root)))
	}
	result.ReviewStatus = reviewStatus
	result.ReviewNotes = strings.TrimSpace(command.ReviewNotes)
	if reason != "" {
		result.ReviewReason = reason
	} else if reviewRequiresReason(reviewStatus) {
		result.ReviewReason = strings.TrimSpace(command.ReviewNotes)
	} else if reviewStatus == ResultReviewApproved || reviewStatus == ResultReviewPending {
		result.ReviewReason = ""
	}
	result.Status = resultStatusForReview(result, reviewStatus)
	result.BindingState = bindingStateForResult(result)
	result.UpdatedAt = now.Format(time.RFC3339)
	index.Results[resultIndex] = result
	index.UpdatedAt = now.Format(time.RFC3339)
	if err := s.writeResultRecord(root, result); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, result.ID, "Result record could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}
	if err := s.writeResultIndex(root, index); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, result.ID, "Result index could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}
	if err := s.updateManifestForResult(root, result, now); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, result.ID, "Project manifest result node could not be updated.", err.Error(), ptr(s.HealthReport(root)))
	}
	_ = s.appendResultAudit(root, manifest.Project.ID, correlationID, now, "result.review", "VideoResult review state updated.", result, map[string]any{
		"reviewStatus": result.ReviewStatus,
		"status":       result.Status,
		"hasReason":    strings.TrimSpace(result.ReviewReason) != "",
	})
	result = s.hydrateResult(root, result)
	return ResultReviewResult{
		OK:      true,
		Result:  &result,
		Results: s.hydrateResults(root, index.Results),
		Health:  ptr(s.HealthReport(root)),
		Events:  []ProjectEvent{s.event("result.review.updated", eventStateForResult(result), "VideoResult review state updated.", correlationID, nil)},
	}
}

func (s *Store) RebindResult(command RebindResultCommand) ResultReviewResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.resultFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "", "Project root is required.", "Choose a project folder before rebinding a result.", nil)
	}
	if strings.TrimSpace(command.Reason) == "" {
		return s.resultFailure(CodeResultRebindReasonNeeded, SeverityBlocking, false, correlationID, command.ResultID, "Rebind reason is required.", "Record why the Shot or Package binding changed.", ptr(s.HealthReport(root)))
	}
	if failure := s.resultWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}
	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "video_result"
		return ResultReviewResult{OK: false, Results: []VideoResultDTO{}, Health: ptr(s.HealthReport(root)), Error: &operationError, Events: []ProjectEvent{s.event("result.rebind", "blocked", operationError.UserMessage, correlationID, &operationError)}}
	}
	index, err := s.loadResultIndex(root, manifest.Project.ID)
	if err != nil {
		return s.resultFailure(CodeResultIndexInvalid, SeverityBlocking, true, correlationID, "", "Result index could not be read.", err.Error(), ptr(s.HealthReport(root)))
	}
	resultIndex := indexOfResult(index.Results, command.ResultID)
	if resultIndex < 0 {
		return s.resultFailure(CodeResultMissing, SeverityBlocking, false, correlationID, command.ResultID, "Result was not found.", strings.TrimSpace(command.ResultID), ptr(s.HealthReport(root)))
	}
	result := index.Results[resultIndex]
	oldShotID := result.ShotID
	oldPackageID := result.PackageID
	target := resultTarget{shotID: result.ShotID, packageID: result.PackageID}
	if command.UnbindShot {
		target.shotID = ""
	}
	if command.UnbindPackage {
		target.packageID = ""
	}
	if strings.TrimSpace(command.ShotID) != "" {
		target.shotID = strings.TrimSpace(command.ShotID)
	}
	if strings.TrimSpace(command.PackageID) != "" {
		target.packageID = strings.TrimSpace(command.PackageID)
	}
	resolvedTarget, targetErr := s.resolveResultTarget(root, manifest, target, correlationID)
	if targetErr != nil {
		return *targetErr
	}
	target = resolvedTarget

	now := s.now().UTC()
	bindingChanged := oldShotID != target.shotID || oldPackageID != target.packageID
	reviewReset := bindingChanged && (result.ReviewStatus != ResultReviewPending || strings.TrimSpace(result.ReviewNotes) != "" || strings.TrimSpace(result.ReviewReason) != "")
	if oldShotID != target.shotID {
		if oldShotID != "" {
			if err := s.setShotResultLink(root, manifest, oldShotID, result.ID, false, now); err != nil {
				return s.resultFailure(CodeShotSaveFailed, SeverityBlocking, true, correlationID, result.ID, "Old Shot result linkage could not be removed.", err.Error(), ptr(s.HealthReport(root)))
			}
		}
		if target.shotID != "" {
			if err := s.setShotResultLink(root, manifest, target.shotID, result.ID, true, now); err != nil {
				return s.resultFailure(CodeShotSaveFailed, SeverityBlocking, true, correlationID, result.ID, "New Shot result linkage could not be saved.", err.Error(), ptr(s.HealthReport(root)))
			}
			result.TakeNumber = nextResultTakeNumber(index.Results, target.shotID, result.ID)
		} else {
			result.TakeNumber = 0
		}
	}
	if bindingChanged {
		result.TakeHistory = append(result.TakeHistory, ResultTakeHistory{
			ShotID:     target.shotID,
			PackageID:  target.packageID,
			TakeNumber: result.TakeNumber,
			Reason:     strings.TrimSpace(command.Reason),
			CreatedAt:  now.Format(time.RFC3339),
		})
	}
	result.ShotID = target.shotID
	result.PackageID = target.packageID
	if bindingChanged {
		result.ReviewStatus = ResultReviewPending
		result.ReviewNotes = ""
		result.ReviewReason = ""
	}
	result.BindingState = bindingStateForResult(result)
	result.Status = resultStatusForReview(result, result.ReviewStatus)
	result.UpdatedAt = now.Format(time.RFC3339)
	index.Results[resultIndex] = result
	index.UpdatedAt = now.Format(time.RFC3339)
	if err := s.writeResultRecord(root, result); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, result.ID, "Result record could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}
	if err := s.writeResultIndex(root, index); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, result.ID, "Result index could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}
	if target.packageID != "" && target.packageID != oldPackageID {
		if err := s.markPackageResultReceived(root, target.packageID, now); err != nil {
			return s.resultFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, result.ID, "Package result_received state could not be saved.", err.Error(), ptr(s.HealthReport(root)))
		}
	}
	if err := s.updateManifestForResult(root, result, now); err != nil {
		return s.resultFailure(CodeResultWriteFailed, SeverityBlocking, true, correlationID, result.ID, "Project manifest result node could not be updated.", err.Error(), ptr(s.HealthReport(root)))
	}
	_ = s.appendResultAudit(root, manifest.Project.ID, correlationID, now, "result.rebind", "VideoResult binding updated without moving the imported file.", result, map[string]any{
		"oldShotId":    oldShotID,
		"oldPackageId": oldPackageID,
		"newShotId":    result.ShotID,
		"newPackageId": result.PackageID,
		"reason":       strings.TrimSpace(command.Reason),
		"relativePath": result.RelativePath,
		"reviewReset":  reviewReset,
	})
	result = s.hydrateResult(root, result)
	return ResultReviewResult{
		OK:      true,
		Result:  &result,
		Results: s.hydrateResults(root, index.Results),
		Health:  ptr(s.HealthReport(root)),
		Events:  []ProjectEvent{s.event("result.rebound", eventStateForResult(result), "VideoResult binding updated without moving the imported file.", correlationID, nil)},
	}
}

func (s *Store) resultWriteLockFailure(root string, correlationID string) *ResultReviewResult {
	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		result := s.resultFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "", "Project is open in another active session.", "Open the project read-only or confirm takeover before changing results.", ptr(s.HealthReport(root)))
		return &result
	}
	if lockInfo.State == LockStateStale {
		result := s.resultFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "", "Project has a stale lock that needs takeover before changing results.", "Open the project with takeover, then retry result import or review.", ptr(s.HealthReport(root)))
		return &result
	}
	return nil
}

func (s *Store) resolveResultTarget(root string, manifest Manifest, target resultTarget, correlationID string) (resultTarget, *ResultReviewResult) {
	target.shotID = strings.TrimSpace(target.shotID)
	target.packageID = strings.TrimSpace(target.packageID)
	if target.packageID != "" {
		pkg, ok := findPackageManifestByID(s.readPackageManifests(root, manifest), target.packageID)
		if !ok {
			result := s.resultFailure(CodeResultTargetInvalid, SeverityBlocking, false, correlationID, target.packageID, "Result package target was not found.", target.packageID, ptr(s.HealthReport(root)))
			return resultTarget{}, &result
		}
		target.packageData = &pkg
		packageShotID := strings.TrimSpace(pkg.ShotID)
		if target.shotID != "" && packageShotID != "" && target.shotID != packageShotID {
			result := s.resultFailure(CodeResultTargetMismatch, SeverityBlocking, false, correlationID, target.packageID, "Result Shot and Package targets do not match.", fmt.Sprintf("package %s belongs to shot %s, not %s", target.packageID, packageShotID, target.shotID), ptr(s.HealthReport(root)))
			return resultTarget{}, &result
		}
		if target.shotID == "" {
			target.shotID = packageShotID
		}
	}
	if target.shotID != "" {
		if _, errResult := s.readShotContextRecord(root, manifest, target.shotID, correlationID); errResult != nil {
			return resultTarget{}, resultFromShotContextFailure(*errResult, target.shotID)
		}
	}
	return target, nil
}

func resultFromShotContextFailure(result ShotContextResult, targetID string) *ResultReviewResult {
	err := result.Error
	if err != nil {
		err.TargetType = "video_result"
		err.TargetID = targetID
	}
	return &ResultReviewResult{
		OK:      false,
		Results: []VideoResultDTO{},
		Error:   err,
		Events:  result.Events,
	}
}

func (s *Store) loadResultIndex(root string, projectID string) (resultIndex, error) {
	index := resultIndex{ProjectID: projectID, Results: []VideoResultDTO{}}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ResultIndexRelativePath)))
	if errors.Is(err, os.ErrNotExist) {
		return index, nil
	}
	if err != nil {
		return resultIndex{}, err
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return resultIndex{}, err
	}
	if strings.TrimSpace(index.ProjectID) == "" {
		index.ProjectID = projectID
	}
	if index.Results == nil {
		index.Results = []VideoResultDTO{}
	}
	for i := range index.Results {
		index.Results[i] = normalizeVideoResult(index.Results[i], projectID)
	}
	sortResults(index.Results)
	return index, nil
}

func (s *Store) writeResultIndex(root string, index resultIndex) error {
	if index.Results == nil {
		index.Results = []VideoResultDTO{}
	}
	sortResults(index.Results)
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return s.atomicWrite(filepath.Join(root, filepath.FromSlash(ResultIndexRelativePath)), data, 0o644, s.instanceID)
}

func (s *Store) writeResultRecord(root string, result VideoResultDTO) error {
	if issues := validateProjectPath("result.recordPath", result.RecordPath); len(issues) > 0 {
		return errors.New(issues[0].TechnicalDetail)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	filename := filepath.Join(root, filepath.FromSlash(result.RecordPath))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return s.atomicWrite(filename, data, 0o644, s.instanceID)
}

func (s *Store) findResult(root string, manifest Manifest, resultID string, correlationID string) (VideoResultDTO, bool, *ResultReviewResult) {
	index, err := s.loadResultIndex(root, manifest.Project.ID)
	if err != nil {
		result := s.resultFailure(CodeResultIndexInvalid, SeverityBlocking, true, correlationID, "", "Result index could not be read.", err.Error(), ptr(s.HealthReport(root)))
		return VideoResultDTO{}, false, &result
	}
	for _, result := range index.Results {
		if strings.TrimSpace(result.ID) == strings.TrimSpace(resultID) {
			return s.hydrateResult(root, result), true, nil
		}
	}
	return VideoResultDTO{}, false, nil
}

func (s *Store) hydrateResults(root string, results []VideoResultDTO) []VideoResultDTO {
	hydrated := make([]VideoResultDTO, 0, len(results))
	for _, result := range results {
		hydrated = append(hydrated, s.hydrateResult(root, result))
	}
	sortResults(hydrated)
	return hydrated
}

func (s *Store) hydrateResult(root string, result VideoResultDTO) VideoResultDTO {
	result = normalizeVideoResult(result, result.ProjectID)
	result.Missing = assetFileMissing(root, result.RelativePath)
	if result.Missing {
		result.Status = ResultStatusMissingFile
	}
	result.BindingState = bindingStateForResult(result)
	return result
}

func normalizeVideoResult(result VideoResultDTO, projectID string) VideoResultDTO {
	result.ID = strings.TrimSpace(result.ID)
	if strings.TrimSpace(result.ProjectID) == "" {
		result.ProjectID = strings.TrimSpace(projectID)
	}
	result.ShotID = strings.TrimSpace(result.ShotID)
	result.PackageID = strings.TrimSpace(result.PackageID)
	result.AssetID = strings.TrimSpace(result.AssetID)
	result.RelativePath = cleanProjectRelativePath(result.RelativePath)
	result.RecordPath = cleanProjectRelativePath(result.RecordPath)
	if result.Source.Kind == "" {
		result.Source.Kind = ResultSourceLocalFile
	}
	if result.ReviewStatus == "" {
		result.ReviewStatus = ResultReviewPending
	}
	if result.Status == "" {
		result.Status = resultStatusForReview(result, result.ReviewStatus)
	}
	if result.BindingState == "" {
		result.BindingState = bindingStateForResult(result)
	}
	if result.TakeHistory == nil {
		result.TakeHistory = []ResultTakeHistory{}
	}
	return result
}

func (s *Store) resultTrace(root string, manifest Manifest, result VideoResultDTO, correlationID string) ResultTraceDTO {
	result = s.hydrateResult(root, result)
	trace := ResultTraceDTO{
		Result:          &result,
		RecoveryActions: resultRecoveryActionsForState(result),
	}
	if index, err := s.loadAssetIndex(root, manifest.Project.ID); err == nil {
		for _, asset := range s.hydrateAssetsForList(root, index.Assets) {
			if asset.ID == result.AssetID {
				trace.Asset = &asset
				break
			}
		}
	}
	if result.ShotID != "" {
		if record, errResult := s.readShotContextRecord(root, manifest, result.ShotID, correlationID); errResult == nil {
			trace.Shot = &record.shot
		}
	}
	if result.PackageID != "" {
		if pkg, ok := findPackageManifestByID(s.readPackageManifests(root, manifest), result.PackageID); ok {
			dto := packageDTOFromManifest(pkg, manifest.Project.ID)
			trace.Package = &dto
		}
	}
	if result.Source.RunID != "" {
		if run, ok, err := s.readMockRunByID(root, manifest, result.Source.RunID); err == nil && ok {
			trace.Run = &run
		}
	}
	return trace
}

func packageDTOFromManifest(pkg packageManifest, projectID string) GenerationPackageDTO {
	return GenerationPackageDTO{
		PackageID:               pkg.PackageID,
		ProjectID:               projectID,
		SceneID:                 pkg.SceneID,
		ShotID:                  pkg.ShotID,
		PackageVersion:          pkg.PackageVersion,
		ProviderProfileID:       pkg.ProviderProfileID,
		GenerationPackageStatus: pkg.GenerationPackageStatus,
		ContextDigest:           pkg.ContextDigest,
		RelativePath:            path.Dir(pkg.ManifestPath),
		ManifestPath:            pkg.ManifestPath,
		PromptPath:              path.Join(path.Dir(pkg.ManifestPath), pkg.PromptPath),
		ScriptExcerptPath:       path.Join(path.Dir(pkg.ManifestPath), pkg.ScriptExcerptPath),
		ContinuityPath:          path.Join(path.Dir(pkg.ManifestPath), pkg.ContinuityPath),
		UploadChecklistPath:     path.Join(path.Dir(pkg.ManifestPath), pkg.UploadChecklistPath),
		References:              []GenerationPackageReferenceDTO{},
		CreatedAt:               pkg.CreatedAt,
	}
}

func (s *Store) setShotResultLink(root string, manifest Manifest, shotID string, resultID string, add bool, now time.Time) error {
	record, errResult := s.readShotContextRecord(root, manifest, shotID, "result-link")
	if errResult != nil {
		if errResult.Error != nil {
			return errors.New(errResult.Error.TechnicalDetail)
		}
		return errors.New("shot result linkage failed")
	}
	shot := record.shot
	if add {
		shot.ResultIDs = addUniqueString(shot.ResultIDs, resultID)
	} else {
		shot.ResultIDs = removeString(shot.ResultIDs, resultID)
	}
	shot.UpdatedAt = now.UTC().Format(time.RFC3339)
	patchShotRaw(record.raw, shot)
	record.raw["resultIds"] = shot.ResultIDs
	return s.writeShotRaw(root, manifest, shot.ID, record.raw)
}

func (s *Store) markPackageResultReceived(root string, packageID string, now time.Time) error {
	manifest, _, err := s.readManifest(root)
	if err != nil {
		return err
	}
	for _, manifests := range s.readPackageManifests(root, manifest) {
		for _, pkg := range manifests {
			if pkg.PackageID != packageID {
				continue
			}
			pkg.GenerationPackageStatus = GenerationPackageStatusResultReceived
			if err := s.writePackageManifest(root, pkg); err != nil {
				return err
			}
			return s.updateManifestPackageStatus(root, pkg.ManifestPath, pkg.PackageID, GenerationPackageStatusResultReceived, now)
		}
	}
	return fmt.Errorf("%s: %s", CodeResultTargetInvalid, packageID)
}

func (s *Store) updateManifestForResult(root string, result VideoResultDTO, now time.Time) error {
	manifest, _, err := s.readManifest(root)
	if err != nil {
		return err
	}
	nodeID := resultGraphNodeID(result.ID)
	updated := false
	for index := range manifest.Graph.Nodes {
		node := &manifest.Graph.Nodes[index]
		if strings.TrimSpace(node.ID) == nodeID {
			node.Kind = "video_result"
			node.RefID = result.RecordPath
			node.Status = result.Status
			updated = true
			break
		}
	}
	if !updated {
		manifest.Graph.Nodes = append(manifest.Graph.Nodes, Node{
			ID:     nodeID,
			Kind:   "video_result",
			RefID:  result.RecordPath,
			Status: result.Status,
		})
	}
	edges := manifest.Graph.Edges[:0]
	for _, edge := range manifest.Graph.Edges {
		if strings.TrimSpace(edge.Source) == nodeID || strings.TrimSpace(edge.Target) == nodeID {
			continue
		}
		edges = append(edges, edge)
	}
	manifest.Graph.Edges = edges
	if result.ShotID != "" {
		shotNodeID := manifestShotNodeID(root, manifest, result.ShotID)
		if shotNodeID == "" {
			return fmt.Errorf("%s: %s", CodeShotGraphNodeMissing, result.ShotID)
		}
		manifest.Graph.Edges = append(manifest.Graph.Edges, Edge{
			ID:     "edge_" + safeFileToken(nodeID+"_"+shotNodeID+"_result_of"),
			Source: nodeID,
			Target: shotNodeID,
			Kind:   "result_of",
		})
	}
	if result.PackageID != "" {
		if packageNodeID := s.packageNodeIDByPackageID(root, manifest, result.PackageID); packageNodeID != "" {
			manifest.Graph.Edges = append(manifest.Graph.Edges, Edge{
				ID:     "edge_" + safeFileToken(nodeID+"_"+packageNodeID+"_result_of"),
				Source: nodeID,
				Target: packageNodeID,
				Kind:   "result_of",
			})
		}
	}
	manifest.Project.UpdatedAt = now.UTC().Format(time.RFC3339)
	manifest.Graph.Version++
	manifest.Integrity.LastGraphVersion = manifest.Graph.Version
	manifest.Integrity.LastCleanShutdown = true
	data, err := EncodeManifest(manifest)
	if err != nil {
		return err
	}
	return s.atomicWrite(filepath.Join(root, ManifestFileName), data, 0o644, s.instanceID)
}

func (s *Store) packageNodeIDByPackageID(root string, manifest Manifest, packageID string) string {
	packages := map[string]packageManifest{}
	for _, manifests := range s.readPackageManifests(root, manifest) {
		for _, pkg := range manifests {
			packages[pkg.PackageID] = pkg
		}
	}
	pkg, hasPackage := packages[packageID]
	for _, node := range manifest.Graph.Nodes {
		if strings.TrimSpace(node.Kind) != "package" {
			continue
		}
		if strings.TrimSpace(node.ID) == "node_package_"+safeFileToken(packageID) {
			return strings.TrimSpace(node.ID)
		}
		if hasPackage && path.Clean(strings.ReplaceAll(node.RefID, "\\", "/")) == path.Clean(pkg.ManifestPath) {
			return strings.TrimSpace(node.ID)
		}
	}
	return ""
}

func (s *Store) appendResultAudit(root string, projectID string, correlationID string, now time.Time, eventType string, summary string, result VideoResultDTO, details map[string]any) error {
	if details == nil {
		details = map[string]any{}
	}
	details["resultId"] = result.ID
	details["status"] = result.Status
	details["reviewStatus"] = result.ReviewStatus
	return s.appendAudit(root, auditEntry{
		EventID:       s.eventID(eventType, correlationID),
		EventType:     eventType,
		ProjectID:     projectID,
		CorrelationID: correlationID,
		CreatedAt:     now.UTC().Format(time.RFC3339),
		Summary:       summary,
		Details:       details,
	})
}

func resultStorageKey(shotID string, packageID string) string {
	if strings.TrimSpace(shotID) != "" {
		return safeFileToken(shotID)
	}
	if strings.TrimSpace(packageID) != "" {
		return "package_" + safeFileToken(packageID)
	}
	return "binding_pending"
}

func resultFileRelativePath(paths Paths, targetKey string, facts assetSourceFacts, assetID string) string {
	base := cleanOrDefaultPath(paths.AssetResults, DefaultPaths().AssetResults)
	readable := safeAssetToken(strings.TrimSuffix(facts.filename, filepath.Ext(facts.filename)))
	if readable == "" {
		readable = "result"
	}
	return path.Join(base, targetKey, readable+"-"+assetID+facts.ext)
}

func resultRecordRelativePath(paths Paths, targetKey string, resultID string) string {
	base := cleanOrDefaultPath(paths.AssetResults, DefaultPaths().AssetResults)
	return path.Join(base, targetKey, resultID+".json")
}

func uniqueResultID(results []VideoResultDTO, targetKey string, takeNumber int, digest string) string {
	take := "pending"
	if takeNumber > 0 {
		take = fmt.Sprintf("take_%03d", takeNumber)
	}
	base := "result_" + safeFileToken(targetKey) + "_" + take + "_" + digestSummary(digest)
	if !resultIDExists(results, base) {
		return base
	}
	for index := 2; ; index++ {
		candidate := fmt.Sprintf("%s_copy_%03d", base, index)
		if !resultIDExists(results, candidate) {
			return candidate
		}
	}
}

func resultIDExists(results []VideoResultDTO, id string) bool {
	for _, result := range results {
		if strings.TrimSpace(result.ID) == id {
			return true
		}
	}
	return false
}

func resultGraphNodeID(resultID string) string {
	return "node_result_" + safeFileToken(resultID)
}

func nextResultTakeNumber(results []VideoResultDTO, shotID string, excludeResultID string) int {
	shotID = strings.TrimSpace(shotID)
	maxTake := 0
	for _, result := range results {
		if strings.TrimSpace(result.ID) == strings.TrimSpace(excludeResultID) {
			continue
		}
		if strings.TrimSpace(result.ShotID) != shotID {
			continue
		}
		if result.TakeNumber > maxTake {
			maxTake = result.TakeNumber
		}
	}
	return maxTake + 1
}

func findDuplicateResult(results []VideoResultDTO, digest string, mimeType string) *VideoResultDTO {
	for index := range results {
		source := results[index].Source
		if strings.EqualFold(source.Digest, digest) && strings.EqualFold(source.MimeType, mimeType) {
			return &results[index]
		}
	}
	return nil
}

func normalizeResultDuplicatePolicy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ResultDuplicatePolicyReuse:
		return ResultDuplicatePolicyReuse
	case ResultDuplicatePolicyNewTake, "new", "copy", "create_new_take":
		return ResultDuplicatePolicyNewTake
	default:
		return ResultDuplicatePolicyCancel
	}
}

func normalizeReviewStatus(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", ResultReviewPending, "review_pending":
		return ResultReviewPending, true
	case ResultReviewApproved:
		return ResultReviewApproved, true
	case ResultReviewNeedsRevision, "needs-revision":
		return ResultReviewNeedsRevision, true
	case ResultReviewRejected:
		return ResultReviewRejected, true
	default:
		return "", false
	}
}

func reviewRequiresReason(status string) bool {
	return status == ResultReviewNeedsRevision || status == ResultReviewRejected
}

func bindingStateForResult(result VideoResultDTO) string {
	if strings.TrimSpace(result.ShotID) == "" && strings.TrimSpace(result.PackageID) == "" {
		return ResultStatusBindingPending
	}
	return ResultStatusReviewPending
}

func resultStatusForReview(result VideoResultDTO, reviewStatus string) string {
	if bindingStateForResult(result) == ResultStatusBindingPending {
		return ResultStatusBindingPending
	}
	switch reviewStatus {
	case ResultReviewApproved:
		return ResultStatusApproved
	case ResultReviewNeedsRevision:
		return ResultStatusNeedsRevision
	case ResultReviewRejected:
		return ResultStatusRejected
	default:
		return ResultStatusReviewPending
	}
}

func eventStateForResult(result VideoResultDTO) string {
	switch result.Status {
	case ResultStatusBindingPending, ResultStatusMissingFile:
		return "blocked"
	case ResultStatusRejected:
		return "failed"
	case ResultStatusNeedsRevision:
		return "blocked"
	default:
		return "completed"
	}
}

func resultRecoveryActionsForState(result VideoResultDTO) []string {
	switch result.Status {
	case ResultStatusBindingPending:
		return []string{"Bind the result to a Shot or Package before review."}
	case ResultStatusMissingFile:
		return []string{"Restore the imported result file, reimport it as a new take, or keep the record as missing evidence."}
	case ResultStatusNeedsRevision:
		return []string{"Create a revision task or run a new mock/provider attempt from the referenced Shot or Package."}
	case ResultStatusRejected:
		return []string{"Keep the rejected take for history and import a replacement result as a new take."}
	default:
		return []string{"Inspect Shot, Package, Asset, and Run trace before changing review state."}
	}
}

func filterResults(results []VideoResultDTO, resultID string, shotID string, packageID string) []VideoResultDTO {
	resultID = strings.TrimSpace(resultID)
	shotID = strings.TrimSpace(shotID)
	packageID = strings.TrimSpace(packageID)
	filtered := make([]VideoResultDTO, 0, len(results))
	for _, result := range results {
		if resultID != "" && result.ID != resultID {
			continue
		}
		if shotID != "" && result.ShotID != shotID {
			continue
		}
		if packageID != "" && result.PackageID != packageID {
			continue
		}
		filtered = append(filtered, result)
	}
	return filtered
}

func indexOfResult(results []VideoResultDTO, resultID string) int {
	resultID = strings.TrimSpace(resultID)
	for index := range results {
		if strings.TrimSpace(results[index].ID) == resultID {
			return index
		}
	}
	return -1
}

func removeString(values []string, target string) []string {
	target = strings.TrimSpace(target)
	next := values[:0]
	for _, value := range values {
		if strings.TrimSpace(value) == "" || strings.TrimSpace(value) == target {
			continue
		}
		next = append(next, strings.TrimSpace(value))
	}
	return cleanStringList(next)
}

func sortResults(results []VideoResultDTO) {
	sort.SliceStable(results, func(i int, j int) bool {
		if results[i].CreatedAt == results[j].CreatedAt {
			return results[i].ID < results[j].ID
		}
		return results[i].CreatedAt < results[j].CreatedAt
	})
}

func cleanResultSourcePath(root string, value string) string {
	value = strings.TrimSpace(value)
	if filepath.IsAbs(value) {
		if relative, err := filepath.Rel(root, value); err == nil && !strings.HasPrefix(relative, "..") {
			return cleanProjectRelativePath(relative)
		}
		return filepath.Base(value)
	}
	return cleanProjectRelativePath(value)
}

func (s *Store) resultFailure(code string, severity string, retryable bool, correlationID string, targetID string, userMessage string, technicalDetail string, health *HealthReport) ResultReviewResult {
	err := s.resultOperationError(code, severity, retryable, correlationID, userMessage, technicalDetail, resultRecoveryActions(code, technicalDetail))
	err.TargetID = strings.TrimSpace(targetID)
	return ResultReviewResult{
		OK:      false,
		Results: []VideoResultDTO{},
		Health:  health,
		Error:   &err,
		Events:  []ProjectEvent{s.event("result.operation", "blocked", userMessage, correlationID, &err)},
	}
}

func (s *Store) resultOperationError(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string, recoveryActions []string) OperationError {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, recoveryActions)
	err.TargetType = "video_result"
	return err
}

func resultRecoveryActions(code string, detail string) []string {
	switch code {
	case CodeResultSourceMissing:
		return []string{"Choose a readable placeholder result file, or provide a mock run with a completed output."}
	case CodeResultDuplicateDigest:
		return []string{"Choose reuse to keep the existing result, or new_take to create another take."}
	case CodeResultReviewNoteMissing:
		return []string{"Add a review reason before marking a result as needs_revision or rejected."}
	case CodeResultTargetMissing:
		return []string{"Bind the imported result to a Shot or Package before review."}
	case CodeResultTargetInvalid:
		return []string{"Choose an existing Shot or exported GenerationPackage."}
	case CodeResultFileMissing:
		return []string{"Restore the result file or reimport it as a new take without deleting the missing record."}
	case CodeResultPathInvalid:
		return []string{"Keep result records and copied files under project-relative assets/results paths."}
	case CodeResultWriteFailed:
		return []string{"Retry after checking project write permissions; imported files are not considered reviewed until the result index is saved."}
	case CodeResultRebindReasonNeeded:
		return []string{"Record why the previous Shot or Package binding was wrong before rebinding."}
	default:
		if strings.TrimSpace(detail) != "" {
			return []string{detail}
		}
		return []string{"Retry after fixing the result workflow input."}
	}
}
