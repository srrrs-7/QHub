package types_test

import (
	"errors"
	"math"
	"testing"
	"utils/types"

	"github.com/stretchr/testify/assert"
)

// TestOk tests the Ok constructor
func TestOk(t *testing.T) {
	type args struct {
		value any
	}
	type expected struct {
		isOk bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "create Ok with int",
			args:     args{value: 42},
			expected: expected{isOk: true},
		},
		{
			testName: "create Ok with string",
			args:     args{value: "test"},
			expected: expected{isOk: true},
		},
		{
			testName: "create Ok with struct",
			args:     args{value: struct{ Name string }{Name: "test"}},
			expected: expected{isOk: true},
		},

		// 📏 境界値
		{
			testName: "Ok with zero value",
			args:     args{value: 0},
			expected: expected{isOk: true},
		},
		{
			testName: "Ok with negative value",
			args:     args{value: -42},
			expected: expected{isOk: true},
		},
		{
			testName: "Ok with max int",
			args:     args{value: math.MaxInt64},
			expected: expected{isOk: true},
		},
		{
			testName: "Ok with min int",
			args:     args{value: math.MinInt64},
			expected: expected{isOk: true},
		},

		// 🔤 特殊文字
		{
			testName: "Ok with emoji",
			args:     args{value: "🎉"},
			expected: expected{isOk: true},
		},
		{
			testName: "Ok with Japanese",
			args:     args{value: "テスト"},
			expected: expected{isOk: true},
		},
		{
			testName: "Ok with special characters",
			args:     args{value: "!@#$%^&*()"},
			expected: expected{isOk: true},
		},

		// 📭 空文字
		{
			testName: "Ok with empty string",
			args:     args{value: ""},
			expected: expected{isOk: true},
		},
		{
			testName: "Ok with whitespace",
			args:     args{value: "   "},
			expected: expected{isOk: true},
		},

		// ⚠️ Nil
		{
			testName: "Ok with nil slice",
			args:     args{value: []int(nil)},
			expected: expected{isOk: true},
		},
		{
			testName: "Ok with empty slice",
			args:     args{value: []int{}},
			expected: expected{isOk: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var result types.Result[any, error]

			switch v := tt.args.value.(type) {
			case int:
				result = types.Result[any, error](types.Ok[any, error](v))
			case string:
				result = types.Result[any, error](types.Ok[any, error](v))
			default:
				result = types.Ok[any, error](v)
			}

			assert.Equal(t, tt.expected.isOk, result.IsOk())
			assert.False(t, result.IsErr())
		})
	}
}

// TestErr tests the Err constructor
func TestErr(t *testing.T) {
	type args struct {
		err error
	}
	type expected struct {
		isErr bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "create Err with error",
			args:     args{err: errors.New("error")},
			expected: expected{isErr: true},
		},
		{
			testName: "create Err with formatted error",
			args:     args{err: errors.New("error: something went wrong")},
			expected: expected{isErr: true},
		},

		// 🔤 特殊文字
		{
			testName: "Err with emoji in error",
			args:     args{err: errors.New("error 🔥")},
			expected: expected{isErr: true},
		},
		{
			testName: "Err with Japanese in error",
			args:     args{err: errors.New("エラーが発生しました")},
			expected: expected{isErr: true},
		},
		{
			testName: "Err with special characters",
			args:     args{err: errors.New("error: !@#$%^&*()")},
			expected: expected{isErr: true},
		},

		// 📭 空文字
		{
			testName: "Err with empty error message",
			args:     args{err: errors.New("")},
			expected: expected{isErr: true},
		},

		// ⚠️ Nil - Note: This is a valid test case for error handling
		{
			testName: "Err with nil error",
			args:     args{err: nil},
			expected: expected{isErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.Err[int, error](tt.args.err)

			assert.Equal(t, tt.expected.isErr, result.IsErr())
			assert.False(t, result.IsOk())
		})
	}
}

