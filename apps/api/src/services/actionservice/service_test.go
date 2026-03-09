package actionservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"api/src/domain/apperror"
	"api/src/domain/prompt"
	"api/src/services/intentservice"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockPromptRepo struct {
	findByIDFn            func(ctx context.Context, id prompt.PromptID) (prompt.Prompt, error)
	updateLatestVersionFn func(ctx context.Context, id prompt.PromptID, version int) (prompt.Prompt, error)
}

func (m *mockPromptRepo) FindByID(ctx context.Context, id prompt.PromptID) (prompt.Prompt, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockPromptRepo) FindByProjectAndSlug(_ context.Context, _ uuid.UUID, _ prompt.PromptSlug) (prompt.Prompt, error) {
	return prompt.Prompt{}, nil
}
func (m *mockPromptRepo) FindAllByProject(_ context.Context, _ uuid.UUID) ([]prompt.Prompt, error) {
	return nil, nil
}
func (m *mockPromptRepo) Create(_ context.Context, _ prompt.PromptCmd) (prompt.Prompt, error) {
	return prompt.Prompt{}, nil
}
func (m *mockPromptRepo) Update(_ context.Context, _ prompt.PromptID, _ prompt.PromptCmd) (prompt.Prompt, error) {
	return prompt.Prompt{}, nil
}
func (m *mockPromptRepo) UpdateLatestVersion(ctx context.Context, id prompt.PromptID, version int) (prompt.Prompt, error) {
	return m.updateLatestVersionFn(ctx, id, version)
}
func (m *mockPromptRepo) UpdateProductionVersion(_ context.Context, _ prompt.PromptID, _ *int) (prompt.Prompt, error) {
	return prompt.Prompt{}, nil
}

type mockVersionRepo struct {
	createFn func(ctx context.Context, cmd prompt.VersionCmd, versionNumber int) (prompt.PromptVersion, error)
}

func (m *mockVersionRepo) FindByPromptAndNumber(_ context.Context, _ prompt.PromptID, _ int) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}
func (m *mockVersionRepo) FindAllByPrompt(_ context.Context, _ prompt.PromptID) ([]prompt.PromptVersion, error) {
	return nil, nil
}
func (m *mockVersionRepo) FindLatest(_ context.Context, _ prompt.PromptID) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}
func (m *mockVersionRepo) FindProduction(_ context.Context, _ prompt.PromptID) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}
func (m *mockVersionRepo) Create(ctx context.Context, cmd prompt.VersionCmd, versionNumber int) (prompt.PromptVersion, error) {
	return m.createFn(ctx, cmd, versionNumber)
}
func (m *mockVersionRepo) UpdateStatus(_ context.Context, _ prompt.PromptVersionID, _ prompt.VersionStatus) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}
func (m *mockVersionRepo) ArchiveProduction(_ context.Context, _ prompt.PromptID) error { return nil }
func (m *mockVersionRepo) UpdateLintResult(_ context.Context, _ prompt.PromptVersionID, _ json.RawMessage) error {
	return nil
}
func (m *mockVersionRepo) UpdateSemanticDiff(_ context.Context, _ prompt.PromptVersionID, _ json.RawMessage) error {
	return nil
}
func (m *mockVersionRepo) UpdateEmbedding(_ context.Context, _ prompt.PromptVersionID, _ []float32) error {
	return nil
}

// ---------------------------------------------------------------------------
// TestExtractActions
// ---------------------------------------------------------------------------

