package project_test

import (
	"errors"
	"strings"
	"testing"

	"api/src/domain/apperror"
	"api/src/domain/project"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewProjectID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		id := uuid.New().String()
		result, err := project.NewProjectID(id)
		assert.NoError(t, err)
		assert.Equal(t, id, result.String())
	})
	t.Run("invalid UUID", func(t *testing.T) {
		_, err := project.NewProjectID("invalid")
		assert.Error(t, err)
	})
	t.Run("empty", func(t *testing.T) {
		_, err := project.NewProjectID("")
		assert.Error(t, err)
	})
	t.Run("FromUUID helper", func(t *testing.T) {
		u := uuid.New()
		id := project.ProjectIDFromUUID(u)
		assert.Equal(t, u.String(), id.String())
		assert.Equal(t, u, id.UUID())
	})
}

func TestNewProjectName(t *testing.T) {
	tests := []struct {
		testName string
		name     string
		wantErr  bool
	}{
		// 正常系
		{testName: "valid name", name: "My Project", wantErr: false},
		{testName: "min length (2)", name: "AB", wantErr: false},
		{testName: "max length (100)", name: strings.Repeat("a", 100), wantErr: false},
		// 異常系
		{testName: "too short (1)", name: "A", wantErr: true},
		{testName: "too long (101)", name: strings.Repeat("a", 101), wantErr: true},
		// 特殊文字
		{testName: "Japanese", name: "プロジェクト", wantErr: false},
		{testName: "emoji", name: "Project 🚀", wantErr: false},
		// 空文字
		{testName: "empty", name: "", wantErr: true},
		{testName: "whitespace only", name: "   ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := project.NewProjectName(tt.name)
			if tt.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.name, result.String())
			}
		})
	}
}

func TestNewProjectSlug(t *testing.T) {
	tests := []struct {
		testName string
		slug     string
		wantErr  bool
	}{
		{testName: "valid", slug: "my-project", wantErr: false},
		{testName: "min (2)", slug: "ab", wantErr: false},
		{testName: "max (50)", slug: strings.Repeat("a", 50), wantErr: false},
		{testName: "too short", slug: "a", wantErr: true},
		{testName: "too long", slug: strings.Repeat("a", 51), wantErr: true},
		{testName: "uppercase", slug: "My-Project", wantErr: true},
		{testName: "starts with hyphen", slug: "-proj", wantErr: true},
		{testName: "ends with hyphen", slug: "proj-", wantErr: true},
		{testName: "empty", slug: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := project.NewProjectSlug(tt.slug)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.slug, result.String())
			}
		})
	}
}

func TestNewProjectDescription(t *testing.T) {
	tests := []struct {
		testName string
		desc     string
		wantErr  bool
	}{
		{testName: "valid", desc: "A great project", wantErr: false},
		{testName: "empty allowed", desc: "", wantErr: false},
		{testName: "max (500)", desc: strings.Repeat("a", 500), wantErr: false},
		{testName: "too long", desc: strings.Repeat("a", 501), wantErr: true},
		{testName: "Japanese", desc: "説明テキスト", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := project.NewProjectDescription(tt.desc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.desc, result.String())
			}
		})
	}
}

func TestNewProject(t *testing.T) {
	id := project.ProjectIDFromUUID(uuid.New())
	orgID := uuid.New()
	name := project.ProjectName("Test")
	slug := project.ProjectSlug("test")
	desc := project.ProjectDescription("desc")

	p := project.NewProject(id, orgID, name, slug, desc)
	assert.Equal(t, id.String(), p.ID.String())
	assert.Equal(t, orgID, p.OrganizationID)
	assert.Equal(t, "Test", p.Name.String())
	assert.Equal(t, "test", p.Slug.String())
	assert.Equal(t, "desc", p.Description.String())
}