// TestFromPair tests the FromPair converter
func TestFromPair(t *testing.T) {
	type args struct {
		value int
		err   error
	}
	type expected struct {
		isOk  bool
		value int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "convert successful pair",
			args:     args{value: 42, err: nil},
			expected: expected{isOk: true, value: 42},
		},
		{
			testName: "convert successful pair with zero",
			args:     args{value: 0, err: nil},
			expected: expected{isOk: true, value: 0},
		},

		// ❌ 異常系
		{
			testName: "convert error pair",
			args:     args{value: 0, err: errors.New("error")},
			expected: expected{isOk: false},
		},
		{
			testName: "convert error pair with non-zero value",
			args:     args{value: 42, err: errors.New("error")},
			expected: expected{isOk: false},
		},

		// 📏 境界値
		{
			testName: "convert with max int",
			args:     args{value: math.MaxInt64, err: nil},
			expected: expected{isOk: true, value: math.MaxInt64},
		},
		{
			testName: "convert with min int",
			args:     args{value: math.MinInt64, err: nil},
			expected: expected{isOk: true, value: math.MinInt64},
		},
		{
			testName: "convert negative value",
			args:     args{value: -1, err: nil},
			expected: expected{isOk: true, value: -1},
		},

		// 🔤 特殊文字（エラーメッセージ）
		{
			testName: "error with emoji",
			args:     args{value: 0, err: errors.New("error 🔥")},
			expected: expected{isOk: false},
		},
		{
			testName: "error with Japanese",
			args:     args{value: 0, err: errors.New("エラー")},
			expected: expected{isOk: false},
		},

		// 📭 空文字（エラーメッセージ）
		{
			testName: "error with empty message",
			args:     args{value: 0, err: errors.New("")},
			expected: expected{isOk: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.FromPair(tt.args.value, tt.args.err)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.False(t, result.IsErr())
				assert.Equal(t, tt.expected.value, result.UnwrapOr(0))
			} else {
				assert.False(t, result.IsOk())
				assert.True(t, result.IsErr())
			}
		})
	}
}

// TestIsOk tests the IsOk method
func TestIsOk(t *testing.T) {
	type args struct {
		result types.Result[int, error]
	}
	type expected struct {
		isOk bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "Ok result returns true",
			args:     args{result: types.Ok[int, error](42)},
			expected: expected{isOk: true},
		},

		// ❌ 異常系
		{
			testName: "Err result returns false",
			args:     args{result: types.Err[int, error](errors.New("error"))},
			expected: expected{isOk: false},
		},

		// 📏 境界値
		{
			testName: "Ok with zero returns true",
			args:     args{result: types.Ok[int, error](0)},
			expected: expected{isOk: true},
		},
		{
			testName: "Ok with negative returns true",
			args:     args{result: types.Ok[int, error](-1)},
			expected: expected{isOk: true},
		},

		// ⚠️ Nil
		{
			testName: "Err with nil error returns false",
			args:     args{result: types.Err[int, error](nil)},
			expected: expected{isOk: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected.isOk, tt.args.result.IsOk())
		})
	}
}

// TestIsErr tests the IsErr method
func TestIsErr(t *testing.T) {
	type args struct {
		result types.Result[int, error]
	}
	type expected struct {
		isErr bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "Err result returns true",
			args:     args{result: types.Err[int, error](errors.New("error"))},
			expected: expected{isErr: true},
		},

		// ❌ 異常系
		{
			testName: "Ok result returns false",
			args:     args{result: types.Ok[int, error](42)},
			expected: expected{isErr: false},
		},

		// 📏 境界値
		{
			testName: "Ok with zero returns false",
			args:     args{result: types.Ok[int, error](0)},
			expected: expected{isErr: false},
		},

		// ⚠️ Nil
		{
			testName: "Err with nil error returns true",
			args:     args{result: types.Err[int, error](nil)},
			expected: expected{isErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected.isErr, tt.args.result.IsErr())
		})
	}
}

