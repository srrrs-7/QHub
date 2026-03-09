package industry_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"api/src/domain/apperror"
	"api/src/domain/industry"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// --- IndustryID ---

func TestNewIndustryID(t *testing.T) {
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
		{testName: "Japanese chars", args: args{id: "テスト"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := industry.NewIndustryID(tt.args.id)

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

func TestIndustryIDFromUUID(t *testing.T) {
	u := uuid.New()
	id := industry.IndustryIDFromUUID(u)
	assert.Equal(t, u.String(), id.String())
	assert.Equal(t, u, id.UUID())
}

// --- IndustrySlug ---

func TestNewIndustrySlug(t *testing.T) {
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
		{testName: "valid slug", args: args{slug: "healthcare"}, expected: expected{wantErr: false, value: "healthcare"}},
		{testName: "slug with hyphens", args: args{slug: "customer-support"}, expected: expected{wantErr: false, value: "customer-support"}},
		{testName: "slug with numbers", args: args{slug: "web3"}, expected: expected{wantErr: false, value: "web3"}},
		{testName: "min length (2)", args: args{slug: "ab"}, expected: expected{wantErr: false, value: "ab"}},
		// 異常系
		{testName: "too short (1 char)", args: args{slug: "a"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "too long (51 chars)", args: args{slug: strings.Repeat("a", 51)}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "uppercase letters", args: args{slug: "Healthcare"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "starts with hyphen", args: args{slug: "-healthcare"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "ends with hyphen", args: args{slug: "healthcare-"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "contains spaces", args: args{slug: "health care"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "contains underscore", args: args{slug: "health_care"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 境界値
		{testName: "exactly 50 chars", args: args{slug: strings.Repeat("a", 50)}, expected: expected{wantErr: false, value: strings.Repeat("a", 50)}},
		// 特殊文字
		{testName: "Japanese chars", args: args{slug: "テスト"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "SQL injection", args: args{slug: "slug'; DROP TABLE--"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "emoji", args: args{slug: "health🏥"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 空文字
		{testName: "empty string", args: args{slug: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "whitespace only", args: args{slug: "   "}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := industry.NewIndustrySlug(tt.args.slug)

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

// --- IndustryName ---

func TestNewIndustryName(t *testing.T) {
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
		{testName: "valid name", args: args{name: "Healthcare"}, expected: expected{wantErr: false, value: "Healthcare"}},
		{testName: "min length (1)", args: args{name: "A"}, expected: expected{wantErr: false, value: "A"}},
		// 異常系
		{testName: "too long (101 chars)", args: args{name: strings.Repeat("a", 101)}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 境界値
		{testName: "exactly 100 chars", args: args{name: strings.Repeat("a", 100)}, expected: expected{wantErr: false, value: strings.Repeat("a", 100)}},
		// 特殊文字
		{testName: "Japanese name", args: args{name: "医療"}, expected: expected{wantErr: false, value: "医療"}},
		{testName: "emoji in name", args: args{name: "Health 🏥"}, expected: expected{wantErr: false, value: "Health 🏥"}},
		{testName: "SQL injection", args: args{name: "'; DROP TABLE--"}, expected: expected{wantErr: false, value: "'; DROP TABLE--"}},
		// 空文字
		{testName: "empty string", args: args{name: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "whitespace only", args: args{name: "   "}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := industry.NewIndustryName(tt.args.name)

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

// --- IndustryDescription ---

func TestNewIndustryDescription(t *testing.T) {
	type args struct {
		desc string
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
		{testName: "valid description", args: args{desc: "Healthcare and medical industry prompts"}, expected: expected{wantErr: false, value: "Healthcare and medical industry prompts"}},
		// 境界値 - empty is valid (optional field)
		{testName: "empty string is valid", args: args{desc: ""}, expected: expected{wantErr: false, value: ""}},
		// 境界値
		{testName: "exactly 1000 chars", args: args{desc: strings.Repeat("a", 1000)}, expected: expected{wantErr: false, value: strings.Repeat("a", 1000)}},
		// 異常系
		{testName: "too long (1001 chars)", args: args{desc: strings.Repeat("a", 1001)}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 特殊文字
		{testName: "Japanese description", args: args{desc: "医療業界のプロンプト管理"}, expected: expected{wantErr: false, value: "医療業界のプロンプト管理"}},
		{testName: "emoji description", args: args{desc: "Health prompts 🏥💊"}, expected: expected{wantErr: false, value: "Health prompts 🏥💊"}},
		{testName: "SQL injection", args: args{desc: "'; DROP TABLE industry_configs; --"}, expected: expected{wantErr: false, value: "'; DROP TABLE industry_configs; --"}},
		// 空文字 - whitespace only is valid (no trimming for descriptions)
		{testName: "whitespace only", args: args{desc: "   "}, expected: expected{wantErr: false, value: "   "}},
		// Null/Nil - zero value
		{testName: "single char", args: args{desc: "x"}, expected: expected{wantErr: false, value: "x"}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := industry.NewIndustryDescription(tt.args.desc)

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

// --- IndustryConfig (Entity) ---

func TestNewIndustryConfig(t *testing.T) {
	id := industry.IndustryIDFromUUID(uuid.New())
	slug := industry.IndustrySlug("healthcare")
	name := industry.IndustryName("Healthcare")
	desc := industry.IndustryDescription("Medical prompts")

	cfg := industry.NewIndustryConfig(id, slug, name, desc, nil, nil, time.Now(), time.Now())

	assert.Equal(t, id.String(), cfg.ID.String())
	assert.Equal(t, "healthcare", cfg.Slug.String())
	assert.Equal(t, "Healthcare", cfg.Name.String())
	assert.Equal(t, "Medical prompts", cfg.Description.String())
}
