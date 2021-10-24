package pluginfx

import (
	"context"
	"fmt"
	"reflect"

	"go.uber.org/fx"
)

// InvalidLifecycleError indicates that a symbol was not usable
// as an uber/fx lifecycle callback via fx.Hook.
type InvalidLifecycleError struct {
	Name string
	Type reflect.Type
}

func (ile *InvalidLifecycleError) Error() string {
	return fmt.Sprintf("Symbol %s of type %T is not a valid lifecycle callback", ile.Name, ile.Type)
}

// LookupLifecycle loads a symbol that is assumed to be a lifecycle callback
// for fx.Lifecycle, either OnStart or OnStop.
//
// The symbol must be a function with one of several signatures:
//
//   - func()
//   - func() error
//   - func(context.Context)
//   - func(context.Context) error
//
// Any of those signatures will be converted as necessary to what is required
// by fx.Hook.
//
// This function returns a *MissingSymbolError if name was not found.
// It returns *InvalidLifecycleError if the symbol was not a function with
// one of the above signatures.
func LookupLifecycle(s Plugin, name string) (func(context.Context) error, error) {
	var callback func(context.Context) error
	symbol, err := Lookup(s, name)

	if err == nil {
		switch f := symbol.(type) {
		case func():
			callback = func(context.Context) error { f(); return nil }

		case func() error:
			callback = func(context.Context) error { return f() }

		case func(context.Context):
			callback = func(ctx context.Context) error { f(ctx); return nil }

		case func(context.Context) error:
			callback = f

		default:
			err = &InvalidLifecycleError{
				Name: name,
				Type: reflect.TypeOf(symbol),
			}
		}
	}

	return callback, err
}

// Lifecycle describes how to bind a plugin to an enclosing application's lifecycle.
type Lifecycle struct {
	// OnStart is the optional symbol name of a function that can be invoked on application startup.
	OnStart string

	// OnStop is the optional symbol name of a function that can be invoked on application shutdown.
	OnStop string

	// IgnoreMissing defines what happens when either OnStart or OnStop are set and not present.
	// If this field is true, a missing OnStart or OnStop is silently ignored.  If this field is false,
	// then a missing OnStart or OnStop from a plugin will shortcircuit application startup with an error.
	IgnoreMissing bool
}

func (lc Lifecycle) Provide(p Plugin) fx.Option {
	var (
		hook    fx.Hook
		options []fx.Option
	)

	if len(lc.OnStart) > 0 {
		var err error
		hook.OnStart, err = LookupLifecycle(p, lc.OnStart)
		missing := IsMissingSymbolError(err)
		if (missing && !lc.IgnoreMissing) || (!missing && err != nil) {
			options = append(options, fx.Error(err))
		}
	}

	if len(lc.OnStop) > 0 {
		var err error
		hook.OnStop, err = LookupLifecycle(p, lc.OnStop)
		missing := IsMissingSymbolError(err)
		if (missing && !lc.IgnoreMissing) || (!missing && err != nil) {
			options = append(options, fx.Error(err))
		}
	}

	if len(options) == 0 && (hook.OnStart != nil || hook.OnStop != nil) {
		return fx.Invoke(
			func(l fx.Lifecycle) {
				l.Append(hook)
			},
		)
	}

	return fx.Options(options...)
}
