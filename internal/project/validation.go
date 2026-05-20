package project

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	SeverityBlocking = "blocking"
	SeverityWarning  = "warning"
	SeverityInfo     = "info"
)

const (
	CodeManifestJSONInvalid     = "project_manifest_json_invalid"
	CodeManifestRequiredMissing = "project_manifest_required_missing"
	CodeSchemaUnsupported       = "project_schema_unsupported"
	CodePathEmpty               = "project_path_empty"
	CodePathAbsolute            = "project_path_absolute"
	CodePathTraversal           = "project_path_traversal"
	CodePathURL                 = "project_path_url"
	CodeSensitiveValue          = "project_sensitive_value"
	CodeGraphPartitionMissing   = "project_graph_partition_missing"
	CodeTimestampInvalid        = "project_timestamp_invalid"
)

type ValidationIssue struct {
	Severity        string   `json:"severity"`
	Code            string   `json:"code"`
	Path            string   `json:"path"`
	UserMessage     string   `json:"userMessage"`
	TechnicalDetail string   `json:"technicalDetail,omitempty"`
	RecoveryActions []string `json:"recoveryActions"`
}

type ValidationReport struct {
	Items []ValidationIssue `json:"items"`
}

func (report ValidationReport) Status() string {
	if report.HasBlocking() {
		return SeverityBlocking
	}
	if len(report.Items) > 0 {
		return SeverityWarning
	}
	return SeverityInfo
}

func (report ValidationReport) HasBlocking() bool {
	for _, item := range report.Items {
		if item.Severity == SeverityBlocking {
			return true
		}
	}
	return false
}

func (report ValidationReport) Codes() []string {
	codes := make([]string, 0, len(report.Items))
	for _, item := range report.Items {
		codes = append(codes, item.Code)
	}
	sort.Strings(codes)
	return codes
}

type ValidationError struct {
	Report ValidationReport
}

func (err ValidationError) Error() string {
	return fmt.Sprintf("project manifest validation failed: %v", err.Report.Codes())
}

func ValidateManifest(manifest Manifest) ValidationReport {
	var issues []ValidationIssue

	if strings.TrimSpace(manifest.SchemaVersion) == "" {
		issues = append(issues, missingRequiredIssue("schemaVersion"))
	} else if manifest.SchemaVersion != CurrentSchemaVersion {
		issues = append(issues, ValidationIssue{
			Severity:        SeverityBlocking,
			Code:            CodeSchemaUnsupported,
			Path:            "schemaVersion",
			UserMessage:     "Project schema version is not supported by this Alpha shell.",
			TechnicalDetail: "schemaVersion=" + manifest.SchemaVersion,
			RecoveryActions: []string{"Open the project with a compatible app version or export a diagnostic copy."},
		})
	}

	issues = append(issues, validateProjectMetadata(manifest.Project)...)
	issues = append(issues, validatePaths(manifest.Paths)...)
	issues = append(issues, validateGraph(manifest.Graph)...)
	issues = append(issues, scanManifestValues(manifest)...)

	return ValidationReport{Items: dedupeIssues(issues)}
}

func validateProjectMetadata(project ProjectMetadata) []ValidationIssue {
	required := map[string]string{
		"project.id":        project.ID,
		"project.name":      project.Name,
		"project.type":      project.Type,
		"project.createdAt": project.CreatedAt,
		"project.updatedAt": project.UpdatedAt,
	}

	var issues []ValidationIssue
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			issues = append(issues, missingRequiredIssue(field))
		}
	}

	for field, value := range map[string]string{
		"project.createdAt": project.CreatedAt,
		"project.updatedAt": project.UpdatedAt,
	} {
		if value == "" {
			continue
		}
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			issues = append(issues, ValidationIssue{
				Severity:        SeverityBlocking,
				Code:            CodeTimestampInvalid,
				Path:            field,
				UserMessage:     "Project manifest timestamp is invalid.",
				TechnicalDetail: err.Error(),
				RecoveryActions: []string{"Rewrite the timestamp as RFC3339 UTC or restore the manifest from backup."},
			})
		}
	}

	return issues
}

func validatePaths(paths Paths) []ValidationIssue {
	allPaths := paths.All()
	issues := make([]ValidationIssue, 0)

	for key, value := range allPaths {
		field := "paths." + key
		issues = append(issues, validateProjectPath(field, value)...)
	}

	return issues
}

func validateProjectPath(field string, value string) []ValidationIssue {
	value = strings.TrimSpace(value)
	if value == "" {
		return []ValidationIssue{{
			Severity:        SeverityBlocking,
			Code:            CodePathEmpty,
			Path:            field,
			UserMessage:     "Project manifest path is empty.",
			RecoveryActions: []string{"Restore the default project-relative path."},
		}}
	}

	if isAbsoluteOrHomePath(value) {
		return []ValidationIssue{{
			Severity:        SeverityBlocking,
			Code:            CodePathAbsolute,
			Path:            field,
			UserMessage:     "Project manifest path must be project-relative.",
			TechnicalDetail: value,
			RecoveryActions: []string{"Replace the path with a relative path inside the project root."},
		}}
	}

	if hasURLScheme(value) {
		return []ValidationIssue{{
			Severity:        SeverityBlocking,
			Code:            CodePathURL,
			Path:            field,
			UserMessage:     "Project manifest path must not be a URL.",
			TechnicalDetail: value,
			RecoveryActions: []string{"Use a project-relative path and store external references as explicit managed references later."},
		}}
	}

	clean := path.Clean(strings.ReplaceAll(value, "\\", "/"))
	if clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return []ValidationIssue{{
			Severity:        SeverityBlocking,
			Code:            CodePathTraversal,
			Path:            field,
			UserMessage:     "Project manifest path must not escape the project root.",
			TechnicalDetail: value,
			RecoveryActions: []string{"Keep the path under the project root and remove '..' traversal."},
		}}
	}

	return nil
}

