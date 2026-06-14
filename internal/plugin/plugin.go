package plugin

// Plugin represents a plugin managed by ss in ~/ss/plugins/.
type Plugin struct {
	Name         string // e.g. "superpowers"
	Marketplace  string // e.g. "claude-plugins-official"
	Version      string // e.g. "5.1.0" or git commit sha
	InstallPath  string // ~/ss/plugins/<marketplace>/<plugin>/<version>/
	GitCommitSha string
	RemoteURL    string // populated from .ss-meta.json when available
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

// PluginMeta is the on-disk record of how a plugin was installed.
// Persisted at <installPath>/.ss-meta.json.
type PluginMeta struct {
	URL         string `json:"url"`
	Marketplace string `json:"marketplace"`
	CommitSha   string `json:"commitSha"`
	Version     string `json:"version"`
	InstalledAt string `json:"installedAt"`
}
