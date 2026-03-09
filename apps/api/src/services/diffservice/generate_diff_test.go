package diffservice

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"api/src/domain/apperror"
	"api/src/domain/intelligence"
	"api/src/domain/prompt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// --- Mock VersionRepository ---

type mockVersionRepo struct {
	versions    map[string]prompt.PromptVersion // key: "promptID:versionNumber"
	updatedDiff map[string]json.RawMessage      // key: versionID
	updateErr   error
}

func newMockVersionRepo() *mockVersionRepo {
	return &mockVersionRepo{
		versions:    make(map[string]prompt.PromptVersion),
		updatedDiff: make(map[string]json.RawMessage),
	}
}

func (m *mockVersionRepo) key(promptID prompt.PromptID, number int) string {
	return promptID.String() + ":" + string(rune('0'+number))
}

func (m *mockVersionRepo) addVersion(v prompt.PromptVersion) {
	k := m.key(v.PromptID, v.VersionNumber)
	m.versions[k] = v
}

func (m *mockVersionRepo) FindByPromptAndNumber(_ context.Context, promptID prompt.PromptID, number int) (prompt.PromptVersion, error) {
	k := m.key(promptID, number)
	v, ok := m.versions[k]
	if !ok {
		return prompt.PromptVersion{}, errors.New("not found")
	}
	return v, nil
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

func (m *mockVersionRepo) Create(_ context.Context, _ prompt.VersionCmd, _ int) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}

func (m *mockVersionRepo) UpdateStatus(_ context.Context, _ prompt.PromptVersionID, _ prompt.VersionStatus) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}

func (m *mockVersionRepo) ArchiveProduction(_ context.Context, _ prompt.PromptID) error {
	return nil
}

func (m *mockVersionRepo) UpdateLintResult(_ context.Context, _ prompt.PromptVersionID, _ json.RawMessage) error {
	return nil
}

func (m *mockVersionRepo) UpdateSemanticDiff(_ context.Context, id prompt.PromptVersionID, diff json.RawMessage) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedDiff[id.String()] = diff
	return nil
}

func (m *mockVersionRepo) UpdateEmbedding(_ context.Context, _ prompt.PromptVersionID, _ []float32) error {
	return nil
}

var _ prompt.VersionRepository = (*mockVersionRepo)(nil)

// --- Helper ---

func makeVersion(promptID uuid.UUID, number int, content string) prompt.PromptVersion {
	contentJSON, _ := json.Marshal(map[string]string{"content": content})
	return prompt.PromptVersion{
		ID:            prompt.PromptVersionIDFromUUID(uuid.New()),
		PromptID:      prompt.PromptIDFromUUID(promptID),
		VersionNumber: number,
		Status:        prompt.StatusDraft,
		Content:       json.RawMessage(contentJSON),
	}
}

// --- Tests for NewDiffService ---

func TestNewDiffService(t *testing.T) {
	repo := newMockVersionRepo()
	svc := NewDiffService(repo, nil)

	if svc == nil {
		t.Fatal("expected non-nil DiffService")
	}
	if svc.versionRepo != repo {
		t.Error("expected versionRepo to be set")
	}
}

// --- Tests for GenerateDiff ---

