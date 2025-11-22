package audit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

// mockLogger captures logged messages for testing
type mockLogger struct {
	infoMessages  []logMessage
	warnMessages  []logMessage
	errorMessages []logMessage
}

type logMessage struct {
	msg  string
	tags []tag.Tag
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		infoMessages:  []logMessage{},
		warnMessages:  []logMessage{},
		errorMessages: []logMessage{},
	}
}

func (m *mockLogger) Debug(msg string, tags ...tag.Tag) {}
func (m *mockLogger) Info(msg string, tags ...tag.Tag) {
	m.infoMessages = append(m.infoMessages, logMessage{msg: msg, tags: tags})
}
func (m *mockLogger) Warn(msg string, tags ...tag.Tag) {
	m.warnMessages = append(m.warnMessages, logMessage{msg: msg, tags: tags})
}
func (m *mockLogger) Error(msg string, tags ...tag.Tag) {
	m.errorMessages = append(m.errorMessages, logMessage{msg: msg, tags: tags})
}
func (m *mockLogger) DPanic(msg string, tags ...tag.Tag) {}
func (m *mockLogger) Panic(msg string, tags ...tag.Tag)  {}
func (m *mockLogger) Fatal(msg string, tags ...tag.Tag)  {}

var _ log.Logger = (*mockLogger)(nil)

func TestAuditLogger_LogAuthZAllow(t *testing.T) {
	mockLog := newMockLogger()
	auditLog := NewLogger(mockLog)

	event := &Event{
		Timestamp:     time.Now().UTC(),
		EventType:     EventTypeAuthZSuccess,
		UserID:        "user@example.com",
		SourceIP:      "192.168.1.100",
		Namespace:     "test-namespace",
		APIName:       "/temporal.api.workflowservice.v1.WorkflowService/StartWorkflowExecution",
		Decision:      DecisionAllow,
		Reason:        "User has namespace:write permission",
		SystemRole:    "",
		NamespaceRole: "write",
		Metadata: map[string]interface{}{
			"workflow_id": "test-workflow",
		},
	}

	auditLog.LogAuthZ(context.Background(), event)

	// Should log at Info level for allow decisions
	assert.Equal(t, 1, len(mockLog.infoMessages))
	assert.Equal(t, 0, len(mockLog.warnMessages))

	// Verify the message contains audit marker
	msg := mockLog.infoMessages[0]
	assert.Equal(t, "Audit event", msg.msg)

	// Verify tags
	tags := msg.tags
	assert.True(t, containsTag(tags, "audit_event", "true"))
	assert.True(t, containsTag(tags, "event_type", string(EventTypeAuthZSuccess)))
	assert.True(t, containsTag(tags, "user_id", "user@example.com"))
	assert.True(t, containsTag(tags, "source_ip", "192.168.1.100"))
	assert.True(t, containsTag(tags, "namespace", "test-namespace"))
	assert.True(t, containsTag(tags, "decision", string(DecisionAllow)))

	// Verify event_json tag contains valid JSON
	eventJSONTag := findTag(tags, "event_json")
	require.NotNil(t, eventJSONTag)

	var unmarshaled Event
	err := json.Unmarshal([]byte(eventJSONTag.Value().(string)), &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, event.UserID, unmarshaled.UserID)
	assert.Equal(t, event.SourceIP, unmarshaled.SourceIP)
	assert.Equal(t, event.Namespace, unmarshaled.Namespace)
	assert.Equal(t, event.APIName, unmarshaled.APIName)
	assert.Equal(t, event.Decision, unmarshaled.Decision)
}

func TestAuditLogger_LogAuthZDeny(t *testing.T) {
	mockLog := newMockLogger()
	auditLog := NewLogger(mockLog)

	event := &Event{
		Timestamp:  time.Now().UTC(),
		EventType:  EventTypeAuthZFailure,
		UserID:     "user@example.com",
		SourceIP:   "192.168.1.100",
		Namespace:  "test-namespace",
		APIName:    "/temporal.api.workflowservice.v1.WorkflowService/TerminateWorkflowExecution",
		Decision:   DecisionDeny,
		Reason:     "User does not have namespace:admin permission",
		SystemRole: "",
		Metadata:   map[string]interface{}{},
	}

	auditLog.LogAuthZ(context.Background(), event)

	// Should log at Warn level for deny decisions
	assert.Equal(t, 0, len(mockLog.infoMessages))
	assert.Equal(t, 1, len(mockLog.warnMessages))

	// Verify the message
	msg := mockLog.warnMessages[0]
	assert.Equal(t, "Audit event", msg.msg)

	// Verify tags
	tags := msg.tags
	assert.True(t, containsTag(tags, "audit_event", "true"))
	assert.True(t, containsTag(tags, "event_type", string(EventTypeAuthZFailure)))
	assert.True(t, containsTag(tags, "decision", string(DecisionDeny)))
}

