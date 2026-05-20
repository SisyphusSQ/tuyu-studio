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

const (
	CodeContinuityRuleMissing      = "continuity_rule_missing"
	CodeContinuityRuleInvalid      = "continuity_rule_invalid"
	CodeContinuityUnlockReason     = "continuity_unlock_reason_required"
	CodeContinuityWriteFailed      = "continuity_write_failed"
	CodeContinuityLockedBinding    = "binding_locked"
	CodeContinuityStaleBinding     = "stale_binding"
	CodeContinuityImpactIncomplete = "continuity_impact_incomplete"
	CodeContinuityDirtyPropagate   = "continuity_dirty_propagate_failed"
	CodeContinuityBindingSelector  = "binding_unlock_selector_invalid"
)

const (
	ContinuitySeverityBlocking   = "blocking"
	ContinuitySeverityWarning    = "warning"
	ContinuitySeveritySuggestion = "suggestion"
)

type ListContinuityCommand struct {
	Root          string `json:"root"`
	CorrelationID string `json:"correlationId"`
}

type SaveContinuityRuleCommand struct {
	Root          string `json:"root"`
	ID            string `json:"id,omitempty"`
	TargetType    string `json:"targetType"`
	TargetID      string `json:"targetId"`
	Rule          string `json:"rule"`
	Severity      string `json:"severity,omitempty"`
	Locked        bool   `json:"locked"`
	CreatedBy     string `json:"createdBy,omitempty"`
	CorrelationID string `json:"correlationId"`
}

type UnlockContinuityRuleCommand struct {
	Root          string `json:"root"`
	RuleID        string `json:"ruleId"`
	Reason        string `json:"reason"`
	CorrelationID string `json:"correlationId"`
}

type UnlockAssetBindingCommand struct {
	Root          string `json:"root"`
	BindingID     string `json:"bindingId,omitempty"`
	AssetID       string `json:"assetId,omitempty"`
	TargetType    string `json:"targetType,omitempty"`
	TargetID      string `json:"targetId,omitempty"`
	Purpose       string `json:"purpose,omitempty"`
	Reason        string `json:"reason"`
	CorrelationID string `json:"correlationId"`
}

