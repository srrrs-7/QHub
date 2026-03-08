package organization_test

import (
	"errors"
	"strings"
	"testing"

	"api/src/domain/apperror"
	"api/src/domain/organization"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// --- Organization (Aggregate) ---

func TestNewOrganization(t *testing.T) {
	id := organization.OrganizationIDFromUUID(uuid.New())
	name := organization.OrganizationName("Test Org")
	slug := organization.OrganizationSlug("test-org")
	plan := organization.PlanFree

	org := organization.NewOrganization(id, name, slug, plan)

	assert.Equal(t, id.String(), org.ID.String())
	assert.Equal(t, "Test Org", org.Name.String())
	assert.Equal(t, "test-org", org.Slug.String())
	assert.Equal(t, "free", org.Plan.String())
}

func TestNewOrganizationCmd(t *testing.T) {
	name := organization.OrganizationName("Cmd Org")
	slug := organization.OrganizationSlug("cmd-org")
	plan := organization.PlanPro

	cmd := organization.NewOrganizationCmd(name, slug, plan)

	assert.Equal(t, "Cmd Org", cmd.Name.String())
	assert.Equal(t, "cmd-org", cmd.Slug.String())
	assert.Equal(t, "pro", cmd.Plan.String())
}

// --- UserID ---

func TestNewUserID(t *testing.T) {
	validUUID := uuid.New().String()

	t.Run("valid UUID", func(t *testing.T) {
		result, err := organization.NewUserID(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, validUUID, result.String())
	})

	t.Run("invalid UUID", func(t *testing.T) {
		_, err := organization.NewUserID("invalid")
		assert.Error(t, err)
		var appErr apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.ValidationErrorName, appErr.ErrorName())
	})

	t.Run("UserIDFromUUID", func(t *testing.T) {
		u := uuid.New()
		result := organization.UserIDFromUUID(u)
		assert.Equal(t, u.String(), result.String())
		assert.Equal(t, u, result.UUID())
	})
}

func TestOrganizationIDHelpers(t *testing.T) {
	u := uuid.New()
	id := organization.OrganizationIDFromUUID(u)
	assert.Equal(t, u.String(), id.String())
	assert.Equal(t, u, id.UUID())
}

// --- OrganizationID ---

func TestNewOrganizationID(t *testing.T) {
	type args struct {
		id string
	}
	type expected struct {
		wantErr bool
		errName string
	}

	validUUID := uuid.New().String()

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{testName: "valid UUID", args: args{id: validUUID}, expected: expected{wantErr: false}},
		// 異常系
		{testName: "invalid UUID", args: args{id: "not-a-uuid"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 空文字
		{testName: "empty string", args: args{id: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 特殊文字
		{testName: "special chars", args: args{id: "<script>alert('xss')</script>"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := organization.NewOrganizationID(tt.args.id)

			if tt.expected.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.expected.errName, appErr.ErrorName())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.args.id, result.String())
			}
		})
	}
}

// --- OrganizationName ---

func TestNewOrganizationName(t *testing.T) {
	type args struct {
		name string
	}
	type expected struct {
		wantErr bool
		errName string
		value   string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{testName: "valid name", args: args{name: "My Organization"}, expected: expected{wantErr: false, value: "My Organization"}},
		{testName: "min length (2)", args: args{name: "AB"}, expected: expected{wantErr: false, value: "AB"}},
		// 異常系
		{testName: "too short (1 char)", args: args{name: "A"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "too long (101 chars)", args: args{name: strings.Repeat("a", 101)}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 境界値
		{testName: "exactly 100 chars", args: args{name: strings.Repeat("a", 100)}, expected: expected{wantErr: false, value: strings.Repeat("a", 100)}},
		// 特殊文字
		{testName: "Japanese name", args: args{name: "株式会社テスト"}, expected: expected{wantErr: false, value: "株式会社テスト"}},
		{testName: "emoji in name", args: args{name: "Team 🚀 Alpha"}, expected: expected{wantErr: false, value: "Team 🚀 Alpha"}},
		// 空文字
		{testName: "empty string", args: args{name: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "whitespace only", args: args{name: "   "}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := organization.NewOrganizationName(tt.args.name)

			if tt.expected.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.expected.errName, appErr.ErrorName())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.value, result.String())
			}
		})
	}
}

