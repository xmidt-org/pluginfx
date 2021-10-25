package pluginfx

import (
	"errors"
	"fmt"
	"plugin"
	"reflect"
)

// Plugin defines the behavior of something that can look up
// exported symbols.  *plugin.Plugin implements this interface.
type Plugin interface {
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

// Open loads a Plugin from a path.  This is the analog to plugin.Open,
// and returns a *OpenError instead of a generated error.
func Open(path string) (Plugin, error) {
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
// This function is a shorthand for situations where calling code only needs
// to be aware that an error indicated a symbol was missing.
func IsMissingSymbolError(err error) bool {
	var mse *MissingSymbolError
	return errors.As(err, &mse)
}

// Lookup invokes s.Lookup and normalizes any error to *MissingSymbolError.
func Lookup(p Plugin, name string) (interface{}, error) {
	var value reflect.Value
	symbol, err := p.Lookup(name)
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