func TestExtractActions(t *testing.T) {
	promptID := uuid.New()

	type args struct {
		intent       intentservice.Intent
		promptID     uuid.UUID
		responseText string
	}
	type expected struct {
		count       int
		actionTypes []ActionType
		hasContent  bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "create intent with single code block extracts one action",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID: promptID,
				responseText: "Here is a prompt:\n```\nYou are a helpful assistant.\n```\n",
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},
		{
			testName: "improve intent with code block extracts action",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentImprove, Confidence: 0.8},
				promptID: promptID,
				responseText: "Improved version:\n```\nYou are a precise assistant.\n```\n",
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},
		{
			testName: "multiple code blocks extracts multiple actions",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID: promptID,
				responseText: "Option A:\n```\nFirst prompt.\n```\nOption B:\n```\nSecond prompt.\n```\n",
			},
			expected: expected{count: 2, actionTypes: []ActionType{ActionCreateVersion, ActionCreateVersion}, hasContent: true},
		},
		{
			testName: "code block with language tag is extracted",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID: promptID,
				responseText: "```markdown\nYou are an assistant.\n```\n",
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},

		// 異常系 (Error Cases)
		{
			testName: "general intent does not extract actions",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentGeneral, Confidence: 0.5},
				promptID:     promptID,
				responseText: "Here is a prompt:\n```\nContent.\n```\n",
			},
			expected: expected{count: 0},
		},
		{
			testName: "explain intent does not extract actions",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentExplain, Confidence: 0.7},
				promptID:     promptID,
				responseText: "```\nExample code.\n```\n",
			},
			expected: expected{count: 0},
		},
		{
			testName: "compare intent does not extract actions",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentCompare, Confidence: 0.8},
				promptID:     promptID,
				responseText: "```\nSome diff.\n```\n",
			},
			expected: expected{count: 0},
		},
		{
			testName: "compliance intent does not extract actions",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentCompliance, Confidence: 0.8},
				promptID:     promptID,
				responseText: "```\nCompliance text.\n```\n",
			},
			expected: expected{count: 0},
		},
		{
			testName: "best_practice intent does not extract actions",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentBestPractice, Confidence: 0.7},
				promptID:     promptID,
				responseText: "```\nBest practice.\n```\n",
			},
			expected: expected{count: 0},
		},
		{
			testName: "no code blocks returns nil",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID:     promptID,
				responseText: "Just a plain text response with no code blocks.",
			},
			expected: expected{count: 0},
		},
		{
			testName: "unclosed code block returns nil",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID:     promptID,
				responseText: "Here:\n```\nUnclosed block without end",
			},
			expected: expected{count: 0},
		},

		// 境界値 (Boundary Values)
		{
			testName: "empty code block is skipped",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID:     promptID,
				responseText: "```\n\n```\n",
			},
			expected: expected{count: 0},
		},
		{
			testName: "whitespace-only code block is skipped",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID:     promptID,
				responseText: "```\n   \n   \n```\n",
			},
			expected: expected{count: 0},
		},
		{
			testName: "single character code block is extracted",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID: promptID,
				responseText: "```\na\n```\n",
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},
		{
			testName: "very long code block content is extracted",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID: promptID,
				responseText: fmt.Sprintf("```\n%s\n```\n", strings.Repeat("A long line of prompt text. ", 200)),
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},

		// 特殊文字 (Special Characters)
		{
			testName: "code block with emoji content",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID: promptID,
				responseText: "```\nYou are a helpful assistant 🤖\n```\n",
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},
		{
			testName: "code block with Japanese content",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID: promptID,
				responseText: "```\nあなたは親切なアシスタントです。\n```\n",
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},
		{
			testName: "code block with SQL injection attempt",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID: promptID,
				responseText: "```\nSELECT * FROM users; DROP TABLE users;--\n```\n",
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},
		{
			testName: "code block with special JSON characters",
			args: args{
				intent:   intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID: promptID,
				responseText: "```\n{\"key\": \"value with \\\"quotes\\\"\"}\n```\n",
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},

		// 空文字 (Empty/Whitespace)
		{
			testName: "empty response text returns nil",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID:     promptID,
				responseText: "",
			},
			expected: expected{count: 0},
		},
		{
			testName: "whitespace-only response text returns nil",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID:     promptID,
				responseText: "   \n\t  \n  ",
			},
			expected: expected{count: 0},
		},

		// Null/Nil (zero values)
		{
			testName: "nil UUID prompt ID still extracts actions",
			args: args{
				intent:       intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8},
				promptID:     uuid.Nil,
				responseText: "```\nContent.\n```\n",
			},
			expected: expected{count: 1, actionTypes: []ActionType{ActionCreateVersion}, hasContent: true},
		},
		{
			testName: "zero-value intent returns nil",
			args: args{
				intent:       intentservice.Intent{},
				promptID:     promptID,
				responseText: "```\nContent.\n```\n",
			},
			expected: expected{count: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := ExtractActions(tt.args.intent, tt.args.promptID, tt.args.responseText)

			if diff := cmp.Diff(tt.expected.count, len(got)); diff != "" {
				t.Errorf("action count mismatch (-want +got):\n%s", diff)
				return
			}

			if tt.expected.count == 0 {
				return
			}

			for i, a := range got {
				if diff := cmp.Diff(string(tt.expected.actionTypes[i]), string(a.Type)); diff != "" {
					t.Errorf("action[%d] type mismatch (-want +got):\n%s", i, diff)
				}
				if diff := cmp.Diff(tt.args.promptID, a.PromptID); diff != "" {
					t.Errorf("action[%d] promptID mismatch (-want +got):\n%s", i, diff)
				}
				if tt.expected.hasContent && len(a.SuggestedContent) == 0 {
					t.Errorf("action[%d] expected non-empty SuggestedContent", i)
				}
				if a.Description == "" {
					t.Errorf("action[%d] expected non-empty Description", i)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestExtractActions_MultipleBlockDescriptions
// ---------------------------------------------------------------------------

func TestExtractActions_MultipleBlockDescriptions(t *testing.T) {
	promptID := uuid.New()
	intent := intentservice.Intent{Type: intentservice.IntentCreate, Confidence: 0.8}
	responseText := "A:\n```\nFirst.\n```\nB:\n```\nSecond.\n```\n"

	actions := ExtractActions(intent, promptID, responseText)
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}

	if !strings.Contains(actions[0].Description, "variant 1") {
		t.Errorf("expected variant 1 in description, got %q", actions[0].Description)
	}
	if !strings.Contains(actions[1].Description, "variant 2") {
		t.Errorf("expected variant 2 in description, got %q", actions[1].Description)
	}
}

// ---------------------------------------------------------------------------
// TestExtractCodeBlocks
// ---------------------------------------------------------------------------

func TestExtractCodeBlocks(t *testing.T) {
	type args struct {
		text string
	}
	type expected struct {
		blocks []string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "single block",
			args:     args{text: "```\nhello world\n```"},
			expected: expected{blocks: []string{"hello world"}},
		},
		{
			testName: "two blocks",
			args:     args{text: "A:\n```\nfirst\n```\nB:\n```\nsecond\n```"},
			expected: expected{blocks: []string{"first", "second"}},
		},
		{
			testName: "block with language tag",
			args:     args{text: "```python\nprint('hi')\n```"},
			expected: expected{blocks: []string{"print('hi')"}},
		},

		// 異常系
		{
			testName: "no blocks",
			args:     args{text: "plain text"},
			expected: expected{blocks: nil},
		},
		{
			testName: "unclosed block",
			args:     args{text: "```\nno close"},
			expected: expected{blocks: nil},
		},
		{
			testName: "only opening markers no newline",
			args:     args{text: "```"},
			expected: expected{blocks: nil},
		},

		// 境界値
		{
			testName: "empty block",
			args:     args{text: "```\n\n```"},
			expected: expected{blocks: nil},
		},
		{
			testName: "whitespace block",
			args:     args{text: "```\n   \n```"},
			expected: expected{blocks: nil},
		},

		// 特殊文字
		{
			testName: "block with backticks inside content",
			args:     args{text: "```\nuse `code` here\n```"},
			expected: expected{blocks: []string{"use `code` here"}},
		},

		// 空文字
		{
			testName: "empty string",
			args:     args{text: ""},
			expected: expected{blocks: nil},
		},

		// Null/Nil
		{
			testName: "whitespace only",
			args:     args{text: "   "},
			expected: expected{blocks: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := extractCodeBlocks(tt.args.text)
			if diff := cmp.Diff(tt.expected.blocks, got); diff != "" {
				t.Errorf("blocks mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestExecuteAction
// ---------------------------------------------------------------------------

func TestExecuteAction(t *testing.T) {
	validPromptID := uuid.New()
	validAuthorID := uuid.New()
	versionID := prompt.PromptVersionIDFromUUID(uuid.New())

	validContent := json.RawMessage(`{"text":"You are an assistant."}`)

	type args struct {
		action   Action
		authorID uuid.UUID
	}
	type expected struct {
		wantErr       bool
		errName       string
		versionNumber int
	}

	tests := []struct {
		testName    string
		args        args
		expected    expected
		promptRepo  *mockPromptRepo
		versionRepo *mockVersionRepo
	}{
		// 正常系 (Happy Path)
		{
			testName: "successful version creation",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      "Suggested prompt from consulting chat",
				},
				authorID: validAuthorID,
			},
			expected: expected{wantErr: false, versionNumber: 4},
			promptRepo: &mockPromptRepo{
				findByIDFn: func(_ context.Context, _ prompt.PromptID) (prompt.Prompt, error) {
					return prompt.Prompt{
						ID:            prompt.PromptIDFromUUID(validPromptID),
						LatestVersion: 3,
					}, nil
				},
				updateLatestVersionFn: func(_ context.Context, _ prompt.PromptID, _ int) (prompt.Prompt, error) {
					return prompt.Prompt{}, nil
				},
			},
			versionRepo: &mockVersionRepo{
				createFn: func(_ context.Context, _ prompt.VersionCmd, vn int) (prompt.PromptVersion, error) {
					return prompt.PromptVersion{
						ID:            versionID,
						VersionNumber: vn,
					}, nil
				},
			},
		},

		// 異常系 (Error Cases)
		{
			testName: "unsupported action type",
			args: args{
				action: Action{
					Type:             ActionType("unknown"),
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      "test",
				},
				authorID: validAuthorID,
			},
			expected:    expected{wantErr: true, errName: "ValidationError"},
			promptRepo:  &mockPromptRepo{},
			versionRepo: &mockVersionRepo{},
		},
		{
			testName: "prompt not found",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      "test",
				},
				authorID: validAuthorID,
			},
			expected: expected{wantErr: true, errName: "NotFoundError"},
			promptRepo: &mockPromptRepo{
				findByIDFn: func(_ context.Context, _ prompt.PromptID) (prompt.Prompt, error) {
					return prompt.Prompt{}, apperror.NewNotFoundError(
						fmt.Errorf("prompt not found"),
						"Prompt",
					)
				},
			},
			versionRepo: &mockVersionRepo{},
		},
		{
			testName: "version create fails",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      "test",
				},
				authorID: validAuthorID,
			},
			expected: expected{wantErr: true, errName: "DatabaseError"},
			promptRepo: &mockPromptRepo{
				findByIDFn: func(_ context.Context, _ prompt.PromptID) (prompt.Prompt, error) {
					return prompt.Prompt{
						ID:            prompt.PromptIDFromUUID(validPromptID),
						LatestVersion: 1,
					}, nil
				},
			},
			versionRepo: &mockVersionRepo{
				createFn: func(_ context.Context, _ prompt.VersionCmd, _ int) (prompt.PromptVersion, error) {
					return prompt.PromptVersion{}, apperror.NewDatabaseError(
						fmt.Errorf("db error"),
						"PromptVersion",
					)
				},
			},
		},
		{
			testName: "update latest version fails",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      "test",
				},
				authorID: validAuthorID,
			},
			expected: expected{wantErr: true, errName: "DatabaseError"},
			promptRepo: &mockPromptRepo{
				findByIDFn: func(_ context.Context, _ prompt.PromptID) (prompt.Prompt, error) {
					return prompt.Prompt{
						ID:            prompt.PromptIDFromUUID(validPromptID),
						LatestVersion: 1,
					}, nil
				},
				updateLatestVersionFn: func(_ context.Context, _ prompt.PromptID, _ int) (prompt.Prompt, error) {
					return prompt.Prompt{}, apperror.NewDatabaseError(
						fmt.Errorf("db error"),
						"Prompt",
					)
				},
			},
			versionRepo: &mockVersionRepo{
				createFn: func(_ context.Context, _ prompt.VersionCmd, vn int) (prompt.PromptVersion, error) {
					return prompt.PromptVersion{ID: versionID, VersionNumber: vn}, nil
				},
			},
		},

		// 境界値 (Boundary Values)
		{
			testName: "prompt with latest version 0 creates version 1",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      "first version",
				},
				authorID: validAuthorID,
			},
			expected: expected{wantErr: false, versionNumber: 1},
			promptRepo: &mockPromptRepo{
				findByIDFn: func(_ context.Context, _ prompt.PromptID) (prompt.Prompt, error) {
					return prompt.Prompt{
						ID:            prompt.PromptIDFromUUID(validPromptID),
						LatestVersion: 0,
					}, nil
				},
				updateLatestVersionFn: func(_ context.Context, _ prompt.PromptID, _ int) (prompt.Prompt, error) {
					return prompt.Prompt{}, nil
				},
			},
			versionRepo: &mockVersionRepo{
				createFn: func(_ context.Context, _ prompt.VersionCmd, vn int) (prompt.PromptVersion, error) {
					return prompt.PromptVersion{ID: versionID, VersionNumber: vn}, nil
				},
			},
		},
		{
			testName: "description at max length (500 chars)",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      strings.Repeat("a", 500),
				},
				authorID: validAuthorID,
			},
			expected: expected{wantErr: false, versionNumber: 2},
			promptRepo: &mockPromptRepo{
				findByIDFn: func(_ context.Context, _ prompt.PromptID) (prompt.Prompt, error) {
					return prompt.Prompt{
						ID:            prompt.PromptIDFromUUID(validPromptID),
						LatestVersion: 1,
					}, nil
				},
				updateLatestVersionFn: func(_ context.Context, _ prompt.PromptID, _ int) (prompt.Prompt, error) {
					return prompt.Prompt{}, nil
				},
			},
			versionRepo: &mockVersionRepo{
				createFn: func(_ context.Context, _ prompt.VersionCmd, vn int) (prompt.PromptVersion, error) {
					return prompt.PromptVersion{ID: versionID, VersionNumber: vn}, nil
				},
			},
		},
		{
			testName: "description over max length (501 chars) fails",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      strings.Repeat("a", 501),
				},
				authorID: validAuthorID,
			},
			expected: expected{wantErr: true, errName: "ValidationError"},
			promptRepo: &mockPromptRepo{
				findByIDFn: func(_ context.Context, _ prompt.PromptID) (prompt.Prompt, error) {
					return prompt.Prompt{
						ID:            prompt.PromptIDFromUUID(validPromptID),
						LatestVersion: 1,
					}, nil
				},
			},
			versionRepo: &mockVersionRepo{},
		},

		// 特殊文字 (Special Characters)
		{
			testName: "description with emoji",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      "New version 🚀",
				},
				authorID: validAuthorID,
			},
			expected: expected{wantErr: false, versionNumber: 2},
			promptRepo: &mockPromptRepo{
				findByIDFn: func(_ context.Context, _ prompt.PromptID) (prompt.Prompt, error) {
					return prompt.Prompt{
						ID:            prompt.PromptIDFromUUID(validPromptID),
						LatestVersion: 1,
					}, nil
				},
				updateLatestVersionFn: func(_ context.Context, _ prompt.PromptID, _ int) (prompt.Prompt, error) {
					return prompt.Prompt{}, nil
				},
			},
			versionRepo: &mockVersionRepo{
				createFn: func(_ context.Context, _ prompt.VersionCmd, vn int) (prompt.PromptVersion, error) {
					return prompt.PromptVersion{ID: versionID, VersionNumber: vn}, nil
				},
			},
		},
		{
			testName: "description with Japanese",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: validContent,
					Description:      "コンサルティングから生成",
				},
				authorID: validAuthorID,
			},
			expected: expected{wantErr: false, versionNumber: 2},
			promptRepo: &mockPromptRepo{
				findByIDFn: func(_ context.Context, _ prompt.PromptID) (prompt.Prompt, error) {
					return prompt.Prompt{
						ID:            prompt.PromptIDFromUUID(validPromptID),
						LatestVersion: 1,
					}, nil
				},
				updateLatestVersionFn: func(_ context.Context, _ prompt.PromptID, _ int) (prompt.Prompt, error) {
					return prompt.Prompt{}, nil
				},
			},
			versionRepo: &mockVersionRepo{
				createFn: func(_ context.Context, _ prompt.VersionCmd, vn int) (prompt.PromptVersion, error) {
					return prompt.PromptVersion{ID: versionID, VersionNumber: vn}, nil
				},
			},
		},

		// 空文字 (Empty/Whitespace)
		{
			testName: "empty suggested content fails",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: json.RawMessage{},
					Description:      "test",
				},
				authorID: validAuthorID,
			},
			expected:    expected{wantErr: true, errName: "ValidationError"},
			promptRepo:  &mockPromptRepo{},
			versionRepo: &mockVersionRepo{},
		},
		{
			testName: "nil suggested content fails",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         validPromptID,
					SuggestedContent: nil,
					Description:      "test",
				},
				authorID: validAuthorID,
			},
			expected:    expected{wantErr: true, errName: "ValidationError"},
			promptRepo:  &mockPromptRepo{},
			versionRepo: &mockVersionRepo{},
		},

		// Null/Nil (zero values)
		{
			testName: "nil prompt ID fails",
			args: args{
				action: Action{
					Type:             ActionCreateVersion,
					PromptID:         uuid.Nil,
					SuggestedContent: validContent,
					Description:      "test",
				},
				authorID: validAuthorID,
			},
			expected:    expected{wantErr: true, errName: "ValidationError"},
			promptRepo:  &mockPromptRepo{},
			versionRepo: &mockVersionRepo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			svc := NewActionService(tt.promptRepo, tt.versionRepo)
			got, err := svc.ExecuteAction(context.Background(), tt.args.action, tt.args.authorID)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expected.versionNumber, got.VersionNumber); diff != "" {
				t.Errorf("version number mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(string(tt.args.action.Type), string(got.Action.Type)); diff != "" {
				t.Errorf("result action type mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestNewActionService
// ---------------------------------------------------------------------------

func TestNewActionService(t *testing.T) {
	t.Run("creates service with valid repos", func(t *testing.T) {
		pr := &mockPromptRepo{}
		vr := &mockVersionRepo{}
		svc := NewActionService(pr, vr)
		if svc == nil {
			t.Fatal("expected non-nil service")
		}
		if svc.promptRepo == nil {
			t.Error("expected promptRepo to be set")
		}
		if svc.versionRepo == nil {
			t.Error("expected versionRepo to be set")
		}
	})

	t.Run("creates service with nil repos", func(t *testing.T) {
		svc := NewActionService(nil, nil)
		if svc == nil {
			t.Fatal("expected non-nil service")
		}
	})
}
