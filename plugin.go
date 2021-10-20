package pluginfx

import (
	"fmt"
	"plugin"
	"reflect"

	"go.uber.org/fx"
)

// Plugin describes how to load a single plugin and integrate it into
// an enclosing fx.App.
type Plugin struct {
	// Name is the optional name of the plugin component within the application.  This
	// field is ignored if Anonymous is set.
	Name string

	// Group is the optional value group to place the loaded plugin into.  This field
	// is ignored if Anonymous is set.
	Group string

	// Anonymous controls whether the plugin itself is provided as a component
	// to the enclosing fx.App.  If this field is true, then the plugin is not
	// placed into the fx.App regardless of the values of Name and Group.
	Anonymous bool

	// Path is the plugin's path.  This field is required.
	Path string

	// IgnoreMissingConstructors controls what happens if a symbol in the Constructors
	// field is not present in the plugin.  If this field is true, missing constructor
	// symbols are silently ignored.  Otherwise, missing constructor symbols will shortcircuit
	// application startup with one or more errors.
	IgnoreMissingConstructors bool

	// Constructors are the optional exported functions from the plugin that participate
	// in dependency injection.  Each constructor is passed to fx.Provide.
	Constructors []string

	// IgnoreMissingLifecycle controls what happens if OnStart or OnStop are not symbols
	// in the plugin.  If this field is true and either OnStart or OnStop are not present,
	// no error is raised.  Otherwise, application startup is shortcircuited with one or
	// more errors.
	IgnoreMissingLifecycle bool

	// OnStart is the optional exported function that should be called when the enclosing
	// application is started.
	OnStart string

	// OnStop is the optional exported function that should be called when the enclosing
	// application is stopped.
	OnStop string
}

func (pl Plugin) appendConstructors(s Symbols, options []fx.Option) []fx.Option {
	EachSymbol(
		s,
		func(name string, value reflect.Value, lookupErr error) error {
			if lookupErr != nil {
				if !pl.IgnoreMissingConstructors {
					options = append(options, fx.Error(
						fmt.Errorf("Unable to load constructor: %s", lookupErr),
					))
				}
			} else if err := checkConstructorSymbol(name, value); err != nil {
				options = append(options, fx.Error(err))
			} else {
				options = append(options, fx.Provide(value.Interface()))
			}

			return nil
		},
		pl.Constructors...,
	)

	return options
}

func (pl Plugin) appendLifecycle(s Symbols, options []fx.Option) []fx.Option {
	var hook fx.Hook
	if len(pl.OnStart) > 0 {
		var err error
		hook.OnStart, err = LookupLifecycle(s, pl.OnStart)
		if err != nil {
			options = append(options, fx.Error(err))
		}
	}

	if len(pl.OnStop) > 0 {
		var err error
		hook.OnStop, err = LookupLifecycle(s, pl.OnStop)
		if err != nil {
			options = append(options, fx.Error(err))
		}
	}

	if hook.OnStart != nil || hook.OnStop != nil {
		options = append(options, fx.Invoke(
			func(l fx.Lifecycle) { l.Append(hook) },
		))
	}

	return options
}

// Provide builds the appropriate options to integrate this plugin into an
// enclosing fx.App.
//
// Typical usage:
//
//   app := fx.New(
//     pluginx.Plugin{
//       Path: "/etc/lib/something.so",
//       Constructors: []string{"ProvideSomething", "NewSomethingElse"},
//       OnStart: "Initialize",
//       /* other fields filled out as desired */
//     }.Provide()
//   )
func (pl Plugin) Provide() fx.Option {
	var options []fx.Option
	p, err := plugin.Open(pl.Path)

	if err != nil {
		err = fmt.Errorf("Unable to load plugin from path %s: %s", pl.Path, err)
	} else {
		options = pl.appendConstructors(p, options)
		options = pl.appendLifecycle(p, options)
	}

	switch {
	case !pl.Anonymous && (len(pl.Name) > 0 || len(pl.Group) > 0):
		options = append(options, fx.Provide(
			fx.Annotated{
				Name:   pl.Name,
				Group:  pl.Group,
				Target: func() (Symbols, error) { return p, err },
			},
		))

	case !pl.Anonymous:
		options = append(options, fx.Provide(
			func() (Symbols, error) { return p, err },
		))

	case err != nil:
		// need to short-circuit startup, even though no component is created
		options = append(options,
			fx.Error(err),
		)
	}

	return fx.Options(options...)
}
