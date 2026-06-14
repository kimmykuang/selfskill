package cc_compat

// FakeRegistry is an in-memory Registry for tests.
type FakeRegistry struct {
	entries map[string][]Entry
}

func NewFakeRegistry() *FakeRegistry {
	return &FakeRegistry{entries: map[string][]Entry{}}
}

func (f *FakeRegistry) List() (map[string][]Entry, error) {
	out := make(map[string][]Entry, len(f.entries))
	for k, v := range f.entries {
		cp := make([]Entry, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out, nil
}

func (f *FakeRegistry) Add(name string, entry Entry) error {
	if entry.InstalledAt == "" {
		if existing, ok := f.entries[name]; ok && len(existing) > 0 && existing[0].InstalledAt != "" {
			entry.InstalledAt = existing[0].InstalledAt
		} else {
			entry.InstalledAt = entry.LastUpdated
		}
	}
	f.entries[name] = []Entry{entry}
	return nil
}

func (f *FakeRegistry) Remove(name string) error {
	delete(f.entries, name)
	return nil
}
