function app() {
    return {
        route: 'prompts',
        prompts: [],
        skills: [],
        plugins: [],
        status: { userLinks: [], projectLinks: [] },
        searchQuery: '',
        searchTags: '',
        editPrompt: { ID: '', Description: '', Body: '', Tags: [], tagsStr: '', isNew: true },

        init() {
            this.handleRoute();
            window.addEventListener('hashchange', () => this.handleRoute());
        },

        handleRoute() {
            const hash = window.location.hash.slice(2) || 'prompts';

            if (hash === 'prompts') {
                this.route = 'prompts';
                this.fetchPrompts();
            } else if (hash === 'prompts/new') {
                this.route = 'prompts/new';
                this.editPrompt = { ID: '', Description: '', Body: '', Tags: [], tagsStr: '', isNew: true };
            } else if (hash.startsWith('prompts/')) {
                this.route = 'prompts/edit';
                this.loadPrompt(hash.replace('prompts/', ''));
            } else if (hash === 'skills') {
                this.route = 'skills';
                this.fetchSkills();
            } else if (hash === 'plugins') {
                this.route = 'plugins';
                this.fetchPlugins();
            } else if (hash === 'status') {
                this.route = 'status';
                this.fetchStatus();
            }
        },

        async fetchPrompts() {
            const params = new URLSearchParams();
            if (this.searchQuery) params.set('q', this.searchQuery);
            if (this.searchTags) params.set('tags', this.searchTags);
            const url = '/api/prompts' + (params.toString() ? '?' + params.toString() : '');
            const resp = await fetch(url);
            this.prompts = await resp.json();
        },

        async loadPrompt(id) {
            const resp = await fetch('/api/prompts/' + encodeURIComponent(id));
            if (!resp.ok) {
                window.location.hash = '#/prompts';
                return;
            }
            const p = await resp.json();
            this.editPrompt = {
                ID: p.ID,
                Description: p.Description || '',
                Body: p.Body || '',
                Tags: p.Tags || [],
                tagsStr: (p.Tags || []).join(', '),
                isNew: false,
            };
        },

        async savePrompt() {
            const tags = this.editPrompt.tagsStr
                .split(',')
                .map(t => t.trim())
                .filter(t => t.length > 0);

            const body = {
                ID: this.editPrompt.ID,
                Description: this.editPrompt.Description,
                Body: this.editPrompt.Body,
                Tags: tags,
            };

            const method = this.editPrompt.isNew ? 'POST' : 'PUT';
            const url = this.editPrompt.isNew
                ? '/api/prompts'
                : '/api/prompts/' + encodeURIComponent(this.editPrompt.ID);

            const resp = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
            });

            if (resp.ok) {
                window.location.hash = '#/prompts';
            } else {
                const err = await resp.json();
                alert(err.error || 'Failed to save prompt');
            }
        },

        async deletePrompt() {
            if (!confirm('Delete prompt "' + this.editPrompt.ID + '"?')) return;
            const resp = await fetch('/api/prompts/' + encodeURIComponent(this.editPrompt.ID), {
                method: 'DELETE',
            });
            if (resp.ok) {
                window.location.hash = '#/prompts';
            }
        },

        async fetchSkills() {
            const resp = await fetch('/api/skills');
            this.skills = await resp.json();
        },

        async fetchPlugins() {
            const resp = await fetch('/api/plugins');
            const plugins = await resp.json();
            // Fetch skills for each plugin in parallel
            await Promise.all(plugins.map(async (p) => {
                try {
                    const skResp = await fetch('/api/plugins/' + encodeURIComponent(p.name) + '/skills');
                    p.skills = skResp.ok ? await skResp.json() : [];
                } catch {
                    p.skills = [];
                }
            }));
            this.plugins = plugins;
        },

        async loadPlugin(name) {
            const resp = await fetch('/api/plugins/' + encodeURIComponent(name) + '/load', { method: 'POST' });
            if (resp.ok) {
                this.fetchPlugins();
            } else {
                const err = await resp.json();
                alert(err.error || 'Failed to load plugin');
            }
        },

        async unloadPlugin(name) {
            const resp = await fetch('/api/plugins/' + encodeURIComponent(name) + '/unload', { method: 'POST' });
            if (resp.ok) {
                this.fetchPlugins();
            } else {
                const err = await resp.json();
                alert(err.error || 'Failed to unload plugin');
            }
        },

        async fetchStatus() {
            const resp = await fetch('/api/status');
            this.status = await resp.json();
        },
    };
}