// TestUnwrapOr tests the UnwrapOr method
func TestUnwrapOr(t *testing.T) {
	type args struct {
		result       types.Result[int, error]
		defaultValue int
	}
	type expected struct {
		value int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "Ok returns value",
			args:     args{result: types.Ok[int, error](42), defaultValue: 0},
			expected: expected{value: 42},
		},
		{
			testName: "Err returns default",
			args:     args{result: types.Err[int, error](errors.New("error")), defaultValue: 999},
			expected: expected{value: 999},
		},

		// 📏 境界値
		{
			testName: "Ok with zero returns zero",
			args:     args{result: types.Ok[int, error](0), defaultValue: 42},
			expected: expected{value: 0},
		},
		{
			testName: "Ok with negative returns negative",
			args:     args{result: types.Ok[int, error](-42), defaultValue: 0},
			expected: expected{value: -42},
		},
		{
			testName: "Err with zero default returns zero",
			args:     args{result: types.Err[int, error](errors.New("error")), defaultValue: 0},
			expected: expected{value: 0},
		},
		{
			testName: "Ok with max int",
			args:     args{result: types.Ok[int, error](math.MaxInt64), defaultValue: 0},
			expected: expected{value: math.MaxInt64},
		},
		{
			testName: "Err with max int default",
			args:     args{result: types.Err[int, error](errors.New("error")), defaultValue: math.MaxInt64},
			expected: expected{value: math.MaxInt64},
		},

		// ⚠️ Nil
		{
			testName: "Err with nil error returns default",
			args:     args{result: types.Err[int, error](nil), defaultValue: 42},
			expected: expected{value: 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected.value, tt.args.result.UnwrapOr(tt.args.defaultValue))
		})
	}
}