type ContinuityRuleDTO struct {
	ID         string `json:"id"`
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Rule       string `json:"rule"`
	Severity   string `json:"severity"`
	Locked     bool   `json:"locked"`
	CreatedBy  string `json:"createdBy"`
	CreatedAt  string `json:"createdAt,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
}

type ContinuityImpactReportDTO struct {
	AffectedAssets   []string                   `json:"affectedAssets"`
	AffectedBindings []string                   `json:"affectedBindings"`
	AffectedProfiles []string                   `json:"affectedProfiles"`
	AffectedShots    []string                   `json:"affectedShots"`
	AffectedPackages []string                   `json:"affectedPackages"`
	Issues           []ContinuityImpactIssueDTO `json:"issues"`
	RecoveryActions  []string                   `json:"recoveryActions"`
	CheckedAt        string                     `json:"checkedAt"`
}

type ContinuityImpactIssueDTO struct {
	Code            string   `json:"code"`
	Severity        string   `json:"severity"`
	TargetType      string   `json:"targetType,omitempty"`
	TargetID        string   `json:"targetId,omitempty"`
	UserMessage     string   `json:"userMessage"`
	RecoveryActions []string `json:"recoveryActions"`
}

type ContinuityResult struct {
	OK     bool                       `json:"ok"`
	Rule   *ContinuityRuleDTO         `json:"rule,omitempty"`
	Rules  []ContinuityRuleDTO        `json:"rules"`
	Impact *ContinuityImpactReportDTO `json:"impact,omitempty"`
	Health *HealthReport              `json:"health,omitempty"`
	Error  *OperationError            `json:"error,omitempty"`
	Events []ProjectEvent             `json:"events"`
}

type continuityImpactIndex struct {
	assetAffected     map[string][]string
	bindingAffected   map[string][]string
	targetAffected    map[string][]string
	assetShots        map[string][]string
	targetShots       map[string][]string
	ruleShots         map[string][]string
	shotPackages      map[string][]string
	targetNodeID      map[string]string
	shotNodeID        map[string]string
	assetsByID        map[string]AssetDTO
	profilesByKey     map[string]profileRecord
	shots             []shotContextRecord
	packagesByShot    map[string][]packageManifest
	incompleteReasons []string
}

func (s *Store) ListContinuity(command ListContinuityCommand) ContinuityResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.continuityFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before listing continuity rules.")
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "continuity"
		return ContinuityResult{
			OK:     false,
			Rules:  []ContinuityRuleDTO{},
			Error:  &operationError,
			Events: []ProjectEvent{s.event("continuity.list", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}
	profiles, err := s.loadProfileRecords(root, manifest)
	if err != nil {
		return s.continuityFailure(CodeBindingProfileInvalid, SeverityBlocking, true, correlationID, "Profile library could not be read.", err.Error())
	}
	index, _ := s.loadAssetIndex(root, manifest.Project.ID)
	impact := s.buildContinuityImpactIndex(root, manifest, index)
	reportDTO := s.continuityImpactReport(impact, "", "", "", "")

	return ContinuityResult{
		OK:     true,
		Rules:  continuityRulesFromProfiles(profiles),
		Impact: &reportDTO,
		Health: ptr(s.HealthReport(root)),
		Events: []ProjectEvent{s.event("continuity.listed", "completed", "Continuity rules and impact report listed.", correlationID, nil)},
	}
}

func (s *Store) SaveContinuityRule(command SaveContinuityRuleCommand) ContinuityResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.continuityFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before saving a continuity rule.")
	}
	if failure := s.continuityWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "continuity"
		return ContinuityResult{
			OK:     false,
			Rules:  []ContinuityRuleDTO{},
			Error:  &operationError,
			Events: []ProjectEvent{s.event("continuity.rule.save", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}
	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.continuityFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "Asset index could not be read.", err.Error())
	}
	targetType, ok := normalizeBindingTargetType(command.TargetType)
	if !ok {
		return s.continuityFailure(CodeContinuityRuleInvalid, SeverityBlocking, false, correlationID, "Continuity rule target type is not supported.", strings.TrimSpace(command.TargetType))
	}
	profile, err := s.loadProfileRecord(root, manifest, targetType, command.TargetID)
	if err != nil {
		return s.continuityFailure(CodeBindingTargetMissing, SeverityBlocking, true, correlationID, "Continuity rule target does not exist.", err.Error())
	}
	ruleText := strings.TrimSpace(command.Rule)
	if ruleText == "" {
		return s.continuityFailure(CodeContinuityRuleInvalid, SeverityBlocking, false, correlationID, "Continuity rule text is required.", "Enter the locked or warning continuity rule before saving.")
	}

	now := s.now().UTC()
	ruleID := strings.TrimSpace(command.ID)
	if ruleID == "" {
		ruleID = continuityRuleID(targetType, profile.dto.ID, ruleText)
	}
	rule := ContinuityRuleDTO{
		ID:         ruleID,
		TargetType: targetType,
		TargetID:   profile.dto.ID,
		Rule:       ruleText,
		Severity:   normalizeContinuitySeverity(command.Severity),
		Locked:     command.Locked,
		CreatedBy:  normalizeCreatedBy(command.CreatedBy),
		UpdatedAt:  now.Format(time.RFC3339),
	}
	previous := findContinuityRule(profile.dto.ContinuityRules, rule.ID)
	if previous != nil && previous.Locked && !rule.Locked {
		return s.continuityFailure(CodeContinuityUnlockReason, SeverityBlocking, false, correlationID, "Locked continuity rule cannot be unlocked through save.", "Use continuity rule unlock with a reason so the audit trail records the override.")
	}
	if previous != nil && strings.TrimSpace(previous.CreatedAt) != "" {
		rule.CreatedAt = previous.CreatedAt
	} else {
		rule.CreatedAt = now.Format(time.RFC3339)
	}
	profile.dto.ContinuityRules = upsertContinuityRule(profile.dto.ContinuityRules, rule)
	profile.dto.LockedRules = syncLockedRuleIDs(profile.dto.LockedRules, profile.dto.ContinuityRules)
	profile.raw = profileRawWithBaseline(profile.raw, profile.dto)

	if err := s.writeProfileAndIndex(root, profile, index); err != nil {
		return s.continuityFailure(CodeContinuityWriteFailed, SeverityBlocking, true, correlationID, "Continuity rule could not be written.", err.Error())
	}

	changedShots, dirtyErr := s.markShotsDirtyForContinuityChange(root, manifest, targetType, profile.dto.ID, rule.ID, "continuity rule updated: "+rule.ID, correlationID, now)
	if dirtyErr != nil {
		return s.continuityFailure(CodeContinuityDirtyPropagate, SeverityBlocking, true, correlationID, "Continuity rule was saved, but affected Shot dirty marking failed.", dirtyErr.Error())
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("continuity.rule.saved", correlationID),
		EventType:     "continuity.rule.saved",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "Continuity rule saved.",
		Details: map[string]any{
			"ruleId":        rule.ID,
			"targetType":    targetType,
			"targetId":      profile.dto.ID,
			"severity":      rule.Severity,
			"locked":        rule.Locked,
			"affectedShots": changedShots,
		},
	})

	profiles, _ := s.loadProfileRecords(root, manifest)
	impact := s.buildContinuityImpactIndex(root, manifest, index)
	impactReport := s.continuityImpactReport(impact, "", targetType, profile.dto.ID, rule.ID)
	return ContinuityResult{
		OK:     true,
		Rule:   &rule,
		Rules:  continuityRulesFromProfiles(profiles),
		Impact: &impactReport,
		Health: ptr(s.HealthReport(root)),
		Events: []ProjectEvent{s.event("continuity.rule.saved", "completed", "Continuity rule saved and affected Shots marked dirty.", correlationID, nil)},
	}
}

func (s *Store) UnlockContinuityRule(command UnlockContinuityRuleCommand) ContinuityResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.continuityFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before unlocking a continuity rule.")
	}
	reason := strings.TrimSpace(command.Reason)
	if reason == "" {
		return s.continuityFailure(CodeContinuityUnlockReason, SeverityBlocking, false, correlationID, "Unlock reason is required.", "Record why the locked continuity rule is being unlocked.")
	}
	if failure := s.continuityWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "continuity"
		return ContinuityResult{
			OK:     false,
			Rules:  []ContinuityRuleDTO{},
			Error:  &operationError,
			Events: []ProjectEvent{s.event("continuity.rule.unlock", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}
	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.continuityFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "Asset index could not be read.", err.Error())
	}
	profiles, err := s.loadProfileRecords(root, manifest)
	if err != nil {
		return s.continuityFailure(CodeBindingProfileInvalid, SeverityBlocking, true, correlationID, "Profile library could not be read.", err.Error())
	}

	ruleID := strings.TrimSpace(command.RuleID)
	profile, rule, ok := profileRecordForRule(profiles, ruleID)
	if !ok {
		return s.continuityFailure(CodeContinuityRuleMissing, SeverityBlocking, false, correlationID, "Continuity rule does not exist.", ruleID)
	}
	rule.Locked = false
	now := s.now().UTC()
	rule.UpdatedAt = now.Format(time.RFC3339)
	profile.dto.ContinuityRules = upsertContinuityRule(profile.dto.ContinuityRules, rule)
	profile.dto.LockedRules = syncLockedRuleIDs(profile.dto.LockedRules, profile.dto.ContinuityRules)
	profile.raw = profileRawWithBaseline(profile.raw, profile.dto)
	if err := s.writeProfileAndIndex(root, profile, index); err != nil {
		return s.continuityFailure(CodeContinuityWriteFailed, SeverityBlocking, true, correlationID, "Continuity rule unlock could not be written.", err.Error())
	}
	changedShots, dirtyErr := s.markShotsDirtyForContinuityChange(root, manifest, rule.TargetType, rule.TargetID, rule.ID, "continuity rule unlocked: "+rule.ID, correlationID, now)
	if dirtyErr != nil {
		return s.continuityFailure(CodeContinuityDirtyPropagate, SeverityBlocking, true, correlationID, "Continuity rule was unlocked, but affected Shot dirty marking failed.", dirtyErr.Error())
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("continuity.rule.unlocked", correlationID),
		EventType:     "continuity.rule.unlocked",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "Continuity rule unlocked with reason.",
		Details: map[string]any{
			"ruleId":        rule.ID,
			"targetType":    rule.TargetType,
			"targetId":      rule.TargetID,
			"reason":        reason,
			"affectedShots": changedShots,
		},
	})

	profiles, _ = s.loadProfileRecords(root, manifest)
	impact := s.buildContinuityImpactIndex(root, manifest, index)
	impactReport := s.continuityImpactReport(impact, "", rule.TargetType, rule.TargetID, rule.ID)
	return ContinuityResult{
		OK:     true,
		Rule:   &rule,
		Rules:  continuityRulesFromProfiles(profiles),
		Impact: &impactReport,
		Health: ptr(s.HealthReport(root)),
		Events: []ProjectEvent{s.event("continuity.rule.unlocked", "completed", "Continuity rule unlocked and affected Shots marked dirty.", correlationID, nil)},
	}
}

func (s *Store) UnlockAssetBinding(command UnlockAssetBindingCommand) ContinuityResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.continuityFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before unlocking an asset binding.")
	}
	reason := strings.TrimSpace(command.Reason)
	if reason == "" {
		return s.continuityFailure(CodeContinuityUnlockReason, SeverityBlocking, false, correlationID, "Unlock reason is required.", "Record why the locked asset binding is being unlocked.")
	}
	if failure := s.continuityWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "continuity"
		return ContinuityResult{
			OK:     false,
			Rules:  []ContinuityRuleDTO{},
			Error:  &operationError,
			Events: []ProjectEvent{s.event("continuity.binding.unlock", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}
	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.continuityFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "Asset index could not be read.", err.Error())
	}
	if !validBindingUnlockSelector(command) {
		return s.continuityFailure(CodeContinuityBindingSelector, SeverityBlocking, false, correlationID, "Asset binding unlock selector is incomplete.", "Provide a binding id, or provide asset id, target type, target id, and purpose.")
	}
	assetIndex, bindingIndex, binding, ok := findAssetBindingForUnlock(index.Assets, command)
	if !ok {
		return s.continuityFailure(CodeBindingTargetMissing, SeverityBlocking, false, correlationID, "Asset binding does not exist.", "Choose an existing binding id or target before unlocking.")
	}
	if !binding.Locked {
		impact := s.buildContinuityImpactIndex(root, manifest, index)
		impactReport := s.continuityImpactReport(impact, binding.AssetID, binding.TargetType, binding.TargetID, "")
		return ContinuityResult{
			OK:     true,
			Rules:  continuityRulesFromProfiles(mustProfileRecords(s.loadProfileRecords(root, manifest))),
			Impact: &impactReport,
			Health: ptr(s.HealthReport(root)),
			Events: []ProjectEvent{s.event("continuity.binding.unlocked", "completed", "Asset binding was already unlocked.", correlationID, nil)},
		}
	}
	now := s.now().UTC()
	index.Assets[assetIndex].Bindings[bindingIndex].Locked = false
	index.Assets[assetIndex].UpdatedAt = now.Format(time.RFC3339)
	index.UpdatedAt = now.Format(time.RFC3339)
	if err := s.writeAssetIndex(root, index); err != nil {
		return s.continuityFailure(CodeAssetIndexWriteFailed, SeverityBlocking, true, correlationID, "Asset binding unlock could not be written.", err.Error())
	}
	changedShots, dirtyErr := s.markShotsDirtyForContinuityChange(root, manifest, binding.TargetType, binding.TargetID, "", "asset binding unlocked: "+binding.ID, correlationID, now)
	if dirtyErr != nil {
		return s.continuityFailure(CodeContinuityDirtyPropagate, SeverityBlocking, true, correlationID, "Asset binding was unlocked, but affected Shot dirty marking failed.", dirtyErr.Error())
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("continuity.binding.unlocked", correlationID),
		EventType:     "continuity.binding.unlocked",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "Asset binding unlocked with reason.",
		Details: map[string]any{
			"bindingId":     binding.ID,
			"assetId":       binding.AssetID,
			"targetType":    binding.TargetType,
			"targetId":      binding.TargetID,
			"reason":        reason,
			"affectedShots": changedShots,
		},
	})

	impact := s.buildContinuityImpactIndex(root, manifest, index)
	impactReport := s.continuityImpactReport(impact, binding.AssetID, binding.TargetType, binding.TargetID, "")
	return ContinuityResult{
		OK:     true,
		Rules:  continuityRulesFromProfiles(mustProfileRecords(s.loadProfileRecords(root, manifest))),
		Impact: &impactReport,
		Health: ptr(s.HealthReport(root)),
		Events: []ProjectEvent{s.event("continuity.binding.unlocked", "completed", "Asset binding unlocked and affected Shots marked dirty.", correlationID, nil)},
	}
}

func (s *Store) buildContinuityImpactIndex(root string, manifest Manifest, index AssetIndex) continuityImpactIndex {
	impact := continuityImpactIndex{
		assetAffected:   map[string][]string{},
		bindingAffected: map[string][]string{},
		targetAffected:  map[string][]string{},
		assetShots:      map[string][]string{},
		targetShots:     map[string][]string{},
		ruleShots:       map[string][]string{},
		shotPackages:    map[string][]string{},
		targetNodeID:    map[string]string{},
		shotNodeID:      map[string]string{},
		assetsByID:      assetMapByID(index.Assets),
		profilesByKey:   map[string]profileRecord{},
		packagesByShot:  map[string][]packageManifest{},
	}
	for _, node := range manifest.Graph.Nodes {
		switch strings.TrimSpace(node.Kind) {
		case "character", "scene", "prop":
			if id := s.profileIDFromNodeRef(root, node.Kind, node.RefID); id != "" {
				impact.targetNodeID[profileKey(node.Kind, id)] = strings.TrimSpace(node.ID)
			}
		case "shot":
			if shotID := shotIDFromRef(node.RefID); shotID != "" {
				impact.shotNodeID[shotID] = strings.TrimSpace(node.ID)
			}
		}
	}

	profiles, err := s.loadProfileRecords(root, manifest)
	if err != nil {
		impact.incompleteReasons = append(impact.incompleteReasons, err.Error())
	} else {
		impact.profilesByKey = profiles
	}
	shots, err := s.readAllShotContextRecords(root, manifest)
	if err != nil {
		impact.incompleteReasons = append(impact.incompleteReasons, err.Error())
	} else {
		impact.shots = shots
	}
	packages := s.readPackageManifests(root, manifest)
	impact.packagesByShot = packages
	for shotID, manifests := range packages {
		for _, item := range manifests {
			impact.shotPackages[shotID] = addUniqueString(impact.shotPackages[shotID], item.PackageID)
		}
	}

	for _, record := range shots {
		shot := record.shot
		for _, targetID := range shot.CharacterIDs {
			impact.addTargetShot(BindingTargetCharacter, targetID, shot.ID)
		}
		for _, targetID := range cleanStringList(append([]string{shot.SceneID, shot.SceneProfileID}, shot.SceneIDRefs...)) {
			impact.addTargetShot(BindingTargetScene, targetID, shot.ID)
		}
		for _, targetID := range shot.PropIDs {
			impact.addTargetShot(BindingTargetProp, targetID, shot.ID)
		}
		for _, ruleID := range shot.ContinuityRuleIDs {
			impact.ruleShots[ruleID] = addUniqueString(impact.ruleShots[ruleID], shot.ID)
		}
	}

	for _, asset := range index.Assets {
		affected := []string{asset.ID}
		for _, binding := range asset.Bindings {
			binding = normalizeAssetBinding(asset.ID, binding)
			if binding.ID != "" {
				affected = addUniqueString(affected, binding.ID)
			}
			targetKey := profileKey(binding.TargetType, binding.TargetID)
			targetAffected := []string{binding.TargetID}
			if nodeID := impact.targetNodeID[targetKey]; nodeID != "" {
				targetAffected = addUniqueString(targetAffected, nodeID)
				affected = addUniqueString(affected, nodeID)
			}
			affected = addUniqueString(affected, binding.TargetID)
			if _, ok := profiles[targetKey]; !ok {
				impact.bindingAffected[binding.ID] = cleanStringList(append(affected, targetAffected...))
				continue
			}
			impact.targetAffected[targetKey] = cleanStringList(append(impact.targetAffected[targetKey], targetAffected...))
		}
		for _, record := range shots {
			if shotReferencesAsset(record.shot, asset.ID) {
				impact.assetShots[asset.ID] = addUniqueString(impact.assetShots[asset.ID], record.shot.ID)
				affected = addShotAffectedObjects(affected, record.shot.ID, impact)
			}
		}
		for _, binding := range asset.Bindings {
			binding = normalizeAssetBinding(asset.ID, binding)
			for _, shotID := range impact.shotIDsForTarget(binding.TargetType, binding.TargetID) {
				impact.assetShots[asset.ID] = addUniqueString(impact.assetShots[asset.ID], shotID)
				affected = addShotAffectedObjects(affected, shotID, impact)
			}
		}
		impact.assetAffected[asset.ID] = cleanStringList(affected)
	}
	return impact
}

func (impact continuityImpactIndex) shotIDsForTarget(targetType string, targetID string) []string {
	return impact.targetShots[profileKey(targetType, targetID)]
}

func (impact *continuityImpactIndex) addTargetShot(targetType string, targetID string, shotID string) {
	targetID = strings.TrimSpace(targetID)
	shotID = strings.TrimSpace(shotID)
	if targetID == "" || shotID == "" {
		return
	}
	key := profileKey(targetType, targetID)
	impact.targetShots[key] = addUniqueString(impact.targetShots[key], shotID)
	affected := []string{targetID}
	if nodeID := impact.targetNodeID[key]; nodeID != "" {
		affected = addUniqueString(affected, nodeID)
	}
	affected = addShotAffectedObjects(affected, shotID, *impact)
	impact.targetAffected[key] = cleanStringList(append(impact.targetAffected[key], affected...))
}

func addShotAffectedObjects(values []string, shotID string, impact continuityImpactIndex) []string {
	values = addUniqueString(values, shotID)
	if nodeID := impact.shotNodeID[shotID]; nodeID != "" {
		values = addUniqueString(values, nodeID)
	}
	for _, packageID := range impact.shotPackages[shotID] {
		values = addUniqueString(values, packageID)
	}
	return values
}

func (s *Store) continuityImpactReport(impact continuityImpactIndex, assetID string, targetType string, targetID string, ruleID string) ContinuityImpactReportDTO {
	report := ContinuityImpactReportDTO{
		AffectedAssets:   []string{},
		AffectedBindings: []string{},
		AffectedProfiles: []string{},
		AffectedShots:    []string{},
		AffectedPackages: []string{},
		Issues:           []ContinuityImpactIssueDTO{},
		RecoveryActions:  []string{},
		CheckedAt:        s.now().UTC().Format(time.RFC3339),
	}
	if assetID == "" && targetType == "" && targetID == "" && ruleID == "" {
		for assetID := range impact.assetAffected {
			report.AffectedAssets = addUniqueString(report.AffectedAssets, assetID)
		}
		for _, asset := range impact.assetsByID {
			for _, binding := range asset.Bindings {
				binding = normalizeAssetBinding(asset.ID, binding)
				report.AffectedBindings = addUniqueString(report.AffectedBindings, binding.ID)
			}
		}
		for _, profile := range impact.profilesByKey {
			report.AffectedProfiles = addUniqueString(report.AffectedProfiles, profile.dto.ID)
		}
		for _, shotIDs := range impact.assetShots {
			for _, shotID := range shotIDs {
				report.AffectedShots = addUniqueString(report.AffectedShots, shotID)
			}
		}
		for _, shotIDs := range impact.targetShots {
			for _, shotID := range shotIDs {
				report.AffectedShots = addUniqueString(report.AffectedShots, shotID)
			}
		}
		for _, shotIDs := range impact.ruleShots {
			for _, shotID := range shotIDs {
				report.AffectedShots = addUniqueString(report.AffectedShots, shotID)
			}
		}
		for _, packageIDs := range impact.shotPackages {
			for _, packageID := range packageIDs {
				report.AffectedPackages = addUniqueString(report.AffectedPackages, packageID)
			}
		}
	}
	if assetID != "" {
		report.AffectedAssets = addUniqueString(report.AffectedAssets, assetID)
		for _, shotID := range impact.assetShots[assetID] {
			report.AffectedShots = addUniqueString(report.AffectedShots, shotID)
		}
	}
	if targetType != "" && targetID != "" {
		targetKey := profileKey(targetType, targetID)
		report.AffectedProfiles = addUniqueString(report.AffectedProfiles, targetID)
		for _, shotID := range impact.targetShots[targetKey] {
			report.AffectedShots = addUniqueString(report.AffectedShots, shotID)
		}
	}
	if ruleID != "" {
		for _, shotID := range impact.ruleShots[ruleID] {
			report.AffectedShots = addUniqueString(report.AffectedShots, shotID)
		}
	}
	for _, shotID := range report.AffectedShots {
		for _, packageID := range impact.shotPackages[shotID] {
			report.AffectedPackages = addUniqueString(report.AffectedPackages, packageID)
		}
	}
	for _, asset := range impact.assetsByID {
		for _, binding := range asset.Bindings {
			binding = normalizeAssetBinding(asset.ID, binding)
			if assetID != "" && binding.AssetID != assetID {
				continue
			}
			if targetType != "" && targetID != "" && (binding.TargetType != targetType || binding.TargetID != targetID) {
				continue
			}
			report.AffectedBindings = addUniqueString(report.AffectedBindings, binding.ID)
		}
	}
	for _, reason := range impact.incompleteReasons {
		report.Issues = append(report.Issues, ContinuityImpactIssueDTO{
			Code:            CodeContinuityImpactIncomplete,
			Severity:        SeverityWarning,
			UserMessage:     "Continuity impact scan could not inspect every object.",
			TargetID:        reason,
			RecoveryActions: []string{"Run project health check and retry the continuity scan after fixing unreadable files."},
		})
	}
	report.RecoveryActions = continuityRecoveryActionsForReport(report)
	sortContinuityImpactReport(&report)
	return report
}

func (s *Store) markShotsDirtyForContinuityChange(root string, manifest Manifest, targetType string, targetID string, ruleID string, reason string, correlationID string, now time.Time) ([]string, error) {
	records, err := s.readAllShotContextRecords(root, manifest)
	if err != nil {
		return nil, err
	}
	var changed []string
	for _, record := range records {
		if !shotReferencesTarget(record.shot, targetType, targetID) && (strings.TrimSpace(ruleID) == "" || !stringListContains(record.shot.ContinuityRuleIDs, ruleID)) {
			continue
		}
		if record.shot.Status == ShotStatusContextDirty {
			continue
		}
		shot := record.shot
		shot.Status = ShotStatusContextDirty
		shot.UpdatedAt = now.UTC().Format(time.RFC3339)
		patchShotRaw(record.raw, shot)
		if err := s.writeShotRaw(root, manifest, shot.ID, record.raw); err != nil {
			return changed, err
		}
		changed = addUniqueString(changed, shot.ID)
	}
	if len(changed) == 0 {
		return changed, nil
	}
	if err := s.updateManifestShotStatuses(root, manifest, changed, ShotStatusContextDirty, now); err != nil {
		return changed, err
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("shot.context_dirty.propagated", correlationID),
		EventType:     "shot.context_dirty.propagated",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.UTC().Format(time.RFC3339),
		Summary:       "Shot context dirty state propagated from continuity change.",
		Details: map[string]any{
			"targetType": targetType,
			"targetId":   targetID,
			"ruleId":     ruleID,
			"reason":     strings.TrimSpace(reason),
			"shotIds":    changed,
		},
	})
	return changed, nil
}

func (s *Store) updateManifestShotStatuses(root string, manifest Manifest, shotIDs []string, status string, now time.Time) error {
	wanted := map[string]struct{}{}
	for _, shotID := range cleanStringList(shotIDs) {
		if nodeID := manifestShotNodeID(root, manifest, shotID); nodeID != "" {
			wanted[nodeID] = struct{}{}
		}
	}
	updated := false
	for index := range manifest.Graph.Nodes {
		node := &manifest.Graph.Nodes[index]
		if strings.TrimSpace(node.Kind) != "shot" {
			continue
		}
		if _, ok := wanted[strings.TrimSpace(node.ID)]; ok {
			node.Status = status
			updated = true
		}
	}
	packageUpdated, err := s.markPackagesStaleForShots(root, &manifest, shotIDs, now)
	if err != nil {
		return err
	}
	if !updated && !packageUpdated {
		return nil
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

func (s *Store) readAllShotContextRecords(root string, manifest Manifest) ([]shotContextRecord, error) {
	dirRelative := path.Clean(strings.ReplaceAll(manifest.Paths.Shots, "\\", "/"))
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dirRelative)))
	if errors.Is(err, os.ErrNotExist) {
		return []shotContextRecord{}, nil
	}
	if err != nil {
		return nil, err
	}
	records := make([]shotContextRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		shotID := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		relative := path.Join(dirRelative, entry.Name())
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return nil, err
		}
		var value struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(data, &value) == nil && strings.TrimSpace(value.ID) != "" {
			shotID = strings.TrimSpace(value.ID)
		}
		record, result := s.readShotContextRecord(root, manifest, shotID, "continuity-impact")
		if result != nil {
			return nil, fmt.Errorf("%s: %s", entry.Name(), result.Error.TechnicalDetail)
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].shot.ID < records[j].shot.ID })
	return records, nil
}

func (s *Store) readPackageManifests(root string, manifest Manifest) map[string][]packageManifest {
	results := map[string][]packageManifest{}
	packagesRoot := filepath.Join(root, filepath.FromSlash(manifest.Paths.Packages))
	_ = filepath.WalkDir(packagesRoot, func(filename string, entry os.DirEntry, err error) error {
		if err != nil || entry == nil || entry.IsDir() || entry.Name() != "manifest.json" {
			return nil
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil
		}
		var packageData packageManifest
		if json.Unmarshal(data, &packageData) != nil {
			return nil
		}
		relative, err := filepath.Rel(root, filename)
		if err == nil {
			packageData.ManifestPath = filepath.ToSlash(relative)
		}
		if packageData.ShotID != "" {
			results[packageData.ShotID] = append(results[packageData.ShotID], packageData)
		}
		return nil
	})
	return results
}

func (s *Store) profileIDFromNodeRef(root string, targetType string, refID string) string {
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return ""
	}
	filename, ok := safeProjectFilePath(root, refID)
	if !ok {
		return ""
	}
	record, err := readProfileRecordFile(filename, path.Clean(strings.ReplaceAll(refID, "\\", "/")), targetType)
	if err != nil {
		return ""
	}
	return record.dto.ID
}

func shotReferencesTarget(shot ShotCardDTO, targetType string, targetID string) bool {
	targetID = strings.TrimSpace(targetID)
	if targetID == "" {
		return false
	}
	switch targetType {
	case BindingTargetCharacter:
		return stringListContains(shot.CharacterIDs, targetID)
	case BindingTargetScene:
		if shot.SceneID == targetID || shot.SceneProfileID == targetID {
			return true
		}
		return stringListContains(shot.SceneIDRefs, targetID)
	case BindingTargetProp:
		return stringListContains(shot.PropIDs, targetID)
	default:
		return false
	}
}

func shotReferencesAsset(shot ShotCardDTO, assetID string) bool {
	if stringListContains(shot.ReferenceAssetIDs, assetID) {
		return true
	}
	for _, ref := range shot.CharacterRefs {
		if strings.TrimSpace(ref.ReferenceAssetID) == assetID {
			return true
		}
	}
	return false
}

func stringListContains(values []string, want string) bool {
	want = strings.TrimSpace(want)
	if want == "" {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == want {
			return true
		}
	}
	return false
}

func shotIDFromRef(refID string) string {
	refID = path.Clean(strings.ReplaceAll(strings.TrimSpace(refID), "\\", "/"))
	if refID == "." || path.Ext(refID) != ".json" {
		return ""
	}
	if path.Base(path.Dir(refID)) != "shots" {
		return ""
	}
	return strings.TrimSuffix(path.Base(refID), ".json")
}

func profileRecordForRule(profiles map[string]profileRecord, ruleID string) (profileRecord, ContinuityRuleDTO, bool) {
	ruleID = strings.TrimSpace(ruleID)
	for _, profile := range profiles {
		if rule := findContinuityRule(profile.dto.ContinuityRules, ruleID); rule != nil {
			return profile, *rule, true
		}
		if stringListContains(profile.dto.LockedRules, ruleID) {
			rule := ContinuityRuleDTO{
				ID:         ruleID,
				TargetType: profile.dto.Type,
				TargetID:   profile.dto.ID,
				Rule:       "Legacy locked continuity rule " + ruleID,
				Severity:   ContinuitySeverityBlocking,
				Locked:     true,
				CreatedBy:  "system",
			}
			return profile, rule, true
		}
	}
	return profileRecord{}, ContinuityRuleDTO{}, false
}

func findAssetBindingForUnlock(assets []AssetDTO, command UnlockAssetBindingCommand) (int, int, AssetBindingDTO, bool) {
	bindingID := strings.TrimSpace(command.BindingID)
	assetID := strings.TrimSpace(command.AssetID)
	targetType, _ := normalizeBindingTargetType(command.TargetType)
	targetID := strings.TrimSpace(command.TargetID)
	purpose := normalizeBindingPurpose(command.Purpose)
	for assetIndex := range assets {
		asset := assets[assetIndex]
		if assetID != "" && asset.ID != assetID {
			continue
		}
		for bindingIndex := range asset.Bindings {
			binding := normalizeAssetBinding(asset.ID, asset.Bindings[bindingIndex])
			if bindingID != "" && binding.ID != bindingID {
				continue
			}
			if bindingID == "" {
				if targetType != "" && binding.TargetType != targetType {
					continue
				}
				if targetID != "" && binding.TargetID != targetID {
					continue
				}
				if strings.TrimSpace(command.Purpose) != "" && binding.Purpose != purpose {
					continue
				}
			}
			return assetIndex, bindingIndex, binding, true
		}
	}
	return -1, -1, AssetBindingDTO{}, false
}

func validBindingUnlockSelector(command UnlockAssetBindingCommand) bool {
	if strings.TrimSpace(command.BindingID) != "" {
		return true
	}
	if strings.TrimSpace(command.AssetID) == "" || strings.TrimSpace(command.TargetID) == "" || strings.TrimSpace(command.Purpose) == "" {
		return false
	}
	_, ok := normalizeBindingTargetType(command.TargetType)
	return ok
}

func parseContinuityRules(raw map[string]any, targetType string, targetID string, lockedRules []string) []ContinuityRuleDTO {
	var rules []ContinuityRuleDTO
	if rawValue, ok := raw["continuityRules"]; ok {
		data, err := json.Marshal(rawValue)
		if err == nil {
			_ = json.Unmarshal(data, &rules)
		}
	}
	for index := range rules {
		rules[index] = normalizeContinuityRule(rules[index], targetType, targetID)
	}
	for _, ruleID := range lockedRules {
		if findContinuityRule(rules, ruleID) != nil {
			continue
		}
		rules = append(rules, ContinuityRuleDTO{
			ID:         ruleID,
			TargetType: targetType,
			TargetID:   targetID,
			Rule:       "Legacy locked continuity rule " + ruleID,
			Severity:   ContinuitySeverityBlocking,
			Locked:     true,
			CreatedBy:  "system",
		})
	}
	sortContinuityRules(rules)
	return rules
}

func normalizeContinuityRule(rule ContinuityRuleDTO, fallbackType string, fallbackID string) ContinuityRuleDTO {
	rule.ID = strings.TrimSpace(rule.ID)
	rule.TargetType, _ = normalizeBindingTargetType(firstNonEmpty(rule.TargetType, fallbackType))
	if rule.TargetType == "" {
		rule.TargetType = fallbackType
	}
	rule.TargetID = firstNonEmpty(rule.TargetID, fallbackID)
	rule.Rule = strings.TrimSpace(rule.Rule)
	rule.Severity = normalizeContinuitySeverity(rule.Severity)
	rule.CreatedBy = normalizeCreatedBy(rule.CreatedBy)
	if rule.ID == "" {
		rule.ID = continuityRuleID(rule.TargetType, rule.TargetID, rule.Rule)
	}
	return rule
}

func normalizeContinuitySeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ContinuitySeverityBlocking, "":
		return ContinuitySeverityBlocking
	case ContinuitySeverityWarning:
		return ContinuitySeverityWarning
	case ContinuitySeveritySuggestion:
		return ContinuitySeveritySuggestion
	default:
		return ContinuitySeverityWarning
	}
}

func continuityRuleID(targetType string, targetID string, rule string) string {
	token := safeFileToken(targetType + "_" + targetID + "_" + rule)
	if len(token) > 64 {
		token = token[:64]
	}
	return "rule_" + strings.Trim(token, "_")
}

func findContinuityRule(rules []ContinuityRuleDTO, ruleID string) *ContinuityRuleDTO {
	for index := range rules {
		if strings.TrimSpace(rules[index].ID) == strings.TrimSpace(ruleID) {
			return &rules[index]
		}
	}
	return nil
}

func upsertContinuityRule(rules []ContinuityRuleDTO, rule ContinuityRuleDTO) []ContinuityRuleDTO {
	rule = normalizeContinuityRule(rule, rule.TargetType, rule.TargetID)
	for index := range rules {
		if strings.TrimSpace(rules[index].ID) == rule.ID {
			rules[index] = rule
			sortContinuityRules(rules)
			return rules
		}
	}
	rules = append(rules, rule)
	sortContinuityRules(rules)
	return rules
}

func syncLockedRuleIDs(existing []string, rules []ContinuityRuleDTO) []string {
	ids := cleanStringList(existing)[:0]
	for _, rule := range rules {
		if rule.Locked {
			ids = addUniqueString(ids, rule.ID)
		}
	}
	return ids
}

func continuityRuleIDs(rules []ContinuityRuleDTO) []string {
	ids := make([]string, 0, len(rules))
	for _, rule := range rules {
		ids = addUniqueString(ids, rule.ID)
	}
	sort.Strings(ids)
	return ids
}

func continuityRulesFromProfiles(profiles map[string]profileRecord) []ContinuityRuleDTO {
	var rules []ContinuityRuleDTO
	for _, profile := range profiles {
		rules = append(rules, profile.dto.ContinuityRules...)
	}
	sortContinuityRules(rules)
	return rules
}

func sortContinuityRules(rules []ContinuityRuleDTO) {
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].TargetType == rules[j].TargetType {
			if rules[i].TargetID == rules[j].TargetID {
				return rules[i].ID < rules[j].ID
			}
			return rules[i].TargetID < rules[j].TargetID
		}
		return rules[i].TargetType < rules[j].TargetType
	})
}

func sortContinuityImpactReport(report *ContinuityImpactReportDTO) {
	sort.Strings(report.AffectedAssets)
	sort.Strings(report.AffectedBindings)
	sort.Strings(report.AffectedProfiles)
	sort.Strings(report.AffectedShots)
	sort.Strings(report.AffectedPackages)
	sort.Strings(report.RecoveryActions)
	sort.SliceStable(report.Issues, func(i, j int) bool {
		if report.Issues[i].Severity == report.Issues[j].Severity {
			return report.Issues[i].Code < report.Issues[j].Code
		}
		return report.Issues[i].Severity < report.Issues[j].Severity
	})
}

func continuityRecoveryActionsForReport(report ContinuityImpactReportDTO) []string {
	var actions []string
	if len(report.AffectedShots) > 0 {
		actions = addUniqueString(actions, "Review affected Shots before package export or provider handoff.")
	}
	if len(report.AffectedAssets) > 0 || len(report.AffectedBindings) > 0 {
		actions = addUniqueString(actions, "Relink missing assets or unlock bindings with a recorded reason before destructive changes.")
	}
	if len(report.Issues) > 0 {
		actions = addUniqueString(actions, "Run project health check after resolving continuity issues.")
	}
	return actions
}

func (s *Store) continuityWriteLockFailure(root string, correlationID string) *ContinuityResult {
	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		result := s.continuityFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "Project is open in another active session.", "Open the project read-only or confirm takeover before editing continuity.")
		return &result
	}
	if lockInfo.State == LockStateStale {
		result := s.continuityFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "Project has a stale lock that needs takeover before editing continuity.", "Open the project with takeover, then retry.")
		return &result
	}
	return nil
}

func (s *Store) continuityFailure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string) ContinuityResult {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, continuityRecoveryActions(code, technicalDetail))
	err.TargetType = "continuity"
	return ContinuityResult{
		OK:     false,
		Rules:  []ContinuityRuleDTO{},
		Error:  &err,
		Events: []ProjectEvent{s.event("continuity.operation", "blocked", userMessage, correlationID, &err)},
	}
}

func continuityRecoveryActions(code string, detail string) []string {
	switch code {
	case CodeContinuityUnlockReason:
		return []string{"Enter an unlock reason so the audit trail explains the continuity override."}
	case CodeContinuityRuleInvalid:
		return []string{"Choose an existing Character, Scene, or Prop target and enter a non-empty rule."}
	case CodeContinuityRuleMissing:
		return []string{"Refresh continuity rules and choose an existing locked rule before retrying."}
	case CodeContinuityLockedBinding:
		return []string{"Unlock the binding with a reason before removing or replacing a locked continuity reference."}
	case CodeContinuityBindingSelector:
		return []string{"Choose a concrete binding id, or provide asset id, target type, target id, and purpose before unlocking."}
	case CodeContinuityDirtyPropagate:
		return []string{"Run project health check, then retry after restoring unreadable Shot files."}
	case CodeContinuityWriteFailed:
		return []string{"Retry the continuity operation. If it fails again, run project health check before continuing."}
	default:
		if strings.TrimSpace(detail) != "" {
			return []string{detail}
		}
		return []string{"Review continuity input and retry."}
	}
}
