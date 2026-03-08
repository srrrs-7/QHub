package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnavailableError(t *testing.T) {
	err := &unavailableError{}

	tests := []struct {
		testName string
		method   string
		expected string
	}{
		{testName: "Error()", method: "error", expected: "embedding service not available"},
		{testName: "ErrorName()", method: "errorName", expected: "ServiceUnavailableError"},
		{testName: "DomainName()", method: "domainName", expected: "Search"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var got string
			switch tt.method {
			case "error":
				got = err.Error()
			case "errorName":
				got = err.ErrorName()
			case "domainName":
				got = err.DomainName()
			}
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		testName string
		msg      string
	}{
		{testName: "custom message", msg: "query is required"},
		{testName: "invalid org_id", msg: "invalid org_id"},
		{testName: "empty message", msg: ""},
		// 特殊文字
		{testName: "unicode message", msg: "クエリが必要です"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := &validationError{msg: tt.msg}
			if diff := cmp.Diff(tt.msg, err.Error()); diff != "" {
				t.Errorf("Error() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff("ValidationError", err.ErrorName()); diff != "" {
				t.Errorf("ErrorName() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff("Search", err.DomainName()); diff != "" {
				t.Errorf("DomainName() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewSearchHandler(t *testing.T) {
	h := NewSearchHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.embSvc != nil {
		t.Error("expected nil embSvc")
	}
	if h.q != nil {
		t.Error("expected nil querier")
	}
}
