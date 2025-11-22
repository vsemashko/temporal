package audit

import (
	"context"
	"encoding/json"
	"time"

	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

// EventType represents the type of audit event
type EventType string

const (
	// EventTypeAuthZSuccess - authorization decision was allow
	EventTypeAuthZSuccess EventType = "authorization.success"
	// EventTypeAuthZFailure - authorization decision was deny
	EventTypeAuthZFailure EventType = "authorization.failure"
	// EventTypeAuthNSuccess - authentication succeeded
	EventTypeAuthNSuccess EventType = "authentication.success"
	// EventTypeAuthNFailure - authentication failed
	EventTypeAuthNFailure EventType = "authentication.failure"
	// EventTypeConfigChange - configuration was changed
	EventTypeConfigChange EventType = "config.change"
	// EventTypeCertRotation - certificate was rotated
	EventTypeCertRotation EventType = "certificate.rotation"
)

// Decision represents the authorization decision
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
)

// Event represents a security-relevant event that should be audited
type Event struct {
	// Timestamp when the event occurred (UTC)
	Timestamp time.Time `json:"timestamp"`

	// EventType categorizes the audit event
	EventType EventType `json:"event_type"`

	// UserID identifies the user making the request (from JWT subject or mTLS cert)
	UserID string `json:"user_id,omitempty"`

	// SourceIP is the IP address of the client making the request
	SourceIP string `json:"source_ip,omitempty"`

	// Namespace being targeted (if applicable)
	Namespace string `json:"namespace,omitempty"`

	// APIName is the full gRPC method name
	// Example: "/temporal.api.workflowservice.v1.WorkflowService/StartWorkflowExecution"
	APIName string `json:"api_name,omitempty"`

	// Decision is the authorization decision (allow/deny)
	Decision Decision `json:"decision,omitempty"`

	// Reason explains why the decision was made
	Reason string `json:"reason,omitempty"`

	// Metadata contains additional context-specific information
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// System role if applicable (system:admin, system:read, system:write)
	SystemRole string `json:"system_role,omitempty"`

	// Namespace role if applicable
	NamespaceRole string `json:"namespace_role,omitempty"`
}

// Logger is the interface for audit logging
type Logger interface {
	// LogAuthZ logs an authorization decision
	LogAuthZ(ctx context.Context, event *Event)

	// LogAuthN logs an authentication attempt
	LogAuthN(ctx context.Context, event *Event)

	// LogConfigChange logs a configuration change
	LogConfigChange(ctx context.Context, event *Event)

	// LogCertRotation logs a certificate rotation
	LogCertRotation(ctx context.Context, event *Event)
}

// Implementation of audit logger
type auditLogger struct {
	logger log.Logger
}

// NewLogger creates a new audit logger
func NewLogger(logger log.Logger) Logger {
	return &auditLogger{
		logger: logger,
	}
}

// LogAuthZ logs an authorization decision
func (a *auditLogger) LogAuthZ(ctx context.Context, event *Event) {
	a.logEvent(ctx, event)
}

// LogAuthN logs an authentication attempt
func (a *auditLogger) LogAuthN(ctx context.Context, event *Event) {
	a.logEvent(ctx, event)
}

// LogConfigChange logs a configuration change
func (a *auditLogger) LogConfigChange(ctx context.Context, event *Event) {
	a.logEvent(ctx, event)
}

// LogCertRotation logs a certificate rotation
func (a *auditLogger) LogCertRotation(ctx context.Context, event *Event) {
	a.logEvent(ctx, event)
}

// logEvent is the internal method that actually writes the audit log
func (a *auditLogger) logEvent(ctx context.Context, event *Event) {
	// Serialize event to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		a.logger.Error("Failed to marshal audit event",
			tag.Error(err),
			tag.NewStringTag("event_type", string(event.EventType)),
		)
		return
	}

	// Log at Info level for successful operations, Warn for denied operations
	tags := []tag.Tag{
		tag.NewStringTag("audit_event", "true"),
		tag.NewStringTag("event_type", string(event.EventType)),
		tag.NewStringTag("event_json", string(eventJSON)),
	}

	// Add optional fields as tags for better queryability
	if event.UserID != "" {
		tags = append(tags, tag.NewStringTag("user_id", event.UserID))
	}
	if event.SourceIP != "" {
		tags = append(tags, tag.NewStringTag("source_ip", event.SourceIP))
	}
	if event.Namespace != "" {
		tags = append(tags, tag.NewStringTag("namespace", event.Namespace))
	}
	if event.APIName != "" {
		tags = append(tags, tag.NewStringTag("api_name", event.APIName))
	}
	if event.Decision != "" {
		tags = append(tags, tag.NewStringTag("decision", string(event.Decision)))
	}

	// Log based on decision
	switch event.Decision {
	case DecisionDeny:
		// Denied requests are logged as warnings
		a.logger.Warn("Audit event", tags...)
	case DecisionAllow:
		// Allowed requests are logged as info
		a.logger.Info("Audit event", tags...)
	default:
		// Other events (config changes, cert rotations, etc.)
		a.logger.Info("Audit event", tags...)
	}
}

// NoopLogger is a no-op implementation of the audit logger
type NoopLogger struct{}

// NewNoopLogger creates a new no-op audit logger
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

// LogAuthZ does nothing
func (n *NoopLogger) LogAuthZ(ctx context.Context, event *Event) {}

// LogAuthN does nothing
func (n *NoopLogger) LogAuthN(ctx context.Context, event *Event) {}

// LogConfigChange does nothing
func (n *NoopLogger) LogConfigChange(ctx context.Context, event *Event) {}

// LogCertRotation does nothing
func (n *NoopLogger) LogCertRotation(ctx context.Context, event *Event) {}
