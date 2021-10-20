package pluginfx

import (
	"context"
	"errors"
	"fmt"
	"plugin"
	"reflect"
)

// PluginError indicates that a plugin could not be loaded.
type PluginError struct {
	Path string
	Err  error
}

func (pe *PluginError) Unwrap() error {
	return pe.Err
}

func (pe *PluginError) Error() string {
	return fmt.Sprintf("Unable to load plugin from path %s: %s", pe.Path, pe.Err)
}

// MissingSymbolError indicates that a symbol was not found.  This error is returned
// by Lookup to normalize errors coming from plugins.
type MissingSymbolError struct {
	Name string
	Err  error
}

func (mse *MissingSymbolError) Unwrap() error {
	return mse.Err
}

func (mse *MissingSymbolError) Error() string {
	if mse.Err != nil {
		return fmt.Sprintf("Missing symbol %s: %s", mse.Name, mse.Err)
	}

	return fmt.Sprintf("Missing symbol %s", mse.Name)
}

// InvalidConstructorError indicates that a symbol was not usable
// as an uber/fx constructor.
type InvalidConstructorError struct {
	Name string
	Type reflect.Type
}

func (ice *InvalidConstructorError) Error() string {
	return fmt.Sprintf("Symbol %s of type %T is not a valid constructor", ice.Name, ice.Type)
}

// InvalidLifecycleError indicates that a symbol was not usable
// as an uber/fx lifecycle callback via fx.Hook.
type InvalidLifecycleError struct {
	Name string
	Type reflect.Type
}

func (ile *InvalidLifecycleError) Error() string {
	return fmt.Sprintf("Symbol %s of type %T is not a valid lifecycle callback", ile.Name, ile.Type)
}

// Symbols defines the behavior of something that can look up
// exported symbols.  *plugin.Plugin implements this interface.
type Symbols interface {
	// Lookup returns the value of the given symbol, or an error
	// if no such symbol exists.
	Lookup(string) (plugin.Symbol, error)
}

// Open loads a set of Symbols from a path.  This is the analog to plugin.Open,
// and returns a *PluginError instead of a generated error.
func Open(path string) (Symbols, error) {
	p, err := plugin.Open(path)
	if err != nil {
		err = &PluginError{
			Path: path,
			Err:  err,
		}
	}

	return p, err
}

// Lookup invokes s.Lookup and normalizes any error to *MissingSymbolError.
func Lookup(s Symbols, name string) (interface{}, error) {
	var value reflect.Value
	symbol, err := s.Lookup(name)
	if err == nil {
		value = reflect.ValueOf(symbol)
	} else {
		var msError *MissingSymbolError
		if !errors.As(err, &msError) {
			err = &MissingSymbolError{
				Name: name,
				Err:  err,
			}
		}
	}

	return value, err
}

// Find attempts to locate a symbol in any of a set of Symbols objects.  Any Symbols
// object that is nil is skipped, which simplifies error handling.
//
// The Symbols are consulted in order, and the first one to return a symbol
// is returned.  If name was not found in any of the Symbols, *MissingSymbolError
// is returned.
//
// The primary use case for this function is for defaults:
//
//   var sm pluginfx.SymbolMap
//   sm.SetValue("port", 8080)
//   p, _ := pluginfx.Open("path.so")
//   v, err := Find("port", p, sm) // fallback to the symbol map if not in the plugin
func Find(name string, s ...Symbols) (interface{}, error) {
	for _, symbols := range s {
		if symbols != nil {
			if v, err := symbols.Lookup(name); err == nil {
				return v, nil
			}
		}
	}

	return nil, &MissingSymbolError{Name: name}
}

// isValidConstructor requires its argument to be a function with at least (1) non-error
// output parameter.
func isValidConstructor(value reflect.Value) bool {
	// easy case:
	if value.Kind() != reflect.Func || value.Type().NumOut() < 1 {
		return false
	}

	errType := reflect.TypeOf((*error)(nil)).Elem()
	vt := value.Type()
	for i := 0; i < vt.NumOut(); i++ {
		if vt.In(i) != errType {
			// first non-error output parameter means we're good
			return true
		}
	}

	return false
}

func checkConstructorSymbol(name string, value reflect.Value) error {
	if !isValidConstructor(value) {
		return &InvalidConstructorError{
			Name: name,
			Type: value.Type(),
		}
	}

	return nil
}

// LookupConstructor loads a symbol and verifies that it can be used as
// a constructor passed to fx.Provide.  The reflect.Value representing
// the function is returned along with any error.
//
// This function returns a *MissingSymbolError if name was not found.
// It returns *InvalidConstructorError if the symbol was found but it
// not a valid fx constructor.
func LookupConstructor(s Symbols, name string) (reflect.Value, error) {
	var value reflect.Value
	symbol, err := Lookup(s, name)
	if err == nil {
		value = reflect.ValueOf(symbol)
		err = checkConstructorSymbol(name, value)
	}

	return value, err
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
func LookupLifecycle(s Symbols, name string) (func(context.Context) error, error) {
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

// SymbolVisitor is a callback for symbols.  If an error is returned from Lookup,
// this callback is passed and invalid value along with that error.  Otherwise,
// this callback is passed the value of the symbol, and lookupErr will be nil.
//
// If this callback itself returns an error, then visitation is halted and that
// error is returned.
type SymbolVisitor func(symbolName string, value reflect.Value, lookupErr error) error

// EachSymbol applies a visitor over a set of symbols.
func EachSymbol(s Symbols, sv SymbolVisitor, names ...string) error {
	for _, n := range names {
		var value reflect.Value
		symbol, err := Lookup(s, n)
		if err == nil {
			value = reflect.ValueOf(symbol)
		}

		if visitorErr := sv(n, value, err); visitorErr != nil {
			return visitorErr
		}
	}

	return nil
}