// TestMap tests the Map function
func TestMap(t *testing.T) {
	type args struct {
		result types.Result[int, error]
		fn     func(int) string
	}
	type expected struct {
		isOk  bool
		value string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "map Ok value to string",
			args: args{
				result: types.Ok[int, error](42),
				fn:     func(i int) string { return "value: 42" },
			},
			expected: expected{isOk: true, value: "value: 42"},
		},
		{
			testName: "map Err preserves error",
			args: args{
				result: types.Err[int, error](errors.New("error")),
				fn:     func(i int) string { return "should not be called" },
			},
			expected: expected{isOk: false},
		},

		// 📏 境界値
		{
			testName: "map with zero value",
			args: args{
				result: types.Ok[int, error](0),
				fn:     func(i int) string { return "zero" },
			},
			expected: expected{isOk: true, value: "zero"},
		},
		{
			testName: "map with negative value",
			args: args{
				result: types.Ok[int, error](-42),
				fn:     func(i int) string { return "negative" },
			},
			expected: expected{isOk: true, value: "negative"},
		},
		{
			testName: "map with max int",
			args: args{
				result: types.Ok[int, error](math.MaxInt64),
				fn:     func(i int) string { return "max" },
			},
			expected: expected{isOk: true, value: "max"},
		},

		// 🔤 特殊文字
		{
			testName: "map to string with emoji",
			args: args{
				result: types.Ok[int, error](1),
				fn:     func(i int) string { return "🎉" },
			},
			expected: expected{isOk: true, value: "🎉"},
		},
		{
			testName: "map to string with Japanese",
			args: args{
				result: types.Ok[int, error](42),
				fn:     func(i int) string { return "値: 42" },
			},
			expected: expected{isOk: true, value: "値: 42"},
		},

		// 📭 空文字
		{
			testName: "map to empty string",
			args: args{
				result: types.Ok[int, error](0),
				fn:     func(i int) string { return "" },
			},
			expected: expected{isOk: true, value: ""},
		},

		// ⚠️ Nil
		{
			testName: "map Err with nil error",
			args: args{
				result: types.Err[int, error](nil),
				fn:     func(i int) string { return "should not be called" },
			},
			expected: expected{isOk: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.Map(tt.args.result, tt.args.fn)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.Equal(t, tt.expected.value, result.UnwrapOr(""))
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}

// TestMapErr tests the MapErr function
func TestMapErr(t *testing.T) {
	type customError struct {
		msg string
	}

	type args struct {
		result types.Result[int, error]
		fn     func(error) customError
	}
	type expected struct {
		isOk      bool
		value     int
		customMsg string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "MapErr transforms error",
			args: args{
				result: types.Err[int, error](errors.New("original error")),
				fn:     func(e error) customError { return customError{msg: "custom: " + e.Error()} },
			},
			expected: expected{isOk: false, customMsg: "custom: original error"},
		},
		{
			testName: "MapErr preserves Ok value",
			args: args{
				result: types.Ok[int, error](42),
				fn:     func(e error) customError { return customError{msg: "should not be called"} },
			},
			expected: expected{isOk: true, value: 42},
		},

		// 🔤 特殊文字
		{
			testName: "MapErr with emoji in error",
			args: args{
				result: types.Err[int, error](errors.New("error 🔥")),
				fn:     func(e error) customError { return customError{msg: e.Error()} },
			},
			expected: expected{isOk: false, customMsg: "error 🔥"},
		},

		// 📭 空文字
		{
			testName: "MapErr with empty error message",
			args: args{
				result: types.Err[int, error](errors.New("")),
				fn:     func(e error) customError { return customError{msg: "empty"} },
			},
			expected: expected{isOk: false, customMsg: "empty"},
		},

		// ⚠️ Nil
		{
			testName: "MapErr with nil error",
			args: args{
				result: types.Err[int, error](nil),
				fn:     func(e error) customError { return customError{msg: "nil error"} },
			},
			expected: expected{isOk: false, customMsg: "nil error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.MapErr(tt.args.result, tt.args.fn)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.Equal(t, tt.expected.value, result.UnwrapOr(0))
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}

// TestFlatMap tests the FlatMap function
func TestFlatMap(t *testing.T) {
	type args struct {
		result types.Result[int, error]
		fn     func(int) types.Result[string, error]
	}
	type expected struct {
		isOk  bool
		value string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "FlatMap Ok to Ok",
			args: args{
				result: types.Ok[int, error](42),
				fn:     func(i int) types.Result[string, error] { return types.Ok[string, error]("success") },
			},
			expected: expected{isOk: true, value: "success"},
		},
		{
			testName: "FlatMap Ok to Err",
			args: args{
				result: types.Ok[int, error](42),
				fn:     func(i int) types.Result[string, error] { return types.Err[string, error](errors.New("fn error")) },
			},
			expected: expected{isOk: false},
		},
		{
			testName: "FlatMap Err preserves error",
			args: args{
				result: types.Err[int, error](errors.New("original error")),
				fn:     func(i int) types.Result[string, error] { return types.Ok[string, error]("should not be called") },
			},
			expected: expected{isOk: false},
		},

		// 📏 境界値
		{
			testName: "FlatMap with zero value",
			args: args{
				result: types.Ok[int, error](0),
				fn:     func(i int) types.Result[string, error] { return types.Ok[string, error]("zero") },
			},
			expected: expected{isOk: true, value: "zero"},
		},

		// 🔤 特殊文字
		{
			testName: "FlatMap to emoji string",
			args: args{
				result: types.Ok[int, error](1),
				fn:     func(i int) types.Result[string, error] { return types.Ok[string, error]("🎉") },
			},
			expected: expected{isOk: true, value: "🎉"},
		},

		// 📭 空文字
		{
			testName: "FlatMap to empty string",
			args: args{
				result: types.Ok[int, error](0),
				fn:     func(i int) types.Result[string, error] { return types.Ok[string, error]("") },
			},
			expected: expected{isOk: true, value: ""},
		},

		// ⚠️ Nil
		{
			testName: "FlatMap Err with nil error",
			args: args{
				result: types.Err[int, error](nil),
				fn:     func(i int) types.Result[string, error] { return types.Ok[string, error]("should not be called") },
			},
			expected: expected{isOk: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.FlatMap(tt.args.result, tt.args.fn)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.Equal(t, tt.expected.value, result.UnwrapOr(""))
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}

// TestMatch tests the Match method
func TestMatch(t *testing.T) {
	type args struct {
		result types.Result[int, error]
	}
	type expected struct {
		called string
		value  int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "Match calls onOk for Ok result",
			args:     args{result: types.Ok[int, error](42)},
			expected: expected{called: "onOk", value: 42},
		},
		{
			testName: "Match calls onErr for Err result",
			args:     args{result: types.Err[int, error](errors.New("error"))},
			expected: expected{called: "onErr"},
		},

		// 📏 境界値
		{
			testName: "Match with zero value calls onOk",
			args:     args{result: types.Ok[int, error](0)},
			expected: expected{called: "onOk", value: 0},
		},
		{
			testName: "Match with negative value calls onOk",
			args:     args{result: types.Ok[int, error](-42)},
			expected: expected{called: "onOk", value: -42},
		},

		// ⚠️ Nil
		{
			testName: "Match with nil error calls onErr",
			args:     args{result: types.Err[int, error](nil)},
			expected: expected{called: "onErr"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var called string
			var value int

			tt.args.result.Match(
				func(v int) {
					called = "onOk"
					value = v
				},
				func(e error) {
					called = "onErr"
				},
			)

			assert.Equal(t, tt.expected.called, called)
			if tt.expected.called == "onOk" {
				assert.Equal(t, tt.expected.value, value)
			}
		})
	}
}

// TestCombine tests the Combine function
func TestCombine(t *testing.T) {
	type args struct {
		results []types.Result[int, error]
	}
	type expected struct {
		isOk   bool
		values []int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "Combine all Ok results",
			args: args{
				results: []types.Result[int, error]{
					types.Ok[int, error](1),
					types.Ok[int, error](2),
					types.Ok[int, error](3),
				},
			},
			expected: expected{isOk: true, values: []int{1, 2, 3}},
		},
		{
			testName: "Combine with one Err fails",
			args: args{
				results: []types.Result[int, error]{
					types.Ok[int, error](1),
					types.Err[int, error](errors.New("error")),
					types.Ok[int, error](3),
				},
			},
			expected: expected{isOk: false},
		},
		{
			testName: "Combine all Err results",
			args: args{
				results: []types.Result[int, error]{
					types.Err[int, error](errors.New("error1")),
					types.Err[int, error](errors.New("error2")),
				},
			},
			expected: expected{isOk: false},
		},

		// 📏 境界値
		{
			testName: "Combine empty slice",
			args: args{
				results: []types.Result[int, error]{},
			},
			expected: expected{isOk: true, values: nil},
		},
		{
			testName: "Combine single Ok",
			args: args{
				results: []types.Result[int, error]{
					types.Ok[int, error](42),
				},
			},
			expected: expected{isOk: true, values: []int{42}},
		},
		{
			testName: "Combine with zero values",
			args: args{
				results: []types.Result[int, error]{
					types.Ok[int, error](0),
					types.Ok[int, error](0),
				},
			},
			expected: expected{isOk: true, values: []int{0, 0}},
		},

		// ⚠️ Nil
		{
			testName: "Combine with nil error fails",
			args: args{
				results: []types.Result[int, error]{
					types.Ok[int, error](1),
					types.Err[int, error](nil),
				},
			},
			expected: expected{isOk: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.Combine(tt.args.results...)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.Equal(t, tt.expected.values, result.UnwrapOr(nil))
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}

// TestPipe2 tests the Pipe2 function
func TestPipe2(t *testing.T) {
	type args struct {
		initial types.Result[int, error]
		f1      func(int) types.Result[string, error]
		f2      func(string) bool
	}
	type expected struct {
		isOk  bool
		value bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "Pipe2 all Ok",
			args: args{
				initial: types.Ok[int, error](42),
				f1:      func(i int) types.Result[string, error] { return types.Ok[string, error]("success") },
				f2:      func(s string) bool { return true },
			},
			expected: expected{isOk: true, value: true},
		},
		{
			testName: "Pipe2 fails at f1",
			args: args{
				initial: types.Ok[int, error](42),
				f1:      func(i int) types.Result[string, error] { return types.Err[string, error](errors.New("f1 error")) },
				f2:      func(s string) bool { return true },
			},
			expected: expected{isOk: false},
		},
		{
			testName: "Pipe2 initial error",
			args: args{
				initial: types.Err[int, error](errors.New("initial error")),
				f1:      func(i int) types.Result[string, error] { return types.Ok[string, error]("should not be called") },
				f2:      func(s string) bool { return true },
			},
			expected: expected{isOk: false},
		},

		// 📏 境界値
		{
			testName: "Pipe2 with zero value",
			args: args{
				initial: types.Ok[int, error](0),
				f1:      func(i int) types.Result[string, error] { return types.Ok[string, error]("") },
				f2:      func(s string) bool { return false },
			},
			expected: expected{isOk: true, value: false},
		},

		// 🔤 特殊文字
		{
			testName: "Pipe2 with emoji",
			args: args{
				initial: types.Ok[int, error](1),
				f1:      func(i int) types.Result[string, error] { return types.Ok[string, error]("🎉") },
				f2:      func(s string) bool { return s == "🎉" },
			},
			expected: expected{isOk: true, value: true},
		},

		// ⚠️ Nil
		{
			testName: "Pipe2 with nil error",
			args: args{
				initial: types.Err[int, error](nil),
				f1:      func(i int) types.Result[string, error] { return types.Ok[string, error]("should not be called") },
				f2:      func(s string) bool { return true },
			},
			expected: expected{isOk: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.Pipe2(tt.args.initial, tt.args.f1, tt.args.f2)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.Equal(t, tt.expected.value, result.UnwrapOr(false))
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}

// TestPipe3 tests the Pipe3 function
func TestPipe3(t *testing.T) {
	type args struct {
		initial types.Result[int, error]
		f1      func(int) types.Result[string, error]
		f2      func(string) types.Result[bool, error]
		f3      func(bool) int
	}
	type expected struct {
		isOk  bool
		value int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "Pipe3 all Ok",
			args: args{
				initial: types.Ok[int, error](42),
				f1:      func(i int) types.Result[string, error] { return types.Ok[string, error]("success") },
				f2:      func(s string) types.Result[bool, error] { return types.Ok[bool, error](true) },
				f3:      func(b bool) int { return 100 },
			},
			expected: expected{isOk: true, value: 100},
		},
		{
			testName: "Pipe3 fails at f1",
			args: args{
				initial: types.Ok[int, error](42),
				f1:      func(i int) types.Result[string, error] { return types.Err[string, error](errors.New("f1 error")) },
				f2:      func(s string) types.Result[bool, error] { return types.Ok[bool, error](true) },
				f3:      func(b bool) int { return 100 },
			},
			expected: expected{isOk: false},
		},
		{
			testName: "Pipe3 fails at f2",
			args: args{
				initial: types.Ok[int, error](42),
				f1:      func(i int) types.Result[string, error] { return types.Ok[string, error]("success") },
				f2:      func(s string) types.Result[bool, error] { return types.Err[bool, error](errors.New("f2 error")) },
				f3:      func(b bool) int { return 100 },
			},
			expected: expected{isOk: false},
		},

		// 📏 境界値
		{
			testName: "Pipe3 with zero values",
			args: args{
				initial: types.Ok[int, error](0),
				f1:      func(i int) types.Result[string, error] { return types.Ok[string, error]("") },
				f2:      func(s string) types.Result[bool, error] { return types.Ok[bool, error](false) },
				f3:      func(b bool) int { return 0 },
			},
			expected: expected{isOk: true, value: 0},
		},

		// ⚠️ Nil
		{
			testName: "Pipe3 initial error",
			args: args{
				initial: types.Err[int, error](nil),
				f1:      func(i int) types.Result[string, error] { return types.Ok[string, error]("should not be called") },
				f2:      func(s string) types.Result[bool, error] { return types.Ok[bool, error](true) },
				f3:      func(b bool) int { return 100 },
			},
			expected: expected{isOk: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.Pipe3(tt.args.initial, tt.args.f1, tt.args.f2, tt.args.f3)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.Equal(t, tt.expected.value, result.UnwrapOr(0))
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}

// TestPipe4 tests the Pipe4 function
func TestPipe4(t *testing.T) {
	type args struct {
		initial types.Result[int, error]
		f1      func(int) types.Result[int, error]
		f2      func(int) types.Result[int, error]
		f3      func(int) types.Result[int, error]
		f4      func(int) int
	}
	type expected struct {
		isOk  bool
		value int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "Pipe4 all Ok - multiplication chain",
			args: args{
				initial: types.Ok[int, error](2),
				f1:      func(i int) types.Result[int, error] { return types.Ok[int, error](i * 2) },
				f2:      func(i int) types.Result[int, error] { return types.Ok[int, error](i * 3) },
				f3:      func(i int) types.Result[int, error] { return types.Ok[int, error](i * 5) },
				f4:      func(i int) int { return i + 10 },
			},
			expected: expected{isOk: true, value: 70}, // 2*2*3*5+10 = 60+10 = 70
		},
		{
			testName: "Pipe4 fails at f3",
			args: args{
				initial: types.Ok[int, error](1),
				f1:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f2:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f3:      func(i int) types.Result[int, error] { return types.Err[int, error](errors.New("f3 error")) },
				f4:      func(i int) int { return i },
			},
			expected: expected{isOk: false},
		},

		// 📏 境界値
		{
			testName: "Pipe4 with zero",
			args: args{
				initial: types.Ok[int, error](0),
				f1:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f2:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f3:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f4:      func(i int) int { return i },
			},
			expected: expected{isOk: true, value: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.Pipe4(tt.args.initial, tt.args.f1, tt.args.f2, tt.args.f3, tt.args.f4)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.Equal(t, tt.expected.value, result.UnwrapOr(0))
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}

// TestPipe5 tests the Pipe5 function
func TestPipe5(t *testing.T) {
	type args struct {
		initial types.Result[int, error]
		f1      func(int) types.Result[int, error]
		f2      func(int) types.Result[int, error]
		f3      func(int) types.Result[int, error]
		f4      func(int) types.Result[int, error]
		f5      func(int) int
	}
	type expected struct {
		isOk  bool
		value int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "Pipe5 all Ok - complex chain",
			args: args{
				initial: types.Ok[int, error](1),
				f1:      func(i int) types.Result[int, error] { return types.Ok[int, error](i + 1) },
				f2:      func(i int) types.Result[int, error] { return types.Ok[int, error](i * 2) },
				f3:      func(i int) types.Result[int, error] { return types.Ok[int, error](i + 3) },
				f4:      func(i int) types.Result[int, error] { return types.Ok[int, error](i * 4) },
				f5:      func(i int) int { return i + 5 },
			},
			expected: expected{isOk: true, value: 33}, // ((((1+1)*2)+3)*4)+5 = 33
		},
		{
			testName: "Pipe5 fails at f4",
			args: args{
				initial: types.Ok[int, error](1),
				f1:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f2:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f3:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f4:      func(i int) types.Result[int, error] { return types.Err[int, error](errors.New("f4 error")) },
				f5:      func(i int) int { return i },
			},
			expected: expected{isOk: false},
		},
		{
			testName: "Pipe5 initial error",
			args: args{
				initial: types.Err[int, error](errors.New("initial error")),
				f1:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f2:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f3:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f4:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f5:      func(i int) int { return i },
			},
			expected: expected{isOk: false},
		},

		// 📏 境界値
		{
			testName: "Pipe5 with zero",
			args: args{
				initial: types.Ok[int, error](0),
				f1:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f2:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f3:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f4:      func(i int) types.Result[int, error] { return types.Ok[int, error](i) },
				f5:      func(i int) int { return i },
			},
			expected: expected{isOk: true, value: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.Pipe5(tt.args.initial, tt.args.f1, tt.args.f2, tt.args.f3, tt.args.f4, tt.args.f5)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.Equal(t, tt.expected.value, result.UnwrapOr(0))
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}

// TestAndThen tests the AndThen function (alias for FlatMap)
func TestAndThen(t *testing.T) {
	type args struct {
		result types.Result[int, error]
		fn     func(int) types.Result[string, error]
	}
	type expected struct {
		isOk  bool
		value string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "AndThen Ok to Ok",
			args: args{
				result: types.Ok[int, error](42),
				fn:     func(i int) types.Result[string, error] { return types.Ok[string, error]("success") },
			},
			expected: expected{isOk: true, value: "success"},
		},
		{
			testName: "AndThen Err preserves error",
			args: args{
				result: types.Err[int, error](errors.New("error")),
				fn:     func(i int) types.Result[string, error] { return types.Ok[string, error]("should not be called") },
			},
			expected: expected{isOk: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := types.AndThen(tt.args.result, tt.args.fn)

			if tt.expected.isOk {
				assert.True(t, result.IsOk())
				assert.Equal(t, tt.expected.value, result.UnwrapOr(""))
			} else {
				assert.True(t, result.IsErr())
			}
		})
	}
}
