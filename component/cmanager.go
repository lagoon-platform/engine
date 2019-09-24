package component

import (
	"log"
	"os"

	"path/filepath"

	"github.com/ekara-platform/engine/component/scm"
	"github.com/ekara-platform/model"
)

var releaseNothing = func() {
	// Doing nothing and it's fine
}

type (
	// ComponentManager downloads and keep track of ekara components on disk.
	ComponentManager struct {
		Logger      *log.Logger
		environment *model.Environment
		Directory   string
		Paths       map[string]scm.FetchedComponent
	}

	localRef struct {
		component model.Component
	}

	// FileMap is used to Marshal the map of downloaded components
	fileMap struct {
		File map[string]string `yaml:"component_path"`
	}
)

//CreateComponentManager creates a new component manager
func CreateComponentManager(logger *log.Logger, baseDir string) *ComponentManager {
	c := &ComponentManager{
		Logger:      logger,
		environment: nil,
		Directory:   filepath.Join(baseDir, "components"),
		Paths:       map[string]scm.FetchedComponent{},
	}
	c.environment = model.InitEnvironment()
	return c
}

func (cm *ComponentManager) isComponentFetched(id string) (val scm.FetchedComponent, present bool) {
	val, present = cm.Paths[id]
	return
}

func (cm *ComponentManager) EnsureOneComponent(c model.Component, data *model.TemplateContext) error {
	cm.Logger.Printf("ensuring component: %s", c.Id)
	path, fetched := cm.isComponentFetched(c.Id)
	if !fetched {
		fComp, err := fetch(cm, c)
		if err != nil {
			cm.Logger.Printf("error fetching the component %s", err.Error())
			return err
		}
		path = fComp
	}
	if path.HasDescriptor() {
		cm.Logger.Printf("creating partial environment based on component %s", c.Id)
		descriptorYaml, err := model.ParseYamlDescriptor(path.DescriptorUrl, data)
		if err != nil {
			cm.Logger.Printf("error parsing the descriptor %s", err.Error())
			return err
		}

		cEnv, err := model.CreateEnvironment(path.DescriptorUrl.String(), descriptorYaml, c.Id)
		if err != nil {
			return err
		}

		// Customize or keep the resulting environment into the global one
		cm.Logger.Println("prepare partial environment customization")
		if cm.environment == nil {
			cm.environment = cEnv
			cm.Logger.Println("no customization required, it's the first built environment ")
		} else {
			// We don't want to customization the templates defined into the environment
			// But instead we want to keep them into the component
			cm.environment.Platform().KeepTemplates(c, cEnv.Templates)
			cEnv.Templates = model.Patterns{}
			cm.Logger.Println("partial environment should used for customization")
			err = cm.environment.Customize(cEnv)

			if err != nil {
				cm.Logger.Printf("error customizing the environment %s", err.Error())
				return err
			}
		}
	}
	data.Model = model.CreateTEnvironmentForEnvironment(*cm.environment)

	return nil
}

func (cm *ComponentManager) Environment() *model.Environment {
	return cm.environment
}

func (cm *ComponentManager) ContainsFile(name string, data *model.TemplateContext, in ...model.ComponentReferencer) MatchingPaths {
	return cm.contains(false, name, data, in...)
}

func (cm *ComponentManager) ContainsDirectory(name string, data *model.TemplateContext, in ...model.ComponentReferencer) MatchingPaths {
	return cm.contains(true, name, data, in...)
}

func (cm *ComponentManager) contains(isFolder bool, name string, data *model.TemplateContext, in ...model.ComponentReferencer) MatchingPaths {
	res := MatchingPaths{
		Paths: make([]MatchingPath, 0, 0),
	}
	if len(in) > 0 {
		for _, v := range in {
			uv, err := cm.Use(v, data)
			if err != nil {
				cm.Logger.Printf("An error occurred using the component %s : %s", v.ComponentName(), err.Error())
			}
			if isFolder {
				if ok, match := uv.ContainsDirectory(name); ok {
					res.Paths = append(res.Paths, match)
				} else {
					uv.Release()
				}
			} else {
				if ok, match := uv.ContainsFile(name); ok {
					res.Paths = append(res.Paths, match)
				} else {
					uv.Release()
				}
			}
		}
	} else {
		for _, comp := range cm.environment.Platform().Components {
			lRef := localRef{
				component: comp,
			}
			uv, err := cm.Use(lRef, data)
			if err != nil {
				cm.Logger.Printf("An error occurred using the component %s : %s", lRef.ComponentName(), err.Error())
			}
			if isFolder {
				if ok, match := uv.ContainsDirectory(name); ok {
					res.Paths = append(res.Paths, match)
				} else {
					uv.Release()
				}
			} else {
				if ok, match := uv.ContainsFile(name); ok {
					res.Paths = append(res.Paths, match)
				} else {
					uv.Release()
				}
			}
		}
	}
	return res
}

//Use returns a UsableComponent matching the given reference.
//If the component corresponding to the reference contains a template
//definition then the component will be duplicated and templated before
// being returned as a UsableComponent.
// Don't forget to Release the UsableComponent once is processing is over...
func (cm *ComponentManager) Use(cr model.ComponentReferencer, data *model.TemplateContext) (UsableComponent, error) {
	c := cm.environment.Platform().Components[cr.ComponentName()]
	if ok, patterns := c.Templatable(); ok {
		path, err := runTemplate(*data, cm.Paths[cr.ComponentName()].LocalPath, patterns, cr)
		if err != nil {
			return usable{}, err
		}
		// No error no path then it has not been templated
		if err == nil && path == "" {
			goto TemplateFalse
		}
		return usable{
			cm:        cm,
			path:      path,
			release:   cleanup(path),
			component: cm.environment.Platform().Components[cr.ComponentName()],
			templated: true,
		}, nil
	}
TemplateFalse:
	return usable{
		cm:        cm,
		release:   releaseNothing,
		path:      filepath.Join(cm.Directory, cr.ComponentName()),
		component: cm.environment.Platform().Components[cr.ComponentName()],
		templated: false,
	}, nil
}

func cleanup(path string) func() {
	return func() {
		os.RemoveAll(path)
	}
}

//Component returns the referenced component
func (r localRef) Component() (model.Component, error) {
	return r.component, nil
}

//ComponentName returns the referenced component name
func (r localRef) ComponentName() string {
	return r.component.Id
}
