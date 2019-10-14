package action

import (
	"encoding/json"
	"fmt"
	"github.com/ekara-platform/engine/ansible"
	"github.com/ekara-platform/engine/component"
	"github.com/ekara-platform/engine/util"
	"gopkg.in/yaml.v2"
	"strings"
)

const (
	setupPlaybook   = "setup.yaml"
	createPlaybook  = "create.yaml"
	installPlaybook = "install.yaml"
	deployPlaybook  = "deploy.yaml"
)

type (
	//ApplyResult contains the results of environment application
	ApplyResult struct {
		Success   bool
		Inventory ansible.Inventory
	}
)

func (r ApplyResult) IsSuccess() bool {
	return r.Success
}

func (r ApplyResult) AsJson() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r ApplyResult) AsYaml() (string, error) {
	b, err := yaml.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r ApplyResult) AsPlainText() ([]string, error) {
	res, err := r.AsYaml()
	if err != nil {
		return []string{}, err
	}
	return strings.Split(res, "\n"), nil
}

var (
	applyAction = Action{
		ApplyActionID,
		CheckActionID,
		"Apply",
		[]step{providerSetup, ansibleInventory, providerCreate, orchestratorSetup, orchestratorInstall, stackDeploy},
	}
)

func providerSetup(rC *runtimeContext) (StepResults, Result) {
	sCs := InitStepResults()
	for _, p := range rC.cM.Environment().Providers {
		sc := InitPlaybookStepResult("Running the setup phase", p, NoCleanUpRequired)

		// Notify setup progress
		rC.pN.NotifyWithGoal("apply.provider.setup", len(rC.cM.Environment().Providers), "Preparing provider '%s'", p.Name)

		// Provider setup exchange folder
		setupProviderEf, ko := createChildExchangeFolder(rC.lC.Ef().Input, "setup_provider_"+p.Name, &sc)
		if ko {
			sCs.Add(sc)
			return *sCs, nil
		}

		setupProviderEfIn := setupProviderEf.Input
		setupProviderEfOut := setupProviderEf.Output

		// Create a new buffer
		buffer := ansible.CreateBuffer()

		// Prepare parameters
		bp := buildBaseParam(rC, "")
		bp.AddNamedMap("params", p.Parameters)
		if ko := saveBaseParams(bp, setupProviderEfIn, &sc); ko {
			sCs.Add(sc)
			return *sCs, nil
		}

		// Prepare extra vars
		exv := ansible.BuildExtraVars("", setupProviderEfIn, setupProviderEfOut, buffer)

		// Prepare environment variables
		env := ansible.BuildEnvVars()
		env.AddDefaultOsVars()
		env.AddProxy(rC.lC.Proxy())

		// Adding the environment variables from the provider
		for envK, envV := range p.EnvVars {
			env.Add(envK, envV)
		}

		// We launch the playbook
		usable, err := rC.cM.Use(p)
		if err != nil {
			FailsOnCode(&sc, err, "An error occurred getting the usable provider", nil)
		}
		defer usable.Release()
		code, err := rC.aM.Play(usable, setupPlaybook, exv, env, rC.pN)
		if err != nil {
			pfd := playBookFailureDetail{
				Playbook:  setupPlaybook,
				Component: p.ComponentName(),
				Code:      code,
			}
			FailsOnPlaybook(&sc, err, "An error occurred executing the playbook", pfd)
			sCs.Add(sc)
			return *sCs, nil
		}
		sCs.Add(sc)
	}

	// Notify setup finish
	rC.pN.Notify("apply.provider.setup", "All providers prepared")

	return *sCs, nil
}

