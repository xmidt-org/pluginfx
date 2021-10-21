package pluginfx

import (
	"errors"
	"fmt"
	"plugin"
	"reflect"
)

// Symbols defines the behavior of something that can look up
// exported symbols.  *plugin.Plugin implements this interface.
type Symbols interface {
	// Lookup returns the value of the given symbol, or an error
	// if no such symbol exists.
	//
	// The *plugin.Plugin type returns a generated error from this method,
	// making error disambiguation hard or impossible.  The Lookup function
	// in this package helps with that by ensuring that a *MissingSymbolError
	// is returned in cases where a symbol cannot be found.
	Lookup(string) (plugin.Symbol, error)
}

// OpenError is returned by Open to indicate that a source of symbols could not be loaded.
type OpenError struct {
	Path string
	Err  error
}

func (oe *OpenError) Unwrap() error {
	return oe.Err
}

func (oe *OpenError) Error() string {
	return fmt.Sprintf("Unable to load plugin from path %s: %s", oe.Path, oe.Err)
}

// Open loads a set of Symbols from a path.  This is the analog to plugin.Open,
// and returns a *OpenError instead of a generated error.
func Open(path string) (Symbols, error) {
	p, err := plugin.Open(path)
	if err != nil {
		err = &OpenError{
			Path: path,
			Err:  err,
		}
	}

	return p, err
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

// IsMissingSymbolError tests err to see if it is a *MissingSymbolError.
func IsMissingSymbolError(err error) bool {
	var mse *MissingSymbolError
	return errors.As(err, &mse)
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

// Find attempts to locate a symbol in any of a sequence of Symbols objects.  Any Symbols
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
