package pluginfx

import (
	"plugin"
	"reflect"
)

// SymbolMap is a map implementation of Symbols.  It allows for an in-memory
// implementation of a plugin for testing or for production defaults.
//
// The zero value of this type is a usable, empty "plugin".  An existing
// map may be copied into a new *SymbolMap by using NewSymbolMap.
type SymbolMap struct {
	symbols map[string]plugin.Symbol
}

// Set adds a symbol to this map.  The value may not be a nil pointer,
// or this method panics.
//
// If value is a function or a non-nil pointer, it is added to this map as is.
// Otherwise, a pointer is created that points to value, and that pointer is added
// to this map.
func (sm *SymbolMap) Set(name string, value interface{}) {
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

	if sm.symbols == nil {
		sm.symbols = make(map[string]plugin.Symbol)
	}

	sm.symbols[name] = value
}

// Del removes a symbol from this map.
func (sm *SymbolMap) Del(name string) {
	if sm.symbols != nil {
		delete(sm.symbols, name)
	}
}

// Lookup implements the Symbols interface.
func (sm SymbolMap) Lookup(name string) (plugin.Symbol, error) {
	if s, ok := sm.symbols[name]; ok {
		return s, nil
	} else {
		// mimic the error returned by the plugin package
		return nil, &MissingSymbolError{Name: name}
	}
}

func NewSymbolMap(m map[string]interface{}) *SymbolMap {
	sm := &SymbolMap{
		symbols: make(map[string]plugin.Symbol, len(m)),
	}

	for k, v := range m {
		sm.Set(k, v)
	}

	return sm
}
