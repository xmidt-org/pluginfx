package pluginfx

import (
	"fmt"
	"plugin"
	"reflect"

	"go.uber.org/fx"
)

// errType is a convenient "cache" value for the reflection type describing error.
var errType = reflect.TypeOf((*error)(nil)).Elem()

// InvalidConstructorError indicates that a symbol was not usable
// as an uber/fx constructor.
type InvalidConstructorError struct {
	Name string
	Type reflect.Type
}

func (ice *InvalidConstructorError) Error() string {
	return fmt.Sprintf("Symbol %s of type %T is not a valid constructor", ice.Name, ice.Type)
}

// LookupConstructor loads a symbol and verifies that it can be used as
// a constructor passed to fx.Provide.  The reflect.Value representing
// the function is returned along with any error.
//
// This function returns a *MissingSymbolError if name was not found.
// It returns *InvalidConstructorError if the symbol was found but it
// not a valid fx constructor.
func LookupConstructor(p Plugin, name string) (value reflect.Value, err error) {
	var symbol plugin.Symbol
	symbol, err = Lookup(p, name)

	if err == nil {
		value = reflect.ValueOf(symbol)
		valueType := value.Type()
		if valueType.Kind() == reflect.Func && valueType.NumOut() > 0 {
			for i := 0; i < valueType.NumOut(); i++ {
				if valueType.Out(i) != errType {
					return // any non-error return type means it's a valid constructor
				}
			}
		}

		err = &InvalidConstructorError{
			Name: name,
			Type: valueType,
		}
	}

	return
}

// LookupTarget locates a symbol that is valid for the fx.Annotated.Target field.
// A Target is a Constructor with an additional constraint:  it my only return exactly
// (1) non-error object and my optionally return an error in addition.
func LookupTarget(s Plugin, name string) (value reflect.Value, err error) {
	var symbol plugin.Symbol
	symbol, err = Lookup(s, name)

	if err == nil {
		value = reflect.ValueOf(symbol)
		valueType := value.Type()
		if valueType.Kind() == reflect.Func {
			switch {
			case valueType.NumOut() == 1 && valueType.Out(0) != errType:
				return

			case valueType.NumOut() == 2 && valueType.Out(0) != errType && valueType.Out(1) == errType:
				return

			case valueType.NumOut() == 2 && valueType.Out(0) == errType && valueType.Out(1) != errType:
				return
			}
		}

		err = &InvalidTargetError{
			Name: name,
			Type: valueType,
		}
	}

	return
}

// Constructors holds information about symbols within a plugin that are to
// be used as fx.Provide functions.
type Constructors struct {
	// Plugin describe the constructor symbols to load.  Each element of this slice
	// must be either a string or an Annotated.
	//
	// If an element is a string, it is taken to be the name of a constructor.  That constructor
	// is loaded and added to the enclosing app with fx.Provide.
	//
	// If an element is an Annotated, then Annotated.Target is the name of a target and it will
	// be added to the enclosing fx.App with fx.Provide(fx.Annotated{...}) using the Name and Group
	// fields.
	//
	// Any other type will shortcircuit application startup with an error.
	Plugin []interface{}

	// IgnoreMissing defines what happens when an element from the Plugin field is not found.
	// If this field is true, missing symbols are silently ignored.  If this field is false (unset),
	// then any missing symbol will shortcircuit application startup with one or more errors.
	IgnoreMissing bool
}

func (ctors Constructors) Provide(s Plugin) fx.Option {
	var options []fx.Option
	for _, v := range ctors.Plugin {
		var (
			f   reflect.Value
			err error
		)

		switch ctor := v.(type) {
		case string:
			f, err = LookupConstructor(s, ctor)
			if err == nil {
				options = append(options, fx.Provide(f.Interface()))
			}

		case Annotated:
			f, err = LookupTarget(s, ctor.Target)
			if err == nil {
				options = append(options, fx.Provide(
					fx.Annotated{
						Name:   ctor.Name,
						Group:  ctor.Group,
						Target: f.Interface(),
					},
				))
			}

		default:
			err = fmt.Errorf("%T is not a valid type for Constructor.Plugin", v)
		}

		missing := IsMissingSymbolError(err)
		if (missing && !ctors.IgnoreMissing) || (!missing && err != nil) {
			options = append(options, fx.Error(err))
		}
	}

	return fx.Options(options...)
}
