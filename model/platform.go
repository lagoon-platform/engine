package model

import (
	"fmt"
)

//Platform the platform used to build an environment
type Platform struct {
	Self       component
	Parents    []component
	Components map[string]component
}

func createPlatform(from component, yamlEkara yamlEkara) (Platform, error) {
	p := Platform{
		Self:       from,
		Components: make(map[string]component),
	}

	// Register the component itself
	err := p.registerComponent(p.Self)
	if err != nil {
		return Platform{}, err
	}

	// Build and register declared components
	for id, yamlC := range yamlEkara.Components {
		c, err := from.buildComponent(yamlEkara.Base, id, yamlC)
		if err != nil {
			return Platform{}, err
		}
		err = p.registerComponent(c)
		if err != nil {
			return Platform{}, err
		}
	}

	// Store local info into the right component
	c, ok := p.Components[from.ComponentId()]
	if ok {
		c.Templates = yamlEkara.Templates
		c.Playbooks = yamlEkara.Playbooks
	} else {
		return Platform{}, fmt.Errorf("missing component %s", from.ComponentId())
	}
	p.Components[from.ComponentId()] = c

	return p, nil
}

func (p *Platform) registerComponent(c component) error {
	if _, ok := p.Components[c.id]; ok {
		existing := p.Components[c.id]
		existing.merge(c)
		p.Components[c.id] = existing
	} else {
		p.Components[c.id] = c
	}
	return nil
}

func (p *Platform) merge(with Platform) {
	// Create component map if necessary
	if p.Components == nil {
		p.Components = make(map[string]component)
	}

	// Old self is added to parents before its identity is updated
	p.Parents = append([]component{p.Self}, p.Parents...)
	p.Self = with.Self

	// Merge components
	if with.Components != nil {
		for _, c := range with.Components {
			existing := p.Components[c.id]
			existing.merge(c)
			p.Components[c.id] = existing
		}
	}
}
