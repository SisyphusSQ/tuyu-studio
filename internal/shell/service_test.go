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
	if len(info.Capabilities) != 3 {
		t.Fatalf("Capabilities length = %d, want 3", len(info.Capabilities))
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
