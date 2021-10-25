package pluginfx

import (
	"fmt"
	"reflect"

	"go.uber.org/fx"
)

// errType is the "cached" reflection type for error.
var errType = reflect.TypeOf((*error)(nil)).Elem()

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

// Annotated is an analog of fx.Annotated for plugin symbols.  This type
// gives more control over how a plugin constructor gets placed into
// the enclosing fx.App.
type Annotated struct {
	// Name is the optional name of the component emitted by the Constructor.lue
	// Either Name or Group must be set, or an error is raised.
	Name string

	// Group is the value group for the component emitted by the Constructor.
	// Either Name or Group must be set, or an error is raised.
	Group string

	// Target is the name of a function symbol that must be legal to
	// use with fx.Annotated.Target.
	Target string
}

// Symbols describes how to bootstrap a set of symbols within an enclosing
// fx.App.
type Symbols struct {
	// Names are the symbol names to load into the enclosing fx.App.  Each
	// element of this slice must be either a string or an Annotated.
	//
	// Each symbol must refer to a function, or an error is raised.
	//
	// If an element is a string, it may be either a constructor or an invoke function.
	// If the function returns nothing or an error, it is wrapped in fx.Invoke.  Otherwise,
	// it is passed to fx.Provide.
	//
	// If an element is an Annotated, then the Target field is used to load a constructor.
	// This target constructor must return exactly (1) non-error value along with an optional
	// error.
	Names []interface{}

	// IgnoreMissing controls what happens when a symbol is not found in a plugin.
	// If this field is true, then missing symbols are silently ignored.  Otherwise,
	// a missing symbol will shortcircuit application startup with an error.
	IgnoreMissing bool
}

func (s Symbols) lookupFunc(p Plugin, n string, o []fx.Option) (reflect.Value, []fx.Option) {
	symbol, err := Lookup(p, n)
	if IsMissingSymbolError(err) {
		if !s.IgnoreMissing {
			o = append(o, fx.Error(err))
		}

		return reflect.Value{}, o
	}

	sv := reflect.ValueOf(symbol)
	if sv.Kind() != reflect.Func {
		return reflect.Value{},
			append(o, fx.Error(
				fmt.Errorf("Symbol %s is not a function", n),
			))
	}

	return sv, o
}

func (s Symbols) constructorOrInvoke(v reflect.Value, o []fx.Option) []fx.Option {
	vt := v.Type()
	for i := 0; i < vt.NumOut(); i++ {
		if vt.Out(i) != errType {
			// any non-error type means it's a constructor
			return append(o, fx.Provide(v.Interface()))
		}
	}

	return append(o, fx.Invoke(v.Interface()))
}

func (s Symbols) target(a Annotated, v reflect.Value, o []fx.Option) []fx.Option {
	vt := v.Type()
	switch {
	case vt.NumOut() < 1 || vt.NumOut() > 3:
		fallthrough

	case vt.NumOut() == 1 && vt.Out(0) == errType:
		fallthrough

	case vt.NumOut() == 2 && vt.Out(0) == errType && vt.Out(1) == errType:
		fallthrough

	case vt.NumOut() == 2 && vt.Out(0) != errType && vt.Out(1) != errType:
		return append(o, fx.Error(
			&InvalidTargetError{
				Name: a.Target,
				Type: vt,
			},
		))
	}

	return append(o, fx.Provide(
		fx.Annotated{
			Name:   a.Name,
			Group:  a.Group,
			Target: v.Interface(),
		},
	))
}

func (s Symbols) Load(p Plugin) fx.Option {
	options := make([]fx.Option, 0, len(s.Names))
	for _, n := range s.Names {
		var v reflect.Value
		switch name := n.(type) {
		case string:
			v, options = s.lookupFunc(p, name, options)
			if v.IsValid() {
				options = s.constructorOrInvoke(v, options)
			}

		case Annotated:
			v, options = s.lookupFunc(p, name.Target, options)
			if v.IsValid() {
				options = s.target(name, v, options)
			}

		default:
			options = append(options, fx.Error(
				fmt.Errorf("%T is not valid for Symbols.Names", n),
			))
		}
	}

	return fx.Options(options...)
}
