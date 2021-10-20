package pluginfx

import (
	"plugin"
	"reflect"
)

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
