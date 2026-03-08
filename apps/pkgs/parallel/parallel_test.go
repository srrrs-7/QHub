package parallel

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/go-cmp/cmp"
)

// sleepWithContext sleeps for the given duration, but returns early if context is canceled.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func TestParallel2(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			type expected struct {
				V1  int
				V2  string
				Err error
			}

			ctx := context.Background()
			f1 := func(ctx context.Context) (int, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return 1, nil
			}
			f2 := func(ctx context.Context) (string, error) {
				sleepWithContext(ctx, 200*time.Millisecond)
				return "b", nil
			}

			v1, v2, err := Parallel2(ctx, f1, f2)
			got := expected{v1, v2, err}
			want := expected{1, "b", nil}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	})

	t.Run("error propagation", func(t *testing.T) {
		tests := []struct {
			name    string
			f1Err   error
			f2Err   error
			wantErr string
		}{
			{"f1 fails", errors.New("f1 error"), nil, "f1 error"},
			{"f2 fails", nil, errors.New("f2 error"), "f2 error"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					ctx := context.Background()
					f1 := func(ctx context.Context) (int, error) { return 0, tt.f1Err }
					f2 := func(ctx context.Context) (string, error) { return "", tt.f2Err }

					_, _, err := Parallel2(ctx, f1, f2)

					if err == nil {
						t.Fatal("expected error, got nil")
					}
					if diff := cmp.Diff(tt.wantErr, err.Error()); diff != "" {
						t.Errorf("error mismatch (-want +got):\n%s", diff)
					}
				})
			})
		}
	})

	t.Run("context cancellation during execution", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			f1 := func(ctx context.Context) (int, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return 0, err
				}
				return 1, nil
			}
			f2 := func(ctx context.Context) (string, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return "", err
				}
				return "b", nil
			}

			done := make(chan struct{})
			var err error
			go func() {
				_, _, err = Parallel2(ctx, f1, f2)
				close(done)
			}()

			synctest.Wait()
			cancel()
			synctest.Wait()
			<-done

			if !errors.Is(err, context.Canceled) {
				t.Errorf("got error %v, want context.Canceled", err)
			}
		})
	})

	t.Run("context already canceled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			f1 := func(ctx context.Context) (int, error) { return 1, nil }
			f2 := func(ctx context.Context) (string, error) { return "b", nil }

			_, _, err := Parallel2(ctx, f1, f2)

			if !errors.Is(err, context.Canceled) {
				t.Errorf("got error %v, want context.Canceled", err)
			}
		})
	})

	t.Run("parallel execution", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			start := time.Now()

			f1 := func(ctx context.Context) (int, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return 1, nil
			}
			f2 := func(ctx context.Context) (string, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return "b", nil
			}

			Parallel2(ctx, f1, f2)
			elapsed := time.Since(start)

			if elapsed >= 150*time.Millisecond {
				t.Errorf("execution took %v, expected parallel execution (~100ms)", elapsed)
			}
		})
	})
}

func TestParallel3(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			type expected struct {
				V1  int
				V2  string
				V3  bool
				Err error
			}

			ctx := context.Background()
			f1 := func(ctx context.Context) (int, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return 1, nil
			}
			f2 := func(ctx context.Context) (string, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return "b", nil
			}
			f3 := func(ctx context.Context) (bool, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return true, nil
			}

			v1, v2, v3, err := Parallel3(ctx, f1, f2, f3)
			got := expected{v1, v2, v3, err}
			want := expected{1, "b", true, nil}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	})

	t.Run("error propagation", func(t *testing.T) {
		tests := []struct {
			name    string
			f1Err   error
			f2Err   error
			f3Err   error
			wantErr string
		}{
			{"f1 fails", errors.New("f1 error"), nil, nil, "f1 error"},
			{"f2 fails", nil, errors.New("f2 error"), nil, "f2 error"},
			{"f3 fails", nil, nil, errors.New("f3 error"), "f3 error"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					ctx := context.Background()
					f1 := func(ctx context.Context) (int, error) { return 0, tt.f1Err }
					f2 := func(ctx context.Context) (string, error) { return "", tt.f2Err }
					f3 := func(ctx context.Context) (bool, error) { return false, tt.f3Err }

					_, _, _, err := Parallel3(ctx, f1, f2, f3)

					if err == nil {
						t.Fatal("expected error, got nil")
					}
					if diff := cmp.Diff(tt.wantErr, err.Error()); diff != "" {
						t.Errorf("error mismatch (-want +got):\n%s", diff)
					}
				})
			})
		}
	})

	t.Run("context cancellation during execution", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			f1 := func(ctx context.Context) (int, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return 0, err
				}
				return 1, nil
			}
			f2 := func(ctx context.Context) (string, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return "", err
				}
				return "b", nil
			}
			f3 := func(ctx context.Context) (bool, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return false, err
				}
				return true, nil
			}

			done := make(chan struct{})
			var err error
			go func() {
				_, _, _, err = Parallel3(ctx, f1, f2, f3)
				close(done)
			}()

			synctest.Wait()
			cancel()
			synctest.Wait()
			<-done

			if !errors.Is(err, context.Canceled) {
				t.Errorf("got error %v, want context.Canceled", err)
			}
		})
	})

	t.Run("context already canceled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			f1 := func(ctx context.Context) (int, error) { return 1, nil }
			f2 := func(ctx context.Context) (string, error) { return "b", nil }
			f3 := func(ctx context.Context) (bool, error) { return true, nil }

			_, _, _, err := Parallel3(ctx, f1, f2, f3)

			if !errors.Is(err, context.Canceled) {
				t.Errorf("got error %v, want context.Canceled", err)
			}
		})
	})
}

