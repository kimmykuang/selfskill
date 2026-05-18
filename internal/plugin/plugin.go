package plugin

// Plugin represents a plugin managed by ss in ~/ss/plugins/.
type Plugin struct {
	Name        string // e.g. "superpowers"
	Marketplace string // e.g. "claude-plugins-official"
	Version     string // e.g. "5.1.0" or git commit sha
	InstallPath string // ~/ss/plugins/<marketplace>/<plugin>/<version>/
	GitCommitSha string
}

// FullName returns the plugin key as used in installed_plugins.json.
func (p *Plugin) FullName() string {
	return p.Name + "@" + p.Marketplace
}

// LoadState represents whether a plugin is loaded into CC.
type LoadState string

const (
	StateInstalled LoadState = "installed"
	StateLoaded    LoadState = "loaded"
)

// PluginInfo combines plugin metadata with its current load state.
type PluginInfo struct {
	Plugin
	State LoadState
}
