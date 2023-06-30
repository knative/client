package plugin

// Allow plugins to register to this slice for inlining
var InternalPlugins PluginList

// Interface describing a plugin
type Plugin interface {
	// Get the name of the plugin (the file name without extensions)
	Name() string

	// Execute the plugin with the given arguments
	Execute(args []string) error

	// Return a description of the plugin (if support by the plugin binary)
	Description() (string, error)

	// The command path leading to this plugin.
	// Eg. for a plugin "kn source github" this will be [ "source", "github" ]
	CommandParts() []string

	// Location of the plugin where it is stored in the filesystem
	Path() string
}

// Used for sorting a list of plugins
type PluginList []Plugin

func (p PluginList) Len() int           { return len(p) }
func (p PluginList) Less(i, j int) bool { return p[i].Name() < p[j].Name() }
func (p PluginList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
