package types

// Result型
type Result[T any, E any] struct {
	value *T
	err   *E
}

// コンストラクタ
func Ok[T any, E any](value T) Result[T, E] {
	return Result[T, E]{value: &value}
}

func Err[T any, E any](err E) Result[T, E] {
	return Result[T, E]{err: &err}
}

// FromPair converts a (value, error) pair to a Result
func FromPair[T any](value T, err error) Result[T, error] {
	if err != nil {
		return Err[T, error](err)
	}
	return Ok[T, error](value)
}

// 判定メソッド
func (r Result[T, E]) IsOk() bool {
	return r.err == nil
}

func (r Result[T, E]) IsErr() bool {
	return r.err != nil
}

// データ取り出し
func (r Result[T, E]) UnwrapOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return *r.value
}

// Map系メソッド
func Map[T, U, E any](r Result[T, E], fn func(T) U) Result[U, E] {
	if r.err != nil {
		return Err[U](*r.err)
	}
	return Ok[U, E](fn(*r.value))
}

func MapErr[T, E, F any](r Result[T, E], fn func(E) F) Result[T, F] {
	if r.err != nil {
		return Err[T](fn(*r.err))
	}
	return Ok[T, F](*r.value)
}

func FlatMap[T, U, E any](r Result[T, E], fn func(T) Result[U, E]) Result[U, E] {
	if r.err != nil {
		return Err[U](*r.err)
	}
	return fn(*r.value)
}

// AndThenMap - 型変換を伴うAndThen
func AndThen[T, U, E any](r Result[T, E], fn func(T) Result[U, E]) Result[U, E] {
	return FlatMap(r, fn)
}

// Match
func (r Result[T, E]) Match(onOk func(T), onErr func(E)) {
	if r.err != nil {
		onErr(*r.err)
	} else {
		onOk(*r.value)
	}
}

// Combine - 複数のResultを結合
func Combine[T, E any](results ...Result[T, E]) Result[[]T, E] {
	var values []T
	for _, r := range results {
		if r.err != nil {
			return Err[[]T](*r.err)
		}
		values = append(values, *r.value)
	}
	return Ok[[]T, E](values)
}

// Pipe2 chains: FlatMap -> Map
// A --(f1)--> B --(f2)--> C
func Pipe2[A, B, C, Err any](
	r Result[A, Err],
	f1 func(A) Result[B, Err],
	f2 func(B) C,
) Result[C, Err] {
	return Map(FlatMap(r, f1), f2)
}

// Pipe3 chains: FlatMap -> FlatMap -> Map
// A --(f1)--> B --(f2)--> C --(f3)--> D
func Pipe3[A, B, C, D, Err any](
	r Result[A, Err],
	f1 func(A) Result[B, Err],
	f2 func(B) Result[C, Err],
	f3 func(C) D,
) Result[D, Err] {
	return Map(FlatMap(FlatMap(r, f1), f2), f3)
}

// Pipe4 chains: FlatMap -> FlatMap -> FlatMap -> Map
// A --(f1)--> B --(f2)--> C --(f3)--> D --(f4)--> E
func Pipe4[A, B, C, D, E, Err any](
	r Result[A, Err],
	f1 func(A) Result[B, Err],
	f2 func(B) Result[C, Err],
	f3 func(C) Result[D, Err],
	f4 func(D) E,
) Result[E, Err] {
	return Map(FlatMap(FlatMap(FlatMap(r, f1), f2), f3), f4)
}

// Pipe5 chains: FlatMap -> FlatMap -> FlatMap -> FlatMap -> Map
// A --(f1)--> B --(f2)--> C --(f3)--> D --(f4)--> E --(f5)--> F
func Pipe5[A, B, C, D, E, F, Err any](
	r Result[A, Err],
	f1 func(A) Result[B, Err],
	f2 func(B) Result[C, Err],
	f3 func(C) Result[D, Err],
	f4 func(D) Result[E, Err],
	f5 func(E) F,
) Result[F, Err] {
	return Map(FlatMap(FlatMap(FlatMap(FlatMap(r, f1), f2), f3), f4), f5)
}