func TestGenerateDiff(t *testing.T) {
	type args struct {
		fromContent string
		toContent   string
		fromVersion int
		toVersion   int
		missingFrom bool
		missingTo   bool
		updateErr   error
	}
	type expected struct {
		wantErr     bool
		errName     string
		hasChanges  bool
		summary     string
		specificity float64
	}

	promptID := uuid.New()

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "different content returns semantic diff with changes",
			args: args{
				fromContent: "Short prompt",
				toContent:   "A much longer prompt with detailed instructions for the assistant",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				wantErr:    false,
				hasChanges: true,
			},
		},
		// 正常系: identical content
		{
			testName: "identical content returns no changes",
			args: args{
				fromContent: "Hello world",
				toContent:   "Hello world",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				wantErr:    false,
				hasChanges: false,
				summary:    "No significant changes detected",
			},
		},
		// 正常系: variable changes
		{
			testName: "variable added detected",
			args: args{
				fromContent: "Hello world",
				toContent:   "Hello {{name}} world",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				wantErr:    false,
				hasChanges: true,
			},
		},
		// 正常系: tone shift
		{
			testName: "tone shift from casual to formal",
			args: args{
				fromContent: "just do it simply like cool",
				toContent:   "Please ensure you must kindly shall comply with required standards",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				wantErr:    false,
				hasChanges: true,
			},
		},
		// 異常系 (Error Cases): from version not found
		{
			testName: "from version not found returns NotFoundError",
			args: args{
				fromContent: "any",
				toContent:   "any",
				fromVersion: 1,
				toVersion:   2,
				missingFrom: true,
			},
			expected: expected{
				wantErr: true,
				errName: "NotFoundError",
			},
		},
		// 異常系: to version not found
		{
			testName: "to version not found returns NotFoundError",
			args: args{
				fromContent: "any",
				toContent:   "any",
				fromVersion: 1,
				toVersion:   2,
				missingTo:   true,
			},
			expected: expected{
				wantErr: true,
				errName: "NotFoundError",
			},
		},
		// 異常系: UpdateSemanticDiff fails
		{
			testName: "update semantic diff failure returns DatabaseError",
			args: args{
				fromContent: "Hello",
				toContent:   "World",
				fromVersion: 1,
				toVersion:   2,
				updateErr:   errors.New("db connection lost"),
			},
			expected: expected{
				wantErr: true,
				errName: "DatabaseError",
			},
		},
		// 境界値 (Boundary Values): large content difference (>500 chars)
		{
			testName: "large content diff triggers high impact",
			args: args{
				fromContent: "x",
				toContent:   string(make([]byte, 600)),
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				wantErr:    false,
				hasChanges: true,
			},
		},
		// 境界値: medium content difference (101-500 chars)
		{
			testName: "medium content diff triggers medium impact",
			args: args{
				fromContent: "x",
				toContent:   string(make([]byte, 200)),
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				wantErr:    false,
				hasChanges: true,
			},
		},
		// 特殊文字 (Special Chars)
		{
			testName: "unicode and emoji content",
			args: args{
				fromContent: "プロンプトテスト",
				toContent:   "新しいプロンプト 🚀 with {{name}}",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				wantErr:    false,
				hasChanges: true,
			},
		},
		// 空文字 (Empty/Whitespace)
		{
			testName: "empty from content",
			args: args{
				fromContent: "",
				toContent:   "New content here",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				wantErr:     false,
				hasChanges:  true,
				specificity: 0.0, // fromContent empty → specificity stays 0
			},
		},
		// 空文字: both empty
		{
			testName: "both empty content",
			args: args{
				fromContent: "",
				toContent:   "",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				wantErr:    false,
				hasChanges: false,
				summary:    "No significant changes detected",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			repo := newMockVersionRepo()
			if tt.args.updateErr != nil {
				repo.updateErr = tt.args.updateErr
			}

			if !tt.args.missingFrom {
				repo.addVersion(makeVersion(promptID, tt.args.fromVersion, tt.args.fromContent))
			}
			if !tt.args.missingTo {
				repo.addVersion(makeVersion(promptID, tt.args.toVersion, tt.args.toContent))
			}

			svc := NewDiffService(repo, nil)
			got, err := svc.GenerateDiff(context.Background(), promptID, tt.args.fromVersion, tt.args.toVersion)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Errorf("expected AppError, got %T: %v", err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil result")
			}

			if tt.expected.hasChanges && len(got.Changes) == 0 {
				t.Error("expected at least one change, got none")
			}
			if !tt.expected.hasChanges && len(got.Changes) != 0 {
				t.Errorf("expected no changes, got %d: %+v", len(got.Changes), got.Changes)
			}

			if tt.expected.summary != "" {
				if diff := cmp.Diff(tt.expected.summary, got.Summary); diff != "" {
					t.Errorf("summary mismatch (-want +got):\n%s", diff)
				}
			}

			// Verify diff was stored in repo
			if len(repo.updatedDiff) == 0 {
				t.Error("expected UpdateSemanticDiff to be called")
			}
		})
	}
}

// --- Tests for GenerateTextDiff ---

