package installer

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimmykuang/selfskill/internal/frontmatter"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/skill"
)

// InstallResult represents the outcome of installing one item.
type InstallResult struct {
	Name    string
	Path    string
	Action  string // "installed", "skipped", "error"
	Detail  string
}

// Installer handles downloading and saving skills/prompts.
type Installer struct {
	httpClient  *http.Client
	skillStore  *skill.Store
	promptStore *prompt.Store
}

func New(ss *skill.Store, ps *prompt.Store) *Installer {
	return &Installer{
		httpClient:  &http.Client{},
		skillStore:  ss,
		promptStore: ps,
	}
}

// resolveSource determines the source type, reusing the installer's HTTP client.
func (i *Installer) resolveSource(input string) Source {
	return resolveSourceWithClient(input, i.httpClient)
}

// InstallSkill installs one or more skills from the given source.
func (i *Installer) InstallSkill(source string, force bool) ([]InstallResult, error) {
	src := i.resolveSource(source)
	results, err := src.Fetch()
	if err != nil {
		return nil, fmt.Errorf("fetching source: %w", err)
	}

	var installResults []InstallResult
	for _, fr := range results {
		r := i.installOneSkill(fr, force)
		installResults = append(installResults, r)
	}
	return installResults, nil
}

// InstallPrompt installs a prompt from the given source (single file only).
func (i *Installer) InstallPrompt(source string, force bool) ([]InstallResult, error) {
	// Prompts are always single files
	var content []byte
	var name string
	var err error

	if isURL(source) {
		content, err = httpGet(i.httpClient, resolveGitHubURL(source))
		if err != nil {
			return nil, err
		}
		name = filenameFromURL(source)
	} else {
		content, err = os.ReadFile(source)
		if err != nil {
			return nil, err
		}
		name = strings.TrimSuffix(filepath.Base(source), ".md")
	}

	// Parse to get ID
	meta, body, err := frontmatter.Parse(content)
	if err != nil {
		return nil, err
	}
	id := name
	if meta != nil {
		if v, ok := meta["id"].(string); ok && v != "" {
			id = v
		}
	}

	if i.promptStore.Exists(id) && !force {
		return []InstallResult{{Name: id, Action: "error", Detail: "already exists (use --force to overwrite)"}}, nil
	}

	var tags []string
	if meta != nil {
		if v, ok := meta["tags"].([]interface{}); ok {
			for _, t := range v {
				if ts, ok := t.(string); ok {
					tags = append(tags, ts)
				}
			}
		}
	}

	desc := ""
	if meta != nil {
		if v, ok := meta["description"].(string); ok {
			desc = v
		}
	}

	p := &prompt.Prompt{
		ID:          id,
		Tags:        tags,
		Description: desc,
		Body:        body,
	}
	if err := i.promptStore.Save(p); err != nil {
		return nil, err
	}

	return []InstallResult{{Name: id, Path: filepath.Join("prompts", id+".md"), Action: "installed"}}, nil
}

func (i *Installer) installOneSkill(fr FetchResult, force bool) InstallResult {
	exists := i.skillStore.Exists(fr.Name)
	if exists && !force {
		return InstallResult{Name: fr.Name, Action: "error", Detail: "already exists (use --force to overwrite)"}
	}
	if exists {
		i.skillStore.Remove(fr.Name)
	}

	if fr.IsDir {
		if err := i.skillStore.Save(fr.Name, fr.DirPath); err != nil {
			return InstallResult{Name: fr.Name, Action: "error", Detail: err.Error()}
		}
	} else {
		if err := i.skillStore.SaveFromContent(fr.Name, fr.Content); err != nil {
			return InstallResult{Name: fr.Name, Action: "error", Detail: err.Error()}
		}
	}

	return InstallResult{
		Name:   fr.Name,
		Path:   i.skillStore.SkillDir(fr.Name),
		Action: "installed",
	}
}
