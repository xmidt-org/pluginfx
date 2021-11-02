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
	if !vv.IsValid() || (vv.Kind() == reflect.Ptr && vv.IsNil()) {
		panic("pluginfx.SymbolMap: A symbol must be either a function or a non-nil pointer")
	}

	if vv.Kind() != reflect.Func && vv.Kind() != reflect.Ptr {
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
func (sm *SymbolMap) Lookup(name string) (plugin.Symbol, error) {
	if s, ok := sm.symbols[name]; ok {
		return s, nil
	} else {
		// mimic the error returned by the plugin package
		return nil, &MissingSymbolError{Name: name}
	}
}

// NewSymbolMap shallow copies the contents of a map onto a new
// SymbolMap instance.  Each symbol value is handled with Set.
func NewSymbolMap(m map[string]interface{}) *SymbolMap {
	sm := &SymbolMap{
		symbols: make(map[string]plugin.Symbol, len(m)),
	}

	for k, v := range m {
		sm.Set(k, v)
	}

	return sm
}

// NewSymbols is like NewSymbolMap, but it uses a sequence of name/value
// pairs.  This function is often easier to use than NewSymbolMap, due to
// the noisiness of declaring a map.
//
// This function panics if namesAndValues is not empty and does not have
// an even number of elements.  It also panics if any even-numbered element (including zero)
// is not a string.
func NewSymbols(namesAndValues ...interface{}) *SymbolMap {
	count := len(namesAndValues)
	if count%2 != 0 {
		panic("NewSymbols: an even number of elements is required")
	}

	sm := &SymbolMap{
		symbols: make(map[string]plugin.Symbol, count/2),
	}

	for i, j := 0, 1; i < count; i, j = i+2, j+2 {
		name := namesAndValues[i].(string)
		sm.Set(name, namesAndValues[j])
	}

	return sm
}
