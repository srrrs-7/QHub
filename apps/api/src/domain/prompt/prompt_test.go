package prompt_test

import (
	"errors"
	"strings"
	"testing"

	"api/src/domain/apperror"
	"api/src/domain/prompt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewPromptID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		id := uuid.New().String()
		result, err := prompt.NewPromptID(id)
		assert.NoError(t, err)
		assert.Equal(t, id, result.String())
	})
	t.Run("invalid", func(t *testing.T) {
		_, err := prompt.NewPromptID("bad")
		assert.Error(t, err)
	})
	t.Run("FromUUID", func(t *testing.T) {
		u := uuid.New()
		id := prompt.PromptIDFromUUID(u)
		assert.Equal(t, u, id.UUID())
	})
}

func TestNewPromptName(t *testing.T) {
	tests := []struct {
		testName string
		name     string
		wantErr  bool
	}{
		{testName: "valid", name: "Customer Support Bot", wantErr: false},
		{testName: "min (2)", name: "AB", wantErr: false},
		{testName: "max (200)", name: strings.Repeat("a", 200), wantErr: false},
		{testName: "too short", name: "A", wantErr: true},
		{testName: "too long", name: strings.Repeat("a", 201), wantErr: true},
		{testName: "Japanese", name: "カスタマーサポートボット", wantErr: false},
		{testName: "empty", name: "", wantErr: true},
		{testName: "whitespace", name: "   ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := prompt.NewPromptName(tt.name)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.name, result.String())
			}
		})
	}
}

func TestNewPromptSlug(t *testing.T) {
	tests := []struct {
		testName string
		slug     string
		wantErr  bool
	}{
		{testName: "valid", slug: "cs-bot", wantErr: false},
		{testName: "min (2)", slug: "ab", wantErr: false},
		{testName: "max (80)", slug: strings.Repeat("a", 80), wantErr: false},
		{testName: "too short", slug: "a", wantErr: true},
		{testName: "too long", slug: strings.Repeat("a", 81), wantErr: true},
		{testName: "uppercase", slug: "CS-Bot", wantErr: true},
		{testName: "empty", slug: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := prompt.NewPromptSlug(tt.slug)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.slug, result.String())
			}
		})
	}
}

func TestNewPromptType(t *testing.T) {
	tests := []struct {
		testName string
		pt       string
		wantErr  bool
	}{
		{testName: "system", pt: "system", wantErr: false},
		{testName: "user", pt: "user", wantErr: false},
		{testName: "combined", pt: "combined", wantErr: false},
		{testName: "invalid", pt: "other", wantErr: true},
		{testName: "empty", pt: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := prompt.NewPromptType(tt.pt)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.pt, result.String())
			}
		})
	}
}

func TestNewVersionStatus(t *testing.T) {
	tests := []struct {
		testName string
		status   string
		wantErr  bool
	}{
		{testName: "draft", status: "draft", wantErr: false},
		{testName: "review", status: "review", wantErr: false},
		{testName: "production", status: "production", wantErr: false},
		{testName: "archived", status: "archived", wantErr: false},
		{testName: "invalid", status: "active", wantErr: true},
		{testName: "empty", status: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := prompt.NewVersionStatus(tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.status, result.String())
			}
		})
	}
}

func TestVersionStatusTransition(t *testing.T) {
	tests := []struct {
		testName string
		from     string
		to       string
		wantErr  bool
	}{
		// Valid transitions
		{testName: "draft to review", from: "draft", to: "review", wantErr: false},
		{testName: "review to production", from: "review", to: "production", wantErr: false},
		{testName: "production to archived", from: "production", to: "archived", wantErr: false},
		{testName: "draft to archived", from: "draft", to: "archived", wantErr: false},
		// Invalid transitions
		{testName: "draft to production", from: "draft", to: "production", wantErr: true},
		{testName: "review to draft", from: "review", to: "draft", wantErr: true},
		{testName: "archived to draft", from: "archived", to: "draft", wantErr: true},
		{testName: "production to draft", from: "production", to: "draft", wantErr: true},
		{testName: "archived to production", from: "archived", to: "production", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			from := prompt.VersionStatus(tt.from)
			to := prompt.VersionStatus(tt.to)
			err := prompt.ValidateStatusTransition(from, to)
			if tt.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewPromptDescription(t *testing.T) {
	tests := []struct {
		testName string
		desc     string
		wantErr  bool
	}{
		{testName: "valid", desc: "A prompt for CS", wantErr: false},
		{testName: "empty allowed", desc: "", wantErr: false},
		{testName: "max (1000)", desc: strings.Repeat("a", 1000), wantErr: false},
		{testName: "too long", desc: strings.Repeat("a", 1001), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := prompt.NewPromptDescription(tt.desc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.desc, result.String())
			}
		})
	}
}

func TestNewChangeDescription(t *testing.T) {
	tests := []struct {
		testName string
		desc     string
		wantErr  bool
	}{
		{testName: "valid", desc: "Added safety constraints", wantErr: false},
		{testName: "empty allowed", desc: "", wantErr: false},
		{testName: "max (500)", desc: strings.Repeat("a", 500), wantErr: false},
		{testName: "too long", desc: strings.Repeat("a", 501), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := prompt.NewChangeDescription(tt.desc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.desc, result.String())
			}
		})
	}
}

func TestNewPrompt(t *testing.T) {
	id := prompt.PromptIDFromUUID(uuid.New())
	projID := uuid.New()
	p := prompt.NewPrompt(id, projID, "Bot", "bot", "system", "desc", 0, nil)
	assert.Equal(t, id.String(), p.ID.String())
	assert.Equal(t, projID, p.ProjectID)
	assert.Equal(t, "Bot", string(p.Name))
	assert.Equal(t, 0, p.LatestVersion)
	assert.Nil(t, p.ProductionVersion)
}
