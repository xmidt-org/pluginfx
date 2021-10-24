package pluginfx

import (
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

	// Path is the plugin's path.  This field is required.
	Path string

	// Constructors are the optional exported functions from the plugin that participate
	// in dependency injection.  Each constructor is passed to fx.Provide.
	//
	// Each element of this slice must either by a string or an Annotated.  If a string,
	// it is the name of a function within the plugin.  If an Annotated, the Annotated.Constructor
	// field is the symbol and the Name and Group fields give control over how
	// the constructor's product is placed into the enclosing fx.App.
	Constructors Constructors

	// Lifecycle is the optional binding from a plugin's symbols to the enclosing
	// application.
	Lifecycle Lifecycle
}

// Provide builds the appropriate options to integrate this plugin into an
// enclosing fx.App.
//
// Typical usage:
//
//   app := fx.New(
//     pluginx.P{
//       Path: "/etc/lib/something.so",
//       Constructors: pluginfx.Constructors {
//       },
//       OnStart: "Initialize",
//       /* other fields filled out as desired */
//     }.Provide()
//   )
func (p P) Provide() fx.Option {
	var options []fx.Option
	plugin, err := Open(p.Path)

	if err == nil {
		options = append(options, p.Constructors.Provide(plugin))
		options = append(options, p.Lifecycle.Provide(plugin))
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

	// Paths are the plugin paths to load.
	Paths []string

	Constructors Constructors

	Lifecycle Lifecycle
}

// Provide opens a list of plugins described in the Paths field.  These plugins are optionally
// put into a value group if the Group field is set.  Each plugin is then examined for symbols
// to provide to the enclosing fx.App in a manner similar to Plugin.Provide.
func (s S) Provide() fx.Option {
	var options []fx.Option
	for _, path := range s.Paths {
		options = append(options,
			P{
				Group:     s.Group,
				Anonymous: len(s.Group) == 0,
				Path:      path,

				Constructors: s.Constructors,
				Lifecycle:    s.Lifecycle,
			}.Provide(),
		)
	}

	return fx.Options(options...)
}
