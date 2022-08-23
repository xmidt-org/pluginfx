package pluginfx

import (
	"os"
	"path/filepath"

	"go.uber.org/fx"
)

// P describes how to load a single plugin and integrate it into
// an enclosing fx.App.
type P struct {
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

	// Path is the plugin's path.  This field is required.  Variables are expanded
	// via os.ExpandEnv.
	Path string

	// Symbols describes the optional set of functions exported by the plugin to be
	// bound to the enclosing fx.App.  Both provide and invoke functions can be defined
	// using this field.
	Symbols Symbols

	// Lifecycle is the optional binding from a plugin's symbols to the enclosing
	// application.
	Lifecycle Lifecycle
}

// Provide builds the appropriate options to integrate this plugin into an
// enclosing fx.App.
//
// Typical usage:
//
//	app := fx.New(
//	  pluginx.P{
//	    Anonymous: true, // leave unset if you want the plugin accessible via DI
//	    Path: "/etc/lib/something.so",
//	    Symbols: pluginfx.Symbols {
//	      Names: []interface{}{
//	        "MyConstructor",
//	      },
//	    },
//	    Lifecycle: pluginfx.Lifecycle {
//	      OnStart: "Initialize",
//	    },
//	  }.Provide()
//	)
func (p P) Provide() fx.Option {
	var options []fx.Option
	plugin, err := Open(os.ExpandEnv(p.Path))

	if err == nil {
		options = append(options, p.Symbols.Load(plugin))
		options = append(options, p.Lifecycle.Bind(plugin))
	}

	// emit the plugin as a component if desired, even when there's an error.
	// this lets the fx.App produce useful error messages.
	switch {
	case !p.Anonymous && (len(p.Name) > 0 || len(p.Group) > 0):
		options = append(options, fx.Provide(
			fx.Annotated{
				Name:   p.Name,
				Group:  p.Group,
				Target: func() (Plugin, error) { return plugin, err },
			},
		))

	case !p.Anonymous:
		options = append(options, fx.Provide(
			func() (Plugin, error) { return plugin, err },
		))

	case err != nil:
		// need to short-circuit startup, even though no component is created
		options = append(options,
			fx.Error(err),
		)
	}

	return fx.Options(options...)
}

// S describes how to load multiple plugins as a bundle and integrate each of them
// into an enclosing fx.App.
type S struct {
	// Group is the optional value group to place each plugin in this set into.  If this
	// field is unset, the loaded plugins are not added as components.
	Group string

	// Paths are the plugin paths to load.  Each of these paths may be a filesystem glob,
	// in which case all matching files are loaded as plugins.  Variable expansion is also
	// done on each element via os.ExpandEnv.
	Paths []string

	// Symbols are the symbols to be loaded from each loaded plugin.
	Symbols Symbols

	// Lifecycle describes the symbols from each loaded plugin to be bound to the
	// enclosing application.
	Lifecycle Lifecycle
}

// Provide opens a list of plugins described in the Paths field.  These plugins are optionally
// put into a value group if the Group field is set.  Each plugin is then examined for symbols
// to provide to the enclosing fx.App in a manner similar to Plugin.Provide.
func (s S) Provide() fx.Option {
	var options []fx.Option
	for _, path := range s.Paths {
		matches, err := filepath.Glob(os.ExpandEnv(path))
		if err != nil {
			options = append(options, fx.Error(err))
			continue
		}

		for _, match := range matches {
			options = append(options,
				P{
					Group:     s.Group,
					Anonymous: len(s.Group) == 0,
					Path:      match,

					Symbols:   s.Symbols,
					Lifecycle: s.Lifecycle,
				}.Provide(),
			)
		}
	}

	return fx.Options(options...)
}
