package concurrent

import (
	"context"
	"reflect"
	"sync"

	slogctx "github.com/veqryn/slog-context"
)

// SetupFunc is described in Setup.
type SetupFunc func(context.Context) (any, error)

// Setup processes map which keys must be references to variables and
// values must be functions which returns values for these variables to
// run in parallel all functions which corresponding variables is nil.
//
//	     var a, b *int
//		err = Setup(ctx, map[interface{}]SetupFunc{
//			&a: setA,
//			&b: setB,
//		})
//
// Returns first non-nil error returned by any of executed functions.
// It will panic if referenced variable can't be nil or corresponding
// function returns value which can't be assigned to that variable.
func Setup(ctx context.Context, vars map[any]SetupFunc) error {
	errc := make(chan error, len(vars))
	var wg sync.WaitGroup
	for v, setup := range vars {
		elem := reflect.ValueOf(v).Elem()
		if elem.IsNil() {
			wg.Add(1)
			go func(setup SetupFunc) {
				res, err := setup(ctx)
				if err == nil {
					elem.Set(reflect.ValueOf(res))
				}
				errc <- err
				wg.Done()
			}(setup)
		}
	}
	wg.Wait()
	close(errc)
	for err := range errc {
		if err != nil {
			return err
		}
	}
	return nil
}

// Serve runs given services in parallel until either ctx.Done or any
// service exits, then it call cancel and wait until all services will
// exit.
//
// Returns error of first service which returned non-nil error, if any.
func Serve(ctx context.Context, cancel func(), services ...func(context.Context) error) (err error) {
	errc := make(chan error)
	for _, service := range services {
		go func() {
			err := service(ctx)
			if err != nil {
				errc <- err
			}
		}()
	}

	select {
	case err := <-errc:
		slogctx.Error(ctx, "failed to start service", "error", err)
		cancel()
	case <-ctx.Done():
		cancel()
	}

	return err
}
