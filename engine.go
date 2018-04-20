package engine

import (
	"log"
	"path"

	"github.com/lagoon-platform/model"
)

type Lagoon interface {
	Environment() model.Environment
	ComponentManager() ComponentManager
}

type context struct {
	logger *log.Logger

	// Subsystems
	componentManager ComponentManager

	// Environment
	baseDir     string
	environment model.Environment
}

// Create creates an environment descriptor based on the provider location.
//
// The location can be an URL over http or https or even a file system location.
func Create(logger *log.Logger, baseDir string, repository string, version string) (lagoon Lagoon, err error) {
	ctx := context{logger: logger, baseDir: baseDir}

	// Create component manager
	ctx.componentManager = createComponentManager(logger, baseDir)

	// Create, register and fetch the main component
	mainComp, err := model.CreateDetachedComponent(repository, version)
	if err != nil {
		return
	}
	ctx.componentManager.RegisterComponent(mainComp)
	ctx.componentManager.Ensure()

	// Parse the environment descriptor from the main component
	envPath, err := ctx.componentManager.ComponentPath(mainComp.Id)
	if err != nil {
		return
	}
	ctx.environment, err = model.Parse(logger, path.Join(envPath, DescriptorFileName))
	if err != nil {
		return
	}

	// Use context as Lagoon facade
	lagoon = &ctx

	return
}

func (c *context) Environment() model.Environment {
	return c.environment
}

func (c *context) ComponentManager() ComponentManager {
	return c.componentManager
}

func CreateLagoon(logger *log.Logger, l Lagoon) {

}

func UpdateLagoon(logger *log.Logger, l Lagoon) {

}