func TestGenerateTextDiff(t *testing.T) {
	type args struct {
		fromContent string
		toContent   string
		fromVersion int
		toVersion   int
		missingFrom bool
		missingTo   bool
	}
	type expected struct {
		wantErr     bool
		errName     string
		fromVersion int
		toVersion   int
		added       int
		removed     int
		equal       int
	}

	promptID := uuid.New()

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path): lines added
		{
			testName: "added lines detected",
			args: args{
				fromContent: "line1\nline2",
				toContent:   "line1\nline2\nline3",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       1,
				removed:     0,
				equal:       2,
			},
		},
		// 正常系: lines removed
		{
			testName: "removed lines detected",
			args: args{
				fromContent: "line1\nline2\nline3",
				toContent:   "line1\nline3",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       0,
				removed:     1,
				equal:       2,
			},
		},
		// 正常系: lines modified (removed + added)
		{
			testName: "modified lines show as removed and added",
			args: args{
				fromContent: "line1\nold line\nline3",
				toContent:   "line1\nnew line\nline3",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       1,
				removed:     1,
				equal:       2,
			},
		},
		// 正常系: identical content
		{
			testName: "identical content produces all equal lines",
			args: args{
				fromContent: "line1\nline2",
				toContent:   "line1\nline2",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       0,
				removed:     0,
				equal:       2,
			},
		},
		// 正常系: completely different content
		{
			testName: "completely different content",
			args: args{
				fromContent: "alpha\nbeta",
				toContent:   "gamma\ndelta",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       2,
				removed:     2,
				equal:       0,
			},
		},
		// 異常系 (Error Cases): from version not found
		{
			testName: "from version not found returns NotFoundError",
			args: args{
				fromContent: "any",
				toContent:   "any",
				fromVersion: 1,
				toVersion:   2,
				missingFrom: true,
			},
			expected: expected{
				wantErr: true,
				errName: "NotFoundError",
			},
		},
		// 異常系: to version not found
		{
			testName: "to version not found returns NotFoundError",
			args: args{
				fromContent: "any",
				toContent:   "any",
				fromVersion: 1,
				toVersion:   2,
				missingTo:   true,
			},
			expected: expected{
				wantErr: true,
				errName: "NotFoundError",
			},
		},
		// 境界値 (Boundary Values): single line
		{
			testName: "single line content",
			args: args{
				fromContent: "one line",
				toContent:   "different line",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       1,
				removed:     1,
				equal:       0,
			},
		},
		// 特殊文字 (Special Chars)
		{
			testName: "unicode and emoji lines",
			args: args{
				fromContent: "こんにちは\n世界",
				toContent:   "こんにちは\n世界\n🚀 rocket",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       1,
				removed:     0,
				equal:       2,
			},
		},
		// 空文字 (Empty/Whitespace)
		{
			testName: "empty from content to non-empty",
			args: args{
				fromContent: "",
				toContent:   "new line",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       1,
				removed:     1, // empty string splits to [""] which is removed
				equal:       0,
			},
		},
		// 空文字: both empty
		{
			testName: "both empty content",
			args: args{
				fromContent: "",
				toContent:   "",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       0,
				removed:     0,
				equal:       1, // empty string splits to [""] which matches
			},
		},
		// Null/Nil: content with only newlines
		{
			testName: "content with only newlines",
			args: args{
				fromContent: "\n\n",
				toContent:   "\n",
				fromVersion: 1,
				toVersion:   2,
			},
			expected: expected{
				fromVersion: 1,
				toVersion:   2,
				added:       0,
				removed:     1,
				equal:       2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			repo := newMockVersionRepo()

			if !tt.args.missingFrom {
				repo.addVersion(makeVersion(promptID, tt.args.fromVersion, tt.args.fromContent))
			}
			if !tt.args.missingTo {
				repo.addVersion(makeVersion(promptID, tt.args.toVersion, tt.args.toContent))
			}

			svc := NewDiffService(repo, nil)
			got, err := svc.GenerateTextDiff(context.Background(), promptID, tt.args.fromVersion, tt.args.toVersion)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Errorf("expected AppError, got %T: %v", err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil result")
			}

			if diff := cmp.Diff(tt.expected.fromVersion, got.FromVersion); diff != "" {
				t.Errorf("fromVersion mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.toVersion, got.ToVersion); diff != "" {
				t.Errorf("toVersion mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.added, got.Stats.Added); diff != "" {
				t.Errorf("added mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.removed, got.Stats.Removed); diff != "" {
				t.Errorf("removed mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.equal, got.Stats.Equal); diff != "" {
				t.Errorf("equal mismatch (-want +got):\n%s", diff)
			}

			// Verify hunks are populated
			if len(got.Hunks) == 0 {
				t.Error("expected at least one hunk")
			}

			// Verify total lines in hunk matches stats
			if len(got.Hunks) > 0 {
				totalLines := got.Stats.Added + got.Stats.Removed + got.Stats.Equal
				if diff := cmp.Diff(totalLines, len(got.Hunks[0].Lines)); diff != "" {
					t.Errorf("hunk lines count mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// --- Additional tests for GenerateDiff semantic diff details ---

func TestGenerateDiffSemanticDetails(t *testing.T) {
	promptID := uuid.New()

	t.Run("high impact for large content change", func(t *testing.T) {
		repo := newMockVersionRepo()
		largeContent := make([]byte, 600)
		for i := range largeContent {
			largeContent[i] = 'a'
		}
		repo.addVersion(makeVersion(promptID, 1, "x"))
		repo.addVersion(makeVersion(promptID, 2, string(largeContent)))

		svc := NewDiffService(repo, nil)
		got, err := svc.GenerateDiff(context.Background(), promptID, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		foundHigh := false
		for _, c := range got.Changes {
			if c.Impact == "high" && c.Category == "modified" {
				foundHigh = true
			}
		}
		if !foundHigh {
			t.Error("expected high impact change for large content diff")
		}
	})

	t.Run("medium impact for medium content change", func(t *testing.T) {
		repo := newMockVersionRepo()
		medContent := make([]byte, 200)
		for i := range medContent {
			medContent[i] = 'b'
		}
		repo.addVersion(makeVersion(promptID, 1, "x"))
		repo.addVersion(makeVersion(promptID, 2, string(medContent)))

		svc := NewDiffService(repo, nil)
		got, err := svc.GenerateDiff(context.Background(), promptID, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		foundMedium := false
		for _, c := range got.Changes {
			if c.Impact == "medium" {
				foundMedium = true
			}
		}
		if !foundMedium {
			t.Error("expected medium impact change for medium content diff")
		}
	})

	t.Run("variable added and removed in diff", func(t *testing.T) {
		repo := newMockVersionRepo()
		repo.addVersion(makeVersion(promptID, 1, "Hello {{name}} and {{age}}"))
		repo.addVersion(makeVersion(promptID, 2, "Hello {{name}} and {{email}}"))

		svc := NewDiffService(repo, nil)
		got, err := svc.GenerateDiff(context.Background(), promptID, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var foundAdded, foundRemoved bool
		for _, c := range got.Changes {
			if c.Category == "added" && c.Description == "Variable {{email}} added" {
				foundAdded = true
			}
			if c.Category == "removed" && c.Description == "Variable {{age}} removed" {
				foundRemoved = true
			}
		}
		if !foundAdded {
			t.Error("expected added change for {{email}}")
		}
		if !foundRemoved {
			t.Error("expected removed change for {{age}}")
		}
	})

	t.Run("specificity is zero when from content is empty", func(t *testing.T) {
		repo := newMockVersionRepo()
		repo.addVersion(makeVersion(promptID, 1, ""))
		repo.addVersion(makeVersion(promptID, 2, "some content"))

		svc := NewDiffService(repo, nil)
		got, err := svc.GenerateDiff(context.Background(), promptID, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Specificity != 0.0 {
			t.Errorf("expected specificity 0.0 for empty from content, got %f", got.Specificity)
		}
	})

	t.Run("specificity is positive when content grows", func(t *testing.T) {
		repo := newMockVersionRepo()
		repo.addVersion(makeVersion(promptID, 1, "short"))
		repo.addVersion(makeVersion(promptID, 2, "much longer content here"))

		svc := NewDiffService(repo, nil)
		got, err := svc.GenerateDiff(context.Background(), promptID, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Specificity <= 0 {
			t.Errorf("expected positive specificity when content grows, got %f", got.Specificity)
		}
	})

	t.Run("specificity is negative when content shrinks", func(t *testing.T) {
		repo := newMockVersionRepo()
		repo.addVersion(makeVersion(promptID, 1, "much longer content here"))
		repo.addVersion(makeVersion(promptID, 2, "short"))

		svc := NewDiffService(repo, nil)
		got, err := svc.GenerateDiff(context.Background(), promptID, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Specificity >= 0 {
			t.Errorf("expected negative specificity when content shrinks, got %f", got.Specificity)
		}
	})
}

// --- Additional tests for GenerateTextDiff line details ---

func TestGenerateTextDiffLineDetails(t *testing.T) {
	promptID := uuid.New()

	t.Run("equal lines have both old and new line numbers", func(t *testing.T) {
		repo := newMockVersionRepo()
		repo.addVersion(makeVersion(promptID, 1, "same"))
		repo.addVersion(makeVersion(promptID, 2, "same"))

		svc := NewDiffService(repo, nil)
		got, err := svc.GenerateTextDiff(context.Background(), promptID, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, line := range got.Hunks[0].Lines {
			if line.Type == "equal" {
				if line.OldLine == 0 {
					t.Error("expected OldLine to be set for equal line")
				}
				if line.NewLine == 0 {
					t.Error("expected NewLine to be set for equal line")
				}
			}
		}
	})

	t.Run("added lines have only new line numbers", func(t *testing.T) {
		repo := newMockVersionRepo()
		repo.addVersion(makeVersion(promptID, 1, "line1"))
		repo.addVersion(makeVersion(promptID, 2, "line1\nadded"))

		svc := NewDiffService(repo, nil)
		got, err := svc.GenerateTextDiff(context.Background(), promptID, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, line := range got.Hunks[0].Lines {
			if line.Type == "added" {
				if line.NewLine == 0 {
					t.Error("expected NewLine to be set for added line")
				}
				if line.OldLine != 0 {
					t.Error("expected OldLine to be 0 for added line")
				}
			}
		}
	})

	t.Run("removed lines have only old line numbers", func(t *testing.T) {
		repo := newMockVersionRepo()
		repo.addVersion(makeVersion(promptID, 1, "line1\nremoved"))
		repo.addVersion(makeVersion(promptID, 2, "line1"))

		svc := NewDiffService(repo, nil)
		got, err := svc.GenerateTextDiff(context.Background(), promptID, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, line := range got.Hunks[0].Lines {
			if line.Type == "removed" {
				if line.OldLine == 0 {
					t.Error("expected OldLine to be set for removed line")
				}
				if line.NewLine != 0 {
					t.Error("expected NewLine to be 0 for removed line")
				}
			}
		}
	})
}

// --- Additional buildDiff edge cases ---

func TestBuildDiffEdgeCases(t *testing.T) {
	tests := []struct {
		testName       string
		fromContent    string
		toContent      string
		expectChanges  int
		expectTone     string
		checkSpecific  bool
		expectSpecZero bool
	}{
		// 空文字: empty from content (covers specificity = 0 branch)
		{
			testName:       "empty from content sets specificity to zero",
			fromContent:    "",
			toContent:      "new content here",
			expectChanges:  1,
			checkSpecific:  true,
			expectSpecZero: true,
		},
		// 境界値: exactly 100 char difference (low impact)
		{
			testName:      "100 char diff is low impact",
			fromContent:   "x",
			toContent:     "x" + string(make([]byte, 100)),
			expectChanges: 1,
		},
		// 境界値: exactly 101 char difference (medium impact)
		{
			testName:      "101 char diff is medium impact",
			fromContent:   "",
			toContent:     string(make([]byte, 101)),
			expectChanges: 1,
		},
		// 境界値: exactly 500 char difference (medium impact)
		{
			testName:      "500 char diff is medium impact",
			fromContent:   "",
			toContent:     string(make([]byte, 500)),
			expectChanges: 1,
		},
		// 境界値: exactly 501 char difference (high impact)
		{
			testName:      "501 char diff is high impact",
			fromContent:   "",
			toContent:     string(make([]byte, 501)),
			expectChanges: 1,
		},
		// 特殊文字: SQL injection-like content
		{
			testName:      "SQL injection content treated as plain text",
			fromContent:   "normal prompt",
			toContent:     "'; DROP TABLE prompts; --",
			expectChanges: 1,
		},
		// Null/Nil equivalent: both empty
		{
			testName:       "both empty strings",
			fromContent:    "",
			toContent:      "",
			expectChanges:  0,
			checkSpecific:  true,
			expectSpecZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := buildDiff(tt.fromContent, tt.toContent)

			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if len(result.Changes) < tt.expectChanges {
				t.Errorf("expected at least %d changes, got %d: %+v", tt.expectChanges, len(result.Changes), result.Changes)
			}
			if tt.expectTone != "" {
				if diff := cmp.Diff(tt.expectTone, result.ToneShift); diff != "" {
					t.Errorf("tone shift mismatch (-want +got):\n%s", diff)
				}
			}
			if tt.checkSpecific && tt.expectSpecZero {
				if result.Specificity != 0.0 {
					t.Errorf("expected specificity 0.0, got %f", result.Specificity)
				}
			}
		})
	}
}

// --- Additional buildSummary tests ---

func TestBuildSummaryWithChanges(t *testing.T) {
	tests := []struct {
		testName string
		changes  []intelligence.DiffChange
		expected string
	}{
		{
			testName: "single change",
			changes: []intelligence.DiffChange{
				{Category: "modified", Description: "Content changed", Impact: "low"},
			},
			expected: "Content changed",
		},
		{
			testName: "multiple changes joined with semicolons",
			changes: []intelligence.DiffChange{
				{Category: "modified", Description: "Length changed", Impact: "low"},
				{Category: "added", Description: "Variable added", Impact: "high"},
			},
			expected: "Length changed; Variable added",
		},
		{
			testName: "nil changes",
			changes:  nil,
			expected: "No significant changes detected",
		},
		{
			testName: "empty slice",
			changes:  []intelligence.DiffChange{},
			expected: "No significant changes detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := buildSummary(tt.changes)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
