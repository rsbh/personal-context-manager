package store

// Resolve returns the contexts that apply to a project, with content:
// global contexts first, then project contexts. A project context replaces
// a global context with the same section and name. Empty project means
// global only. A project that does not exist gives global only.
func (s *Store) Resolve(project, section string) ([]Context, error) {
	if project != "" {
		if err := validName("project", project); err != nil {
			return nil, err
		}
	}
	if section != "" {
		if err := validName("section", section); err != nil {
			return nil, err
		}
	}

	global, err := s.List(Filter{Scope: ScopeGlobal, Section: section})
	if err != nil {
		return nil, err
	}
	var local []Context
	if project != "" {
		if local, err = s.List(Filter{Scope: ScopeProject, Project: project, Section: section}); err != nil {
			return nil, err
		}
	}

	key := func(c Context) string { return c.Section + "/" + c.Name }
	replaced := make(map[string]bool, len(local))
	for _, c := range local {
		replaced[key(c)] = true
	}
	out := make([]Context, 0, len(global)+len(local))
	for _, c := range global {
		if !replaced[key(c)] {
			out = append(out, c)
		}
	}
	out = append(out, local...)

	for i := range out {
		if out[i], err = s.Get(out[i].ID); err != nil {
			return nil, err
		}
	}
	return out, nil
}