// --- OrganizationSlug ---

func TestNewOrganizationSlug(t *testing.T) {
	type args struct {
		slug string
	}
	type expected struct {
		wantErr bool
		errName string
		value   string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{testName: "valid slug", args: args{slug: "my-org"}, expected: expected{wantErr: false, value: "my-org"}},
		{testName: "slug with numbers", args: args{slug: "org-123"}, expected: expected{wantErr: false, value: "org-123"}},
		{testName: "min length (2)", args: args{slug: "ab"}, expected: expected{wantErr: false, value: "ab"}},
		// 異常系
		{testName: "too short (1 char)", args: args{slug: "a"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "too long (51 chars)", args: args{slug: strings.Repeat("a", 51)}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "uppercase letters", args: args{slug: "My-Org"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "starts with hyphen", args: args{slug: "-my-org"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "ends with hyphen", args: args{slug: "my-org-"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "contains spaces", args: args{slug: "my org"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "contains underscore", args: args{slug: "my_org"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 境界値
		{testName: "exactly 50 chars", args: args{slug: strings.Repeat("a", 50)}, expected: expected{wantErr: false, value: strings.Repeat("a", 50)}},
		// 特殊文字
		{testName: "Japanese chars", args: args{slug: "テスト"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "SQL injection", args: args{slug: "org'; DROP TABLE--"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 空文字
		{testName: "empty string", args: args{slug: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := organization.NewOrganizationSlug(tt.args.slug)

			if tt.expected.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.expected.errName, appErr.ErrorName())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.value, result.String())
			}
		})
	}
}

// --- Plan ---

func TestNewPlan(t *testing.T) {
	type args struct {
		plan string
	}
	type expected struct {
		wantErr bool
		errName string
		value   string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{testName: "free plan", args: args{plan: "free"}, expected: expected{wantErr: false, value: "free"}},
		{testName: "pro plan", args: args{plan: "pro"}, expected: expected{wantErr: false, value: "pro"}},
		{testName: "team plan", args: args{plan: "team"}, expected: expected{wantErr: false, value: "team"}},
		{testName: "enterprise plan", args: args{plan: "enterprise"}, expected: expected{wantErr: false, value: "enterprise"}},
		// 異常系
		{testName: "invalid plan", args: args{plan: "premium"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 空文字
		{testName: "empty string", args: args{plan: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 特殊文字
		{testName: "uppercase", args: args{plan: "FREE"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := organization.NewPlan(tt.args.plan)

			if tt.expected.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.expected.errName, appErr.ErrorName())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.value, result.String())
			}
		})
	}
}

// --- MemberRole ---

func TestNewMemberRole(t *testing.T) {
	type args struct {
		role string
	}
	type expected struct {
		wantErr bool
		errName string
		value   string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{testName: "owner", args: args{role: "owner"}, expected: expected{wantErr: false, value: "owner"}},
		{testName: "admin", args: args{role: "admin"}, expected: expected{wantErr: false, value: "admin"}},
		{testName: "member", args: args{role: "member"}, expected: expected{wantErr: false, value: "member"}},
		{testName: "viewer", args: args{role: "viewer"}, expected: expected{wantErr: false, value: "viewer"}},
		// 異常系
		{testName: "invalid role", args: args{role: "superadmin"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 空文字
		{testName: "empty string", args: args{role: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := organization.NewMemberRole(tt.args.role)

			if tt.expected.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.expected.errName, appErr.ErrorName())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.value, result.String())
			}
		})
	}
}
