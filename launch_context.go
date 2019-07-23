package engine

import (
	"log"

	"github.com/ekara-platform/engine/ansible"
	"github.com/ekara-platform/engine/util"
	"github.com/ekara-platform/model"
)

type (
	//LaunchContext Represents the informations required to run the engine
	LaunchContext interface {
		//Name The environment descriptor name to process
		Name() string
		//User The user to log into the descriptor repository
		User() string
		//Password The user to log into the descriptor repository
		Password() string
		//Location specifies where to look for the environment descriptor
		Location() string
		//Log the looger to used during the ekara execution
		Log() *log.Logger
		//Ef the exchange folder where to find informations required
		// to complete the Ekara execution of where to write informations
		// produced by the execution
		Ef() *util.ExchangeFolder
		//Ekara the engine in charge of the process
		Ekara() Engine
		//QualifiedName The qualifie name of the environment to process
		QualifiedName() string
		//HTTPProxy The http proxy used by the engine during the process execution
		HTTPProxy() string
		//HTTPSProxy The https proxy used by the engine during the process execution
		HTTPSProxy() string
		//NoProxy The no proxy configuration used by the engine during the process execution
		NoProxy() string
		//SSHPublicKey the public key used by the engine during the process execution to
		// connect the created nodes
		SSHPublicKey() string
		//SSHPrivateKey the private key used by the engine during the process execution to
		// connect the created nodes
		SSHPrivateKey() string
		//TemplateContext the parameters provided by the user to fill the environment descriptor as a template
		TemplateContext() *model.TemplateContext
		//Error if any, the error which has broken the process
		Error() error
	}
)

//BuildBaseParam Returns the base skeleton of the parameters for a node set
func BuildBaseParam(c LaunchContext, nodeSetName string, provider string) ansible.BaseParam {
	return ansible.BuildBaseParam(*c.Ekara().ComponentManager().Environment(), nodeSetName, provider, c.SSHPublicKey(), c.SSHPrivateKey())
}