func providerCreate(rC *runtimeContext) (StepResults, Result) {
	sCs := InitStepResults()
	for _, n := range rC.cM.Environment().NodeSets {
		sc := InitPlaybookStepResult("Running the create phase", n, NoCleanUpRequired)

		// Resolve provider
		p, err := n.Provider.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occurred resolving the provider"), nil)
			sCs.Add(sc)
			return *sCs, nil
		}

		// Notify creation progress
		rC.pN.NotifyWithGoal("apply.provider.create", len(rC.cM.Environment().NodeSets), "Creating node set '%s' with provider '%s'", n.Name, p.Name)

		// Create a new buffer
		buffer := ansible.CreateBuffer()

		// Prepare parameters
		bp := buildBaseParam(rC, n.Name)
		bp.AddInt("instances", n.Instances)
		bp.AddInterface("labels", n.Labels)
		bp.AddNamedMap("params", p.Parameters)
		bp.AddInterface("proxy", p.Proxy)

		// Prepare environment variables
		env := ansible.BuildEnvVars()
		env.AddDefaultOsVars()
		env.AddProxy(rC.lC.Proxy())

		// Adding the environment variables from the nodeset provider
		for envK, envV := range p.EnvVars {
			env.Add(envK, envV)
		}

		// Process hook : environment - provision - before
		runHookBefore(
			rC,
			sCs,
			rC.cM.Environment().Hooks.Provision,
			hookContext{"create", n, "environment", "provision", bp, env, buffer},
			NoCleanUpRequired,
		)

		// Process hook : nodeset - provision - before
		runHookBefore(
			rC,
			sCs,
			n.Hooks.Provision,
			hookContext{"create", n, "nodeset", "provision", bp, env, buffer},
			NoCleanUpRequired,
		)

		// Node creation exchange folder
		nodeCreateEf, ko := createChildExchangeFolder(rC.lC.Ef().Input, "create_"+n.Name, &sc)
		if ko {
			sCs.Add(sc)
			return *sCs, nil
		}

		if ko := saveBaseParams(bp, nodeCreateEf.Input, &sc); ko {
			sCs.Add(sc)
			return *sCs, nil
		}

		// Prepare extra vars
		exv := ansible.BuildExtraVars("", nodeCreateEf.Input, nodeCreateEf.Output, buffer)

		// Make the component usable
		usable, err := rC.cM.Use(p)
		if err != nil {
			FailsOnCode(&sc, err, "An error occurred getting the usable provider", nil)
		}
		defer usable.Release()

		// Launch the playbook
		code, err := rC.aM.Play(usable, createPlaybook, exv, env, rC.pN)
		if err != nil {
			pfd := playBookFailureDetail{
				Playbook:  createPlaybook,
				Component: p.ComponentName(),
				Code:      code,
			}
			FailsOnPlaybook(&sc, err, "An error occurred executing the playbook", pfd)
			sCs.Add(sc)
			return *sCs, nil
		}
		sCs.Add(sc)

		// Process hook : nodeset - provision - after
		runHookAfter(
			rC,
			sCs,
			n.Hooks.Provision,
			hookContext{"create", n, "nodeset", "provision", bp, env, buffer},
			NoCleanUpRequired,
		)

		// Process hook : environment - provision - after
		runHookAfter(
			rC,
			sCs,
			rC.cM.Environment().Hooks.Provision,
			hookContext{"create", n, "environment", "provision", bp, env, buffer},
			NoCleanUpRequired,
		)
	}

	// Notify creation finish
	rC.pN.Notify("apply.provider.create", "All node sets created")

	return *sCs, nil
}

func orchestratorSetup(rC *runtimeContext) (StepResults, Result) {
	o := rC.cM.Environment().Orchestrator
	sCs := InitStepResults()
	sc := InitPlaybookStepResult("Running the orchestrator setup phase", o, NoCleanUpRequired)

	// Notify setup progress
	rC.pN.NotifyWithGoal("apply.orchestrator.setup", 1, "Preparing orchestrator")

	// Create a new buffer
	buffer := ansible.CreateBuffer()

	// Orchestrator setup exchange folder
	setupOrchestratorEf, ko := createChildExchangeFolder(rC.lC.Ef().Input, "setup_orchestrator", &sc)
	if ko {
		sCs.Add(sc)
		return *sCs, nil
	}

	// Prepare parameters
	bp := buildBaseParam(rC, "")
	bp.AddNamedMap("params", o.Parameters)
	if ko := saveBaseParams(bp, setupOrchestratorEf.Input, &sc); ko {
		sCs.Add(sc)
		return *sCs, nil
	}

	// Prepare environment variables
	env := ansible.BuildEnvVars()
	env.AddDefaultOsVars()
	env.AddProxy(rC.lC.Proxy())

	// Adding the environment variables from the nodeset orchestrator
	for envK, envV := range o.EnvVars {
		env.Add(envK, envV)
	}

	// Prepare extra vars
	exv := ansible.BuildExtraVars("", setupOrchestratorEf.Input, setupOrchestratorEf.Output, buffer)

	// Make the component usable
	usable, err := rC.cM.Use(o)
	if err != nil {
		FailsOnCode(&sc, err, "An error occurred getting the usable orchestrator", nil)
	}
	defer usable.Release()

	// We launch the playbook
	code, err := rC.aM.Play(usable, setupPlaybook, exv, env, rC.pN)
	if err != nil {
		pfd := playBookFailureDetail{
			Playbook:  setupPlaybook,
			Component: o.ComponentName(),
			Code:      code,
		}
		FailsOnPlaybook(&sc, err, "An error occurred executing the playbook", pfd)
		sCs.Add(sc)
		return *sCs, nil
	}

	// Notify setup progress
	rC.pN.Notify("apply.orchestrator.setup", "Orchestrator prepared")

	sCs.Add(sc)
	return *sCs, nil
}