func TestParallel4(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			type expected struct {
				V1  int
				V2  string
				V3  bool
				V4  float64
				Err error
			}

			ctx := context.Background()
			f1 := func(ctx context.Context) (int, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return 1, nil
			}
			f2 := func(ctx context.Context) (string, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return "b", nil
			}
			f3 := func(ctx context.Context) (bool, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return true, nil
			}
			f4 := func(ctx context.Context) (float64, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return 3.14, nil
			}

			v1, v2, v3, v4, err := Parallel4(ctx, f1, f2, f3, f4)
			got := expected{v1, v2, v3, v4, err}
			want := expected{1, "b", true, 3.14, nil}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	})

	t.Run("error propagation", func(t *testing.T) {
		tests := []struct {
			name    string
			f1Err   error
			f2Err   error
			f3Err   error
			f4Err   error
			wantErr string
		}{
			{"f1 fails", errors.New("f1 error"), nil, nil, nil, "f1 error"},
			{"f2 fails", nil, errors.New("f2 error"), nil, nil, "f2 error"},
			{"f3 fails", nil, nil, errors.New("f3 error"), nil, "f3 error"},
			{"f4 fails", nil, nil, nil, errors.New("f4 error"), "f4 error"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					ctx := context.Background()
					f1 := func(ctx context.Context) (int, error) { return 0, tt.f1Err }
					f2 := func(ctx context.Context) (string, error) { return "", tt.f2Err }
					f3 := func(ctx context.Context) (bool, error) { return false, tt.f3Err }
					f4 := func(ctx context.Context) (float64, error) { return 0, tt.f4Err }

					_, _, _, _, err := Parallel4(ctx, f1, f2, f3, f4)

					if err == nil {
						t.Fatal("expected error, got nil")
					}
					if diff := cmp.Diff(tt.wantErr, err.Error()); diff != "" {
						t.Errorf("error mismatch (-want +got):\n%s", diff)
					}
				})
			})
		}
	})

	t.Run("context cancellation during execution", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			f1 := func(ctx context.Context) (int, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return 0, err
				}
				return 1, nil
			}
			f2 := func(ctx context.Context) (string, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return "", err
				}
				return "b", nil
			}
			f3 := func(ctx context.Context) (bool, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return false, err
				}
				return true, nil
			}
			f4 := func(ctx context.Context) (float64, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return 0, err
				}
				return 3.14, nil
			}

			done := make(chan struct{})
			var err error
			go func() {
				_, _, _, _, err = Parallel4(ctx, f1, f2, f3, f4)
				close(done)
			}()

			synctest.Wait()
			cancel()
			synctest.Wait()
			<-done

			if !errors.Is(err, context.Canceled) {
				t.Errorf("got error %v, want context.Canceled", err)
			}
		})
	})

	t.Run("context already canceled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			f1 := func(ctx context.Context) (int, error) { return 1, nil }
			f2 := func(ctx context.Context) (string, error) { return "b", nil }
			f3 := func(ctx context.Context) (bool, error) { return true, nil }
			f4 := func(ctx context.Context) (float64, error) { return 3.14, nil }

			_, _, _, _, err := Parallel4(ctx, f1, f2, f3, f4)

			if !errors.Is(err, context.Canceled) {
				t.Errorf("got error %v, want context.Canceled", err)
			}
		})
	})
}

