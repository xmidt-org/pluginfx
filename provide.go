package pluginfx

import (
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
	//
	// Each element of this slice must either by a string or an Annotated.  If a string,
	// it is the name of a function within the plugin.  If an Annotated, the Annotated.Constructor
	// field is the symbol and the Name and Group fields give control over how
	// the constructor's product is placed into the enclosing fx.App.
	Constructors Constructors

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

func (pl Plugin) appendLifecycle(s Symbols, options []fx.Option) []fx.Option {
	var hook fx.Hook
	if len(pl.OnStart) > 0 {
		var err error
		hook.OnStart, err = LookupLifecycle(s, pl.OnStart)
		if !pl.IgnoreMissingLifecycle && err != nil {
			options = append(options, fx.Error(err))
		}
	}

	if len(pl.OnStop) > 0 {
		var err error
		hook.OnStop, err = LookupLifecycle(s, pl.OnStop)
		if !pl.IgnoreMissingLifecycle && err != nil {
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
//       Constructors: pluginfx.Constructors {
//       },
//       OnStart: "Initialize",
//       /* other fields filled out as desired */
//     }.Provide()
//   )
func (pl Plugin) Provide() fx.Option {
	var options []fx.Option
	symbols, err := Open(pl.Path)

	if err == nil {
		options = append(options, pl.Constructors.Provide(symbols))
		options = pl.appendLifecycle(symbols, options)
	}

	// emit the plugin as a component if desired, even when there's an error.
	// this lets the fx.App produce useful error messages.
	switch {
	case !pl.Anonymous && (len(pl.Name) > 0 || len(pl.Group) > 0):
		options = append(options, fx.Provide(
			fx.Annotated{
				Name:   pl.Name,
				Group:  pl.Group,
				Target: func() (Symbols, error) { return symbols, err },
			},
		))

	case !pl.Anonymous:
		options = append(options, fx.Provide(
			func() (Symbols, error) { return symbols, err },
		))

	case err != nil:
		// need to short-circuit startup, even though no component is created
		options = append(options,
			fx.Error(err),
		)
	}

	return fx.Options(options...)
}

// Set describes how to load multiple plugins as a bundle and integrate each of them
// into an enclosing fx.App.
type Set struct {
	// Group is the optional value group to place each plugin in this set into.  If this
	// field is unset, the loaded plugins are not added as components.
	Group string

	// Paths are the plugin paths to load.
	Paths []string

	// IgnoreMissingConstructors controls what happens if a symbol in the Constructors
	// field is not present in any of the plugins.  If this field is true, missing constructor
	// symbols are silently ignored.  Otherwise, missing constructor symbols will shortcircuit
	// application startup with one or more errors.
	IgnoreMissingConstructors bool

	// Constructors are the optional exported functions from each plugin that participate
	// in dependency injection.  Each constructor is passed to fx.Provide.
	Constructors Constructors

	// IgnoreMissingLifecycle controls what happens if OnStart or OnStop are not symbols
	// in each plugin.  If this field is true and either OnStart or OnStop are not present,
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

// Provide opens a list of plugins described in the Paths field.  These plugins are optionally
// put into a value group if the Group field is set.  Each plugin is then examined for symbols
// to provide to the enclosing fx.App in a manner similar to Plugin.Provide.
func (s Set) Provide() fx.Option {
	var options []fx.Option
	for _, path := range s.Paths {
		options = append(options,
			Plugin{
				Group:     s.Group,
				Anonymous: len(s.Group) == 0,
				Path:      path,

				IgnoreMissingConstructors: s.IgnoreMissingConstructors,
				Constructors:              s.Constructors,

				IgnoreMissingLifecycle: s.IgnoreMissingLifecycle,
				OnStart:                s.OnStart,
				OnStop:                 s.OnStop,
			}.Provide(),
		)
	}

	return fx.Options(options...)
}