func orchestratorInstall(rC *runtimeContext) (StepResults, Result) {
	sCs := InitStepResults()

	for _, n := range rC.cM.Environment().NodeSets {
		sc := InitPlaybookStepResult("Running the orchestrator installation phase", n, NoCleanUpRequired)

		// Resolve the provider
		p, err := n.Provider.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occurred resolving the provider reference"), nil)
			sCs.Add(sc)
			return *sCs, nil
		}

		// Resolve the orchestrator
		o, err := n.Orchestrator.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occurred resolving the orchestrator reference"), nil)
			sCs.Add(sc)
			return *sCs, nil
		}

		// Notify setup progress
		rC.pN.NotifyWithGoal("apply.orchestrator.install", len(rC.cM.Environment().NodeSets), "Installing orchestrator on node set '%s'", n.Name)

		// Create a new buffer
		buffer := ansible.CreateBuffer()

		// Orchestrator install exchange folder
		installOrchestratorEf, ko := createChildExchangeFolder(rC.lC.Ef().Input, "install_orchestrator_"+n.Name, &sc)
		if ko {
			sCs.Add(sc)
			return *sCs, nil
		}

		// Prepare parameters
		bp := buildBaseParam(rC, n.Name)
		bp.AddInterface("labels", n.Labels)
		bp.AddNamedMap("params", o.Parameters)
		bp.AddInterface("proxy", p.Proxy)
		if ko := saveBaseParams(bp, installOrchestratorEf.Input, &sc); ko {
			sCs.Add(sc)
			return *sCs, nil
		}

		// Prepare environment variables
		env := ansible.BuildEnvVars()
		env.AddDefaultOsVars()
		env.AddProxy(rC.lC.Proxy())

		// Adding the environment variables from the nodeset orchestrator
		for envK, envV := range o.EnvVars {
			env.Add(envK, envV)
		}

		// Prepare extra vars
		exv := ansible.BuildExtraVars("", installOrchestratorEf.Input, installOrchestratorEf.Output, buffer)

		// Make the component usable
		usable, err := rC.cM.Use(o)
		if err != nil {
			FailsOnCode(&sc, err, "An error occurred getting the usable orchestrator", nil)
		}
		defer usable.Release()

		// Launch the playbook
		code, err := rC.aM.Play(usable, installPlaybook, exv, env, rC.pN)
		if err != nil {
			pfd := playBookFailureDetail{
				Playbook:  installPlaybook,
				Component: o.ComponentName(),
				Code:      code,
			}
			FailsOnPlaybook(&sc, err, "An error occurred executing the playbook", pfd)
			sCs.Add(sc)
			return *sCs, nil
		}
		sCs.Add(sc)
	}

	// Notify setup finish
	rC.pN.NotifyWithGoal("apply.orchestrator.install", len(rC.cM.Environment().NodeSets), "Orchestrator installed on all node sets")

	return *sCs, nil
}

