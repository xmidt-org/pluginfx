package pluginfx

import (
	"fmt"
	"reflect"

	"go.uber.org/fx"
)

// InvalidConstructorError indicates that a symbol was not usable
// as an uber/fx constructor.
type InvalidConstructorError struct {
	Name string
	Type reflect.Type
}

func (ice *InvalidConstructorError) Error() string {
	return fmt.Sprintf("Symbol %s of type %T is not a valid constructor", ice.Name, ice.Type)
}

func ValidateConstructorSymbol(name string, value reflect.Value) error {
	vt := value.Type()
	if vt.Kind() == reflect.Func && vt.NumOut() > 0 {
		errType := reflect.TypeOf((*error)(nil)).Elem()
		for i := 0; i < vt.NumOut(); i++ {
			if vt.In(i) != errType {
				// first non-error output parameter means we're good
				return nil
			}
		}
	}

	return &InvalidConstructorError{
		Name: name,
		Type: value.Type(),
	}
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
		err = ValidateConstructorSymbol(name, value)
	}

	return value, err
}

// InvalidTargetError indicates that a type was not valid for the
// fx.Annotated.Target field.  This is more restrictive than a constructor.
// Targets must return exactly (1) non-error object, with an optional error.
type InvalidTargetError struct {
	Name string
	Type reflect.Type
}

func (ite *InvalidTargetError) Error() string {
	return fmt.Sprintf("Symbol %s of type %T is not a valid target", ite.Name, ite.Type)
}

func ValidateTargetSymbol(name string, value reflect.Value) error {
	vt := value.Type()
	if vt.Kind() == reflect.Func && vt.NumOut() > 0 && vt.NumOut() < 3 {
		errType := reflect.TypeOf((*error)(nil)).Elem()
		var nonErrorCount int
		for i := 0; i < vt.NumOut(); i++ {
			if vt.In(i) != errType {
				nonErrorCount++
			}
		}

		if nonErrorCount == 1 {
			return nil
		}
	}

	return &InvalidTargetError{
		Name: name,
		Type: value.Type(),
	}
}

func LookupTargetSymbol(s Symbols, name string) (reflect.Value, error) {
	var value reflect.Value
	symbol, err := Lookup(s, name)
	if err == nil {
		value = reflect.ValueOf(symbol)
		err = ValidateTargetSymbol(name, value)
	}

	return value, err
}

// Annotated is an analog of fx.Annotated for plugin symbols.  This type
// gives more control over how a plugin constructor gets placed into
// the enclosing fx.App.
type Annotated struct {
	// Name is the optional name of the component emitted by the Constructor.
	// Either Name or Group must be set, or an error is raised.
	Name string

	// Group is the value group for the component emitted by the Constructor.
	// Either Name or Group must be set, or an error is raised.
	Group string

	// Target is the name of a function symbol that must be legal to
	// use with fx.Annotated.Target.
	Target string
}

// Constructors holds information about symbols within a plugin that are to
// be used as fx.Provide functions.
type Constructors struct {
	Symbols       []interface{}
	IgnoreMissing bool
}

func (ctors Constructors) Provide(s Symbols) fx.Option {
	var options []fx.Option
	for _, v := range ctors.Symbols {
		switch ctor := v.(type) {
		case string:
			f, err := Lookup(s, ctor)
			switch {
			case err != nil && !ctors.IgnoreMissing:
				options = append(options, fx.Error(err))

			case err == nil:
				if err := ValidateConstructorSymbol(ctor, reflect.ValueOf(f)); err != nil {
					options = append(options, fx.Error(err))
				} else {
					options = append(options, fx.Provide(f))
				}
			}

		case Annotated:
			f, err := Lookup(s, ctor.Target)
			switch {
			case err != nil && !ctors.IgnoreMissing:
				options = append(options, fx.Error(err))

			case err == nil:
				if err := ValidateTargetSymbol(ctor.Target, reflect.ValueOf(f)); err != nil {
					options = append(options, fx.Error(err))
				} else {
					options = append(options, fx.Provide(
						fx.Annotated{
							Name:   ctor.Name,
							Group:  ctor.Group,
							Target: f,
						},
					))
				}
			}

		default:
		}
	}

	return fx.Options(options...)
}