func TestParallel5(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			type expected struct {
				V1  int
				V2  string
				V3  bool
				V4  float64
				V5  byte
				Err error
			}

			ctx := context.Background()
			f1 := func(ctx context.Context) (int, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return 1, nil
			}
			f2 := func(ctx context.Context) (string, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return "b", nil
			}
			f3 := func(ctx context.Context) (bool, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return true, nil
			}
			f4 := func(ctx context.Context) (float64, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return 3.14, nil
			}
			f5 := func(ctx context.Context) (byte, error) {
				sleepWithContext(ctx, 100*time.Millisecond)
				return 'a', nil
			}

			v1, v2, v3, v4, v5, err := Parallel5(ctx, f1, f2, f3, f4, f5)
			got := expected{v1, v2, v3, v4, v5, err}
			want := expected{1, "b", true, 3.14, 'a', nil}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	})

	t.Run("error propagation", func(t *testing.T) {
		tests := []struct {
			name    string
			f1Err   error
			f2Err   error
			f3Err   error
			f4Err   error
			f5Err   error
			wantErr string
		}{
			{"f1 fails", errors.New("f1 error"), nil, nil, nil, nil, "f1 error"},
			{"f2 fails", nil, errors.New("f2 error"), nil, nil, nil, "f2 error"},
			{"f3 fails", nil, nil, errors.New("f3 error"), nil, nil, "f3 error"},
			{"f4 fails", nil, nil, nil, errors.New("f4 error"), nil, "f4 error"},
			{"f5 fails", nil, nil, nil, nil, errors.New("f5 error"), "f5 error"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					ctx := context.Background()
					f1 := func(ctx context.Context) (int, error) { return 0, tt.f1Err }
					f2 := func(ctx context.Context) (string, error) { return "", tt.f2Err }
					f3 := func(ctx context.Context) (bool, error) { return false, tt.f3Err }
					f4 := func(ctx context.Context) (float64, error) { return 0, tt.f4Err }
					f5 := func(ctx context.Context) (byte, error) { return 0, tt.f5Err }

					_, _, _, _, _, err := Parallel5(ctx, f1, f2, f3, f4, f5)

					if err == nil {
						t.Fatal("expected error, got nil")
					}
					if diff := cmp.Diff(tt.wantErr, err.Error()); diff != "" {
						t.Errorf("error mismatch (-want +got):\n%s", diff)
					}
				})
			})
		}
	})

	t.Run("context cancellation during execution", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			f1 := func(ctx context.Context) (int, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return 0, err
				}
				return 1, nil
			}
			f2 := func(ctx context.Context) (string, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return "", err
				}
				return "b", nil
			}
			f3 := func(ctx context.Context) (bool, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return false, err
				}
				return true, nil
			}
			f4 := func(ctx context.Context) (float64, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return 0, err
				}
				return 3.14, nil
			}
			f5 := func(ctx context.Context) (byte, error) {
				if err := sleepWithContext(ctx, time.Second); err != nil {
					return 0, err
				}
				return 'a', nil
			}

			done := make(chan struct{})
			var err error
			go func() {
				_, _, _, _, _, err = Parallel5(ctx, f1, f2, f3, f4, f5)
				close(done)
			}()

			synctest.Wait()
			cancel()
			synctest.Wait()
			<-done

			if !errors.Is(err, context.Canceled) {
				t.Errorf("got error %v, want context.Canceled", err)
			}
		})
	})

	t.Run("context already canceled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			f1 := func(ctx context.Context) (int, error) { return 1, nil }
			f2 := func(ctx context.Context) (string, error) { return "b", nil }
			f3 := func(ctx context.Context) (bool, error) { return true, nil }
			f4 := func(ctx context.Context) (float64, error) { return 3.14, nil }
			f5 := func(ctx context.Context) (byte, error) { return 'a', nil }

			_, _, _, _, _, err := Parallel5(ctx, f1, f2, f3, f4, f5)

			if !errors.Is(err, context.Canceled) {
				t.Errorf("got error %v, want context.Canceled", err)
			}
		})
	})
}
