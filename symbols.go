package pluginfx

import (
	"context"
	"errors"
	"fmt"
	"plugin"
	"reflect"
)

// MissingSymbolError indicates that a symbol was not found.
//
// *plugin.Plugin does not return this error.  It returns a generated
// using fmt.Errorf.  This package's code normalizes these errors
// to errors of this type.
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

// SymbolMap is a map implementation of Symbols.  It allows for an in-memory
// implementation of a plugin for testing or for production defaults.
//
// The zero value of this type is a usable, empty "plugin".
type SymbolMap map[string]plugin.Symbol

func (sm *SymbolMap) set(name string, value plugin.Symbol) {
	if *sm == nil {
		*sm = make(SymbolMap)
	}

	(*sm)[name] = value
}

// Set establishes a symbol, overwriting any existing symbol with the given name.
// This method panics if value is not a function or a non-nil pointer, which are the
// only allowed types of symbols in an actual plugin.
func (sm *SymbolMap) Set(name string, value plugin.Symbol) {
	vv := reflect.ValueOf(value)
	if vv.Kind() != reflect.Func && (vv.Kind() != reflect.Ptr || vv.IsNil()) {
		panic("pluginfx.SymbolMap: A symbol must be either a function or a non-nil pointer")
	}

	sm.set(name, value)
}

// SetValue is similar to Set, but does not panic if value is not a function or
// a pointer.  Instead, in that case this method creates a pointer to the value
// and use that pointer as the value.  This method will still panic if value is a nil pointer, however.
func (sm *SymbolMap) SetValue(name string, value interface{}) {
	vv := reflect.ValueOf(value)
	if vv.Kind() == reflect.Ptr {
		if vv.IsNil() {
			panic("pluginfx.SymbolMap: A symbol must be either a function or a non-nil pointer")
		}
	} else if vv.Kind() != reflect.Func {
		newValue := reflect.New(vv.Type())
		newValue.Elem().Set(vv)
		value = newValue.Interface()
	}

	sm.set(name, value)
}

// Del removes a symbol from this map.
func (sm *SymbolMap) Del(name string) {
	if *sm != nil {
		delete(*sm, name)
	}
}

// Lookup implements the Symbols interface.
func (sm SymbolMap) Lookup(name string) (plugin.Symbol, error) {
	if s, ok := sm[name]; ok {
		return s, nil
	} else {
		// mimic the error returned by the plugin package
		return nil, &MissingSymbolError{Name: name}
	}
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

func LookupConstructor(s Symbols, name string) (reflect.Value, error) {
	var value reflect.Value
	symbol, err := Lookup(s, name)
	if err == nil {
		value = reflect.ValueOf(symbol)
		err = checkConstructorSymbol(name, value)
	}

	return value, err
}

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
