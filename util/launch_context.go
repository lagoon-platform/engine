package util

import (
	"github.com/ekara-platform/engine/model"
	"log"
)

type (
	//LaunchContext Represents the information required to run the engine
	LaunchContext interface {
		//Feedback is used to notify progress to the end-user.
		Feedback() FeedbackNotifier
		//Skip is the requested level of skipping
		Skipping() int
		//Verbosity is the requested verbosity level from the engine
		Verbosity() int
		//Log the looger to used during the Ekara execution
		Log() *log.Logger
		//Ef the exchange folder
		Ef() ExchangeFolder
		//Proxy returns launch context proxy settings
		Proxy() model.Proxy
		//SSHPublicKey the public key used by the engine during the process execution to connect the created nodes
		SSHPublicKey() string
		//SSHPrivateKey the private key used by the engine during the process execution to connect the created nodes
		SSHPrivateKey() string
		//ParamsFile returns the content the parameters provided by the user to fill the environment descriptor as a template
		ExternalVars() model.Parameters
	}
)
