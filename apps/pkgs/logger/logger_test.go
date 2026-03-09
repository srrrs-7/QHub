package logger

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewRequestLog(t *testing.T) {
	// 正常系 (Happy Path)
	t.Run("attaches fresh RequestLog to context", func(t *testing.T) {
		ctx, rl := NewRequestLog(context.Background())
		if rl == nil {
			t.Fatal("expected non-nil RequestLog")
		}
		// The same pointer should be retrievable from the context.
		got := RequestLogFrom(ctx)
		if got != rl {
			t.Errorf("expected same pointer: got %p, want %p", got, rl)
		}
	})

	// 正常系 — mutation is visible through context
	t.Run("mutations to returned pointer are visible via RequestLogFrom", func(t *testing.T) {
		ctx, rl := NewRequestLog(context.Background())
		rl.UserID = "user-123"
		rl.OrgID = "org-456"
		rl.AuthMethod = "bearer"
		rl.RequestID = "req-789"

		got := RequestLogFrom(ctx)
		if diff := cmp.Diff("user-123", got.UserID); diff != "" {
			t.Errorf("UserID mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("org-456", got.OrgID); diff != "" {
			t.Errorf("OrgID mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("bearer", got.AuthMethod); diff != "" {
			t.Errorf("AuthMethod mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("req-789", got.RequestID); diff != "" {
			t.Errorf("RequestID mismatch (-want +got):\n%s", diff)
		}
	})

	// 境界値 — zero fields
	t.Run("new RequestLog has empty fields", func(t *testing.T) {
		_, rl := NewRequestLog(context.Background())
		if diff := cmp.Diff("", rl.UserID); diff != "" {
			t.Errorf("UserID should be empty (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("", rl.AuthMethod); diff != "" {
			t.Errorf("AuthMethod should be empty (-want +got):\n%s", diff)
		}
	})
}

func TestRequestLogFrom(t *testing.T) {
	type expected struct {
		userID     string
		orgID      string
		authMethod string
	}

	tests := []struct {
		testName string
		ctx      context.Context
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "returns RequestLog when present in context",
			ctx: func() context.Context {
				ctx, rl := NewRequestLog(context.Background())
				rl.UserID = "u1"
				rl.OrgID = "o1"
				rl.AuthMethod = "apikey"
				return ctx
			}(),
			expected: expected{"u1", "o1", "apikey"},
		},

		// Null/Nil — context without RequestLog returns safe zero value
		{
			testName: "returns zero-value RequestLog when not in context",
			ctx:      context.Background(),
			expected: expected{"", "", ""},
		},

		// 異常系 — wrong type in context key
		{
			testName: "returns zero-value RequestLog when context has wrong type",
			ctx:      context.WithValue(context.Background(), requestLogKey{}, "not-a-pointer"),
			expected: expected{"", "", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := RequestLogFrom(tt.ctx)
			if got == nil {
				t.Fatal("RequestLogFrom must never return nil")
			}
			if diff := cmp.Diff(tt.expected.userID, got.UserID); diff != "" {
				t.Errorf("UserID mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.orgID, got.OrgID); diff != "" {
				t.Errorf("OrgID mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.authMethod, got.AuthMethod); diff != "" {
				t.Errorf("AuthMethod mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRequestLog_SharedPointer(t *testing.T) {
	// Verify that the pointer-in-context pattern works across derived contexts
	// (simulating the middleware chain where each layer creates r.WithContext(ctx)).
	ctx, rl := NewRequestLog(context.Background())

	// Simulate downstream middleware adding a value and re-wrapping context.
	child := context.WithValue(ctx, struct{ k string }{"key"}, "value")

	// Mutation on the original pointer.
	rl.UserID = "propagated-user"

	// The mutation should be visible even from the child context.
	got := RequestLogFrom(child)
	if diff := cmp.Diff("propagated-user", got.UserID); diff != "" {
		t.Errorf("UserID not visible in child context (-want +got):\n%s", diff)
	}
}
