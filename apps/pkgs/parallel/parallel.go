package parallel

import "context"

type result[T any] struct {
	val T
	err error
}

func run[T any](ctx context.Context, f func(context.Context) (T, error)) <-chan result[T] {
	ch := make(chan result[T], 1)
	go func() {
		v, err := f(ctx)
		ch <- result[T]{val: v, err: err}
	}()
	return ch
}

func Parallel2[T1, T2 any](ctx context.Context, f1 func(context.Context) (T1, error), f2 func(context.Context) (T2, error)) (T1, T2, error) {
	ch1 := run(ctx, f1)
	ch2 := run(ctx, f2)

	var (
		res1 T1
		res2 T2
	)

	for count := 2; count > 0; {
		select {
		case <-ctx.Done():
			return res1, res2, ctx.Err()
		case r := <-ch1:
			if r.err != nil {
				return res1, res2, r.err
			}
			res1, ch1 = r.val, nil
			count--
		case r := <-ch2:
			if r.err != nil {
				return res1, res2, r.err
			}
			res2, ch2 = r.val, nil
			count--
		}
	}
	return res1, res2, nil
}

func Parallel3[T1, T2, T3 any](ctx context.Context, f1 func(context.Context) (T1, error), f2 func(context.Context) (T2, error), f3 func(context.Context) (T3, error)) (T1, T2, T3, error) {
	ch1 := run(ctx, f1)
	ch2 := run(ctx, f2)
	ch3 := run(ctx, f3)

	var (
		res1 T1
		res2 T2
		res3 T3
	)

	for count := 3; count > 0; {
		select {
		case <-ctx.Done():
			return res1, res2, res3, ctx.Err()
		case r := <-ch1:
			if r.err != nil {
				return res1, res2, res3, r.err
			}
			res1, ch1 = r.val, nil
			count--
		case r := <-ch2:
			if r.err != nil {
				return res1, res2, res3, r.err
			}
			res2, ch2 = r.val, nil
			count--
		case r := <-ch3:
			if r.err != nil {
				return res1, res2, res3, r.err
			}
			res3, ch3 = r.val, nil
			count--
		}
	}
	return res1, res2, res3, nil
}

func Parallel4[T1, T2, T3, T4 any](ctx context.Context, f1 func(context.Context) (T1, error), f2 func(context.Context) (T2, error), f3 func(context.Context) (T3, error), f4 func(context.Context) (T4, error)) (T1, T2, T3, T4, error) {
	ch1 := run(ctx, f1)
	ch2 := run(ctx, f2)
	ch3 := run(ctx, f3)
	ch4 := run(ctx, f4)

	var (
		res1 T1
		res2 T2
		res3 T3
		res4 T4
	)

	for count := 4; count > 0; {
		select {
		case <-ctx.Done():
			return res1, res2, res3, res4, ctx.Err()
		case r := <-ch1:
			if r.err != nil {
				return res1, res2, res3, res4, r.err
			}
			res1, ch1 = r.val, nil
			count--
		case r := <-ch2:
			if r.err != nil {
				return res1, res2, res3, res4, r.err
			}
			res2, ch2 = r.val, nil
			count--
		case r := <-ch3:
			if r.err != nil {
				return res1, res2, res3, res4, r.err
			}
			res3, ch3 = r.val, nil
			count--
		case r := <-ch4:
			if r.err != nil {
				return res1, res2, res3, res4, r.err
			}
			res4, ch4 = r.val, nil
			count--
		}
	}
	return res1, res2, res3, res4, nil
}

func Parallel5[T1, T2, T3, T4, T5 any](ctx context.Context, f1 func(context.Context) (T1, error), f2 func(context.Context) (T2, error), f3 func(context.Context) (T3, error), f4 func(context.Context) (T4, error), f5 func(context.Context) (T5, error)) (T1, T2, T3, T4, T5, error) {
	ch1 := run(ctx, f1)
	ch2 := run(ctx, f2)
	ch3 := run(ctx, f3)
	ch4 := run(ctx, f4)
	ch5 := run(ctx, f5)

	var (
		res1 T1
		res2 T2
		res3 T3
		res4 T4
		res5 T5
	)

	for count := 5; count > 0; {
		select {
		case <-ctx.Done():
			return res1, res2, res3, res4, res5, ctx.Err()
		case r := <-ch1:
			if r.err != nil {
				return res1, res2, res3, res4, res5, r.err
			}
			res1, ch1 = r.val, nil
			count--
		case r := <-ch2:
			if r.err != nil {
				return res1, res2, res3, res4, res5, r.err
			}
			res2, ch2 = r.val, nil
			count--
		case r := <-ch3:
			if r.err != nil {
				return res1, res2, res3, res4, res5, r.err
			}
			res3, ch3 = r.val, nil
			count--
		case r := <-ch4:
			if r.err != nil {
				return res1, res2, res3, res4, res5, r.err
			}
			res4, ch4 = r.val, nil
			count--
		case r := <-ch5:
			if r.err != nil {
				return res1, res2, res3, res4, res5, r.err
			}
			res5, ch5 = r.val, nil
			count--
		}
	}
	return res1, res2, res3, res4, res5, nil
}
