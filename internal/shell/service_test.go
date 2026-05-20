package shell

import (
	"testing"
	"time"
)

func TestServiceInfo(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 3, 0, 0, 0, time.UTC)
	info := NewService(startedAt).Info()

	if info.AppName != "Tuyu Studio" {
		t.Fatalf("AppName = %q, want Tuyu Studio", info.AppName)
	}
	if info.Stage != "alpha-shell" {
		t.Fatalf("Stage = %q, want alpha-shell", info.Stage)
	}
	if info.StartedAt != "2026-05-20T03:00:00Z" {
		t.Fatalf("StartedAt = %q, want RFC3339 UTC", info.StartedAt)
	}
	if len(info.Capabilities) != 9 {
		t.Fatalf("Capabilities length = %d, want 9", len(info.Capabilities))
	}
}

func TestServiceHealthDocumentsDomainBoundary(t *testing.T) {
	health := NewService(time.Now()).Health()

	if health.Status != "ready" {
		t.Fatalf("Status = %q, want ready", health.Status)
	}
	if health.Severity != "info" {
		t.Fatalf("Severity = %q, want info", health.Severity)
	}
	if health.TechnicalDetail == "" {
		t.Fatal("TechnicalDetail must document the TOO-160 domain-service boundary")
	}
	if len(health.RecoveryActions) == 0 {
		t.Fatal("RecoveryActions must include the next implementation step")
	}
}

func TestWorkbenchProbeStatus(t *testing.T) {
	result := NewService(time.Now()).WorkbenchProbe(WorkbenchProbeCommand{
		Mode:          "status",
		CorrelationID: "test-correlation",
	})

	if !result.OK {
		t.Fatalf("OK = false, want true: %#v", result.Error)
	}
	if result.Snapshot.ServiceName != "shell.workbench" {
		t.Fatalf("ServiceName = %q, want shell.workbench", result.Snapshot.ServiceName)
	}
	if result.Error != nil {
		t.Fatalf("Error = %#v, want nil", result.Error)
	}
	if len(result.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(result.Events))
	}
	if result.Events[0].State != "completed" {
		t.Fatalf("Event state = %q, want completed", result.Events[0].State)
	}
	if result.Events[0].EventID == "" {
		t.Fatal("EventID must be set")
	}
}

func TestWorkbenchProbeStructuredError(t *testing.T) {
	result := NewService(time.Now()).WorkbenchProbe(WorkbenchProbeCommand{
		Mode:          "structured_error",
		CorrelationID: "test-error",
	})

	if result.OK {
		t.Fatal("OK = true, want false")
	}
	if result.Error == nil {
		t.Fatal("Error = nil, want AppError")
	}
	if result.Error.Code != "workbench_probe_blocked" {
		t.Fatalf("Error code = %q, want workbench_probe_blocked", result.Error.Code)
	}
	if result.Error.Severity != "blocking" {
		t.Fatalf("Error severity = %q, want blocking", result.Error.Severity)
	}
	if len(result.Error.RecoveryActions) == 0 {
		t.Fatal("RecoveryActions must be set")
	}
	if len(result.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(result.Events))
	}
	if result.Events[0].Error == nil {
		t.Fatal("Event error must carry AppError")
	}
}