func validateGraph(graph Graph) []ValidationIssue {
	var issues []ValidationIssue
	if graph.Version <= 0 {
		issues = append(issues, missingGraphIssue("graph.version"))
	}
	if graph.Nodes == nil {
		issues = append(issues, missingGraphIssue("graph.nodes"))
	}
	if graph.Edges == nil {
		issues = append(issues, missingGraphIssue("graph.edges"))
	}
	if graph.Viewport.Zoom <= 0 {
		issues = append(issues, missingGraphIssue("graph.viewport.zoom"))
	}
	return issues
}

func validateRequiredRawFields(raw map[string]json.RawMessage) []ValidationIssue {
	required := []string{
		"schemaVersion",
		"project",
		"paths",
		"defaults",
		"integrity",
		"graph",
		"principles",
		"styleBible",
	}

	var issues []ValidationIssue
	for _, field := range required {
		if _, ok := raw[field]; !ok {
			issues = append(issues, missingRequiredIssue(field))
		}
	}

	if graphRaw, ok := raw["graph"]; ok {
		var graph map[string]json.RawMessage
		if err := json.Unmarshal(graphRaw, &graph); err == nil {
			for _, field := range []string{"version", "nodes", "edges", "viewport"} {
				if _, ok := graph[field]; !ok {
					issues = append(issues, missingGraphIssue("graph."+field))
				}
			}
		}
	}

	return issues
}

func scanManifestValues(manifest Manifest) []ValidationIssue {
	data, err := json.Marshal(manifest)
	if err != nil {
		return []ValidationIssue{{
			Severity:        SeverityBlocking,
			Code:            CodeManifestJSONInvalid,
			Path:            ManifestFileName,
			UserMessage:     "Project manifest cannot be encoded for validation.",
			TechnicalDetail: err.Error(),
			RecoveryActions: []string{"Inspect the manifest fields and remove invalid values."},
		}}
	}

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil
	}

	var issues []ValidationIssue
	walkStringValues("", value, func(path string, text string) {
		if looksSensitive(text) {
			issues = append(issues, ValidationIssue{
				Severity:        SeverityBlocking,
				Code:            CodeSensitiveValue,
				Path:            path,
				UserMessage:     "Project manifest contains a value that looks like private machine data or a credential.",
				TechnicalDetail: redactSensitiveDetail(text),
				RecoveryActions: []string{"Replace the value with a project-relative path or non-sensitive placeholder."},
			})
		}
	})

	return issues
}

func walkStringValues(prefix string, value any, visit func(string, string)) {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			next := key
			if prefix != "" {
				next = prefix + "." + key
			}
			walkStringValues(next, typed[key], visit)
		}
	case []any:
		for i, item := range typed {
			walkStringValues(fmt.Sprintf("%s[%d]", prefix, i), item, visit)
		}
	case string:
		visit(prefix, typed)
	}
}

func missingRequiredIssue(path string) ValidationIssue {
	return ValidationIssue{
		Severity:        SeverityBlocking,
		Code:            CodeManifestRequiredMissing,
		Path:            path,
		UserMessage:     "Project manifest is missing a required field.",
		RecoveryActions: []string{"Create the project from the current baseline template or restore project.tuyu.json from backup."},
	}
}

func missingGraphIssue(path string) ValidationIssue {
	return ValidationIssue{
		Severity:        SeverityBlocking,
		Code:            CodeGraphPartitionMissing,
		Path:            path,
		UserMessage:     "Project manifest graph partition is incomplete.",
		RecoveryActions: []string{"Restore the graph partition from a valid baseline manifest or backup."},
	}
}

func hasURLScheme(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme != ""
}

func isAbsoluteOrHomePath(value string) bool {
	if filepath.IsAbs(value) || strings.HasPrefix(value, "~") {
		return true
	}
	return regexp.MustCompile(`^[A-Za-z]:`).MatchString(value) || strings.HasPrefix(value, `\`)
}

func looksSensitive(value string) bool {
	if value == "" {
		return false
	}

	lower := strings.ToLower(value)
	if isAbsoluteOrHomePath(value) || looksLikeExternalReference(value) {
		return true
	}

	if strings.HasPrefix(value, "sk-") || strings.HasPrefix(value, "ghp_") || strings.HasPrefix(value, "xox") {
		return true
	}

	return strings.Contains(lower, "api_key=") ||
		strings.Contains(lower, "apikey=") ||
		strings.Contains(lower, "token=") ||
		strings.Contains(lower, "secret=") ||
		strings.Contains(lower, "credential=")
}

func looksLikeExternalReference(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" {
		return false
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "file", "ftp", "s3", "gs", "data", "ipfs":
		return true
	default:
		return parsed.Host != "" || strings.Contains(value, "://")
	}
}

func redactSensitiveDetail(value string) string {
	if len(value) <= 8 {
		return "<redacted>"
	}
	return value[:4] + "<redacted>"
}

func dedupeIssues(issues []ValidationIssue) []ValidationIssue {
	seen := make(map[string]bool, len(issues))
	deduped := make([]ValidationIssue, 0, len(issues))

	for _, issue := range issues {
		key := issue.Code + "\x00" + issue.Path
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, issue)
	}

	return deduped
}
