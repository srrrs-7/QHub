package tasks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetRequestValidate(t *testing.T) {
	type args struct {
		id string
	}
	type expected struct {
		hasError bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		{
			testName: "valid uuid",
			args: args{
				id: "550e8400-e29b-41d4-a716-446655440000",
			},
			expected: expected{
				hasError: false,
			},
		},
		{
			testName: "invalid uuid",
			args: args{
				id: "invalid-uuid",
			},
			expected: expected{
				hasError: true,
			},
		},
		{
			testName: "empty id",
			args: args{
				id: "",
			},
			expected: expected{
				hasError: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := getRequest{ID: tt.args.id}
			_, err := req.validate()

			gotHasError := err != nil
			if diff := cmp.Diff(tt.expected.hasError, gotHasError); diff != "" {
				t.Errorf("validation result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListRequestValidate(t *testing.T) {
	type args struct {
		id          string
		title       string
		description string
		status      string
	}
	type expected struct {
		hasError bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		{
			testName: "valid empty request",
			args:     args{},
			expected: expected{hasError: false},
		},
		{
			testName: "valid with all fields",
			args:     args{id: "550e8400-e29b-41d4-a716-446655440000", title: "Valid", status: "pending"},
			expected: expected{hasError: false},
		},
		{
			testName: "invalid uuid",
			args:     args{id: "not-a-uuid"},
			expected: expected{hasError: true},
		},
		{
			testName: "title too short",
			args:     args{title: "ab"},
			expected: expected{hasError: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := listRequest{
				ID:          tt.args.id,
				Title:       tt.args.title,
				Description: tt.args.description,
				Status:      tt.args.status,
			}
			_, err := req.validate()

			gotHasError := err != nil
			if diff := cmp.Diff(tt.expected.hasError, gotHasError); diff != "" {
				t.Errorf("validation result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPutRequestValidate(t *testing.T) {
	type args struct {
		id          string
		title       string
		description string
		status      string
	}
	type expected struct {
		hasError bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		{
			testName: "valid request",
			args:     args{id: "550e8400-e29b-41d4-a716-446655440000", title: "Valid Title", status: "pending"},
			expected: expected{hasError: false},
		},
		{
			testName: "invalid uuid",
			args:     args{id: "not-a-uuid", title: "Valid Title", status: "pending"},
			expected: expected{hasError: true},
		},
		{
			testName: "empty title",
			args:     args{id: "550e8400-e29b-41d4-a716-446655440000", title: "", status: "pending"},
			expected: expected{hasError: true},
		},
		{
			testName: "title too short",
			args:     args{id: "550e8400-e29b-41d4-a716-446655440000", title: "ab", status: "pending"},
			expected: expected{hasError: true},
		},
		{
			testName: "invalid status",
			args:     args{id: "550e8400-e29b-41d4-a716-446655440000", title: "Valid Title", status: "invalid"},
			expected: expected{hasError: true},
		},
		{
			testName: "status omitted is valid",
			args:     args{id: "550e8400-e29b-41d4-a716-446655440000", title: "Valid Title", status: ""},
			expected: expected{hasError: false},
		},
		{
			testName: "HTML tags stripped leaving short title",
			args:     args{id: "550e8400-e29b-41d4-a716-446655440000", title: "<b>ab</b>", status: "pending"},
			expected: expected{hasError: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := putRequest{
				ID:          tt.args.id,
				Title:       tt.args.title,
				Description: tt.args.description,
				Status:      tt.args.status,
			}
			_, err := req.validate()

			gotHasError := err != nil
			if diff := cmp.Diff(tt.expected.hasError, gotHasError); diff != "" {
				t.Errorf("validation result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPostRequestValidate(t *testing.T) {
	type args struct {
		title       string
		description string
	}
	type expected struct {
		hasError bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		{
			testName: "valid request",
			args: args{
				title:       "Valid Task Title",
				description: "Valid description",
			},
			expected: expected{
				hasError: false,
			},
		},
		{
			testName: "title too short",
			args: args{
				title:       "ab",
				description: "Valid description",
			},
			expected: expected{
				hasError: true,
			},
		},
		{
			testName: "empty title",
			args: args{
				title:       "",
				description: "Valid description",
			},
			expected: expected{
				hasError: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := postRequest{
				Title:       tt.args.title,
				Description: tt.args.description,
			}
			_, err := req.validate()

			gotHasError := err != nil
			if diff := cmp.Diff(tt.expected.hasError, gotHasError); diff != "" {
				t.Errorf("validation result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