func TestAuditLogger_LogAuthNSuccess(t *testing.T) {
	mockLog := newMockLogger()
	auditLog := NewLogger(mockLog)

	event := &Event{
		Timestamp:  time.Now().UTC(),
		EventType:  EventTypeAuthNSuccess,
		UserID:     "user@example.com",
		SourceIP:   "192.168.1.100",
		SystemRole: "system:admin",
		Metadata: map[string]interface{}{
			"auth_method": "jwt",
		},
	}

	auditLog.LogAuthN(context.Background(), event)

	assert.Equal(t, 1, len(mockLog.infoMessages))
	assert.Equal(t, 0, len(mockLog.warnMessages))
}

func TestAuditLogger_LogAuthNFailure(t *testing.T) {
	mockLog := newMockLogger()
	auditLog := NewLogger(mockLog)

	event := &Event{
		Timestamp: time.Now().UTC(),
		EventType: EventTypeAuthNFailure,
		SourceIP:  "192.168.1.100",
		Reason:    "Invalid JWT signature",
		Metadata: map[string]interface{}{
			"auth_method": "jwt",
			"error":       "signature verification failed",
		},
	}

	auditLog.LogAuthN(context.Background(), event)

	// Authentication failures should be logged as info
	assert.Equal(t, 1, len(mockLog.infoMessages))
}

func TestAuditLogger_LogConfigChange(t *testing.T) {
	mockLog := newMockLogger()
	auditLog := NewLogger(mockLog)

	event := &Event{
		Timestamp: time.Now().UTC(),
		EventType: EventTypeConfigChange,
		UserID:    "admin@example.com",
		Reason:    "Updated TLS configuration",
		Metadata: map[string]interface{}{
			"config_key": "tls.minVersion",
			"old_value":  "1.2",
			"new_value":  "1.3",
		},
	}

	auditLog.LogConfigChange(context.Background(), event)

	assert.Equal(t, 1, len(mockLog.infoMessages))
}

func TestAuditLogger_LogCertRotation(t *testing.T) {
	mockLog := newMockLogger()
	auditLog := NewLogger(mockLog)

	event := &Event{
		Timestamp: time.Now().UTC(),
		EventType: EventTypeCertRotation,
		UserID:    "admin@example.com",
		Reason:    "Scheduled certificate rotation",
		Metadata: map[string]interface{}{
			"cert_subject":    "CN=temporal.example.com",
			"cert_expiry_old": "2025-12-31",
			"cert_expiry_new": "2026-12-31",
			"rotation_type":   "planned",
		},
	}

	auditLog.LogCertRotation(context.Background(), event)

	assert.Equal(t, 1, len(mockLog.infoMessages))
}

func TestNoopLogger(t *testing.T) {
	// NoopLogger should not panic or cause errors
	noop := NewNoopLogger()

	event := &Event{
		Timestamp: time.Now().UTC(),
		EventType: EventTypeAuthZSuccess,
		UserID:    "user@example.com",
		Decision:  DecisionAllow,
	}

	// Should not panic
	noop.LogAuthZ(context.Background(), event)
	noop.LogAuthN(context.Background(), event)
	noop.LogConfigChange(context.Background(), event)
	noop.LogCertRotation(context.Background(), event)
}

// Helper functions

func containsTag(tags []tag.Tag, key, value string) bool {
	for _, t := range tags {
		if t.Key() == key && t.Value() == value {
			return true
		}
	}
	return false
}

func findTag(tags []tag.Tag, key string) tag.Tag {
	for _, t := range tags {
		if t.Key() == key {
			return t
		}
	}
	return nil
}