func stackDeploy(rC *runtimeContext) (StepResults, Result) {
	sCs := InitStepResults()
	for _, st := range rC.cM.Environment().Stacks {
		sc := InitPlaybookStepResult("Deploying stack", st, NoCleanUpRequired)
		sCs.Add(sc)

		// Notify stack deploy
		rC.pN.NotifyWithGoal("apply.stack.deploy", len(rC.cM.Environment().Stacks), "Deploying stack '%s'", st.Name)

		// Stack deploy exchange folder for the given provider
		fName := fmt.Sprintf("deploy_stack_%s", st.Name)

		stackEf, ko := createChildExchangeFolder(rC.lC.Ef().Input, fName, &sc)
		if ko {
			sCs.Add(sc)
			return *sCs, nil
		}

		// Create a new buffer
		buffer := ansible.CreateBuffer()

		// Prepare parameters
		bp := buildBaseParam(rC, "")
		bp.AddNamedMap("params", st.Parameters)
		if ko := saveBaseParams(bp, stackEf.Input, &sc); ko {
			sCs.Add(sc)
			return *sCs, nil
		}

		// Prepare environment variables
		env := ansible.BuildEnvVars()
		env.AddDefaultOsVars()
		env.AddProxy(rC.lC.Proxy())

		// Adding the environment variables from the stack
		for envK, envV := range st.EnvVars {
			env.Add(envK, envV)
		}

		// Process hook : environment - deploy - before
		runHookBefore(
			rC,
			sCs,
			rC.cM.Environment().Hooks.Deploy,
			hookContext{"deploy", st, "environment", "deploy", bp, env, buffer},
			NoCleanUpRequired,
		)

		// Process hook : stack - deploy - before
		runHookBefore(
			rC,
			sCs,
			st.Hooks.Deploy,
			hookContext{"deploy", st, "stack", "deploy", bp, env, buffer},
			NoCleanUpRequired,
		)

		// Make the stack usable
		ust, err := rC.cM.Use(st)
		if err != nil {
			FailsOnCode(&sc, err, "An error occurred getting the usable stack", nil)
		}
		defer ust.Release()

		// If the stack is not self deployable, use the orchestrator deploy playbook
		var target component.UsableComponent
		var deployExtraVars string
		if ok, _ := ust.ContainsFile(deployPlaybook); !ok {
			o, err := rC.cM.Use(rC.cM.Environment().Orchestrator)
			if err != nil {
				FailsOnCode(&sc, err, "An error occurred getting the usable orchestrator", nil)
			}
			defer o.Release()
			target = o
			deployExtraVars = fmt.Sprintf("stack_path=%s stack_name=%s", ust.RootPath(), st.Name)
		} else {
			target = ust
		}

		// Prepare the extra vars
		exv := ansible.BuildExtraVars(
			deployExtraVars,
			stackEf.Input,
			stackEf.Output,
			buffer)

		// Execute the playbook
		code, err := rC.aM.Play(target, deployPlaybook, exv, env, rC.pN)
		if err != nil {
			pfd := playBookFailureDetail{
				Playbook:  deployPlaybook,
				Component: target.Name(),
				Code:      code,
			}
			FailsOnPlaybook(&sc, err, "An error occurred executing the playbook", pfd)
			sCs.Add(sc)
			return *sCs, nil
		}

		// Process hook : stack - deploy - after
		runHookAfter(
			rC,
			sCs,
			st.Hooks.Deploy,
			hookContext{"deploy", st, "stack", "deploy", bp, env, buffer},
			NoCleanUpRequired,
		)

		// Process hook : environment - deploy - after
		runHookAfter(
			rC,
			sCs,
			rC.cM.Environment().Hooks.Deploy,
			hookContext{"deploy", st, "environment", "deploy", bp, env, buffer},
			NoCleanUpRequired,
		)

		sCs.Add(sc)
	}

	// Notify stack deploy finish
	rC.pN.Notify("apply.stack.deploy", "All stacks deployed")

	return *sCs, nil
}

func ansibleInventory(rC *runtimeContext) (StepResults, Result) {
	sCs := InitStepResults()
	sr := InitPlaybookStepResult("Building inventory", nil, NoCleanUpRequired)

	inv, err := rC.aM.Inventory()
	if err != nil {
		FailsOnCode(&sr, err, "An error occurred during inventory", nil)
	}

	sCs.Add(sr)
	return *sCs, ApplyResult{
		Success:   true,
		Inventory: inv,
	}
}

func buildBaseParam(rC *runtimeContext, nodeSetName string) ansible.BaseParam {
	return ansible.BuildBaseParam(rC.cM.Environment(), rC.lC.SSHPublicKey(), rC.lC.SSHPrivateKey(), nodeSetName)
}

func createChildExchangeFolder(parent util.FolderPath, name string, sr *StepResult) (util.ExchangeFolder, bool) {
	ef, e := parent.AddChildExchangeFolder(name)
	if e != nil {
		err := fmt.Errorf(ErrorAddingExchangeFolder, name, e.Error())
		FailsOnCode(sr, e, err.Error(), nil)
		return ef, true
	}
	e = ef.Create()
	if e != nil {
		err := fmt.Errorf(ErrorCreatingExchangeFolder, name, e.Error())
		FailsOnCode(sr, e, err.Error(), nil)
		return ef, true
	}
	return ef, false
}

func saveBaseParams(bp ansible.BaseParam, dest util.FolderPath, sr *StepResult) bool {
	b, e := bp.Content()
	if e != nil {
		FailsOnCode(sr, e, fmt.Sprintf("An error occurred creating the base parameters"), nil)
		return true
	}
	_, e = util.SaveFile(dest, util.ParamYamlFileName, b)
	if e != nil {
		FailsOnCode(sr, e, fmt.Sprintf("An error occurred saving the parameter file into :%v", dest.Path()), nil)
		return true
	}
	return false
}
