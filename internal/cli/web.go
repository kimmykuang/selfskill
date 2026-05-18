package cli

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/skill"
	"github.com/kimmykuang/selfskill/internal/web"
	"github.com/spf13/cobra"
)

func newWebCmd(skillStore *skill.Store, promptStore *prompt.Store, pluginStore *plugin.Store, pluginLoader *plugin.Loader, lnk *linker.Linker, staticFS fs.FS) *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "web",
		Short: "Start the web management UI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stdout, "Starting web server on http://localhost:%d\n", port)
			srv := &web.Server{
				SkillStore:   skillStore,
				PromptStore:  promptStore,
				PluginStore:  pluginStore,
				PluginLoader: pluginLoader,
				Linker:       lnk,
				StaticFS:     staticFS,
				Port:         port,
			}
			return srv.Start()
		},
	}

	cmd.Flags().IntVar(&port, "port", 8484, "Port to listen on")
	return cmd
}
