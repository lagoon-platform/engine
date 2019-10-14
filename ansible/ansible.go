package ansible

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"github.com/ekara-platform/engine/component"
	"github.com/ekara-platform/engine/util"
)

const (
	taskPrefix = "TASK ["
)

type (
	//Manager executes an ansible playbook
	Manager interface {
		// Play runs a playbook within a component
		//
		// Parameters:
		//		cr: the component holding the playbook to launch
		//		playbook: the name of the playbook to launch
		//		extraVars: the extra vars passed to the playbook
		//		envVars: the environment variables set before launching the playbook
		//		pN: progress notifier
		//
		Play(cr component.UsableComponent, playbook string, extraVars ExtraVars, envVars EnvVars, pN util.ProgressNotifier) (int, error)
		// Inventory returns the current inventory of environment nodes
		Inventory() (Inventory, error)
	}

	manager struct {
		lC               util.LaunchContext
		componentManager component.Manager
	}

	execChan struct {
		out    chan string
		err    chan string
		status chan int
	}
)

//CreateAnsibleManager returns a new AnsibleManager, providing managed execution of
//Ansible commands
func CreateAnsibleManager(lC util.LaunchContext, componentManager component.Manager) Manager {
	return &manager{
		lC:               lC,
		componentManager: componentManager,
	}
}

func (aM manager) Play(uc component.UsableComponent, playbook string, extraVars ExtraVars, envVars EnvVars, pN util.ProgressNotifier) (int, error) {
	ok, playBookPath := uc.ContainsFile(playbook)
	if !ok {
		return 0, fmt.Errorf("component \"%s\" does not contain playbook: %s", uc.Name(), playbook)
	}
	aM.lC.Log().Printf("Executing playbook %s from component %s", playBookPath.RelativePath(), playBookPath.Component().Name())

	var args = []string{playbook}

	// Discovered modules
	modulePaths := aM.findModulePaths()
	defer modulePaths.Release()
	args = append(args, aM.buildModuleArgs(modulePaths)...)

	// Discovered inventory sources
	inventoryPaths := aM.findInventoryPaths()
	defer inventoryPaths.Release()
	args = append(args, aM.buildInventoryArgs(inventoryPaths)...)

	// Extra vars
	args = append(args, aM.buildExtraVarsArgs(extraVars)...)

	eC, err := aM.exec(uc.RootPath(), "ansible-playbook", args, envVars)
	if err != nil {
		return 0, err
	}

	// Read the logs as they come until a status code is returned
	for {
		select {
		case <-eC.err:
			// drop err lines
		case outLine := <-eC.out:
			// Detect tasks to show progression
			sTrim := strings.TrimSpace(outLine)
			if strings.Index(sTrim, "TASK [") == 0 {
				pN.Detail(sTrim[len(taskPrefix):strings.LastIndex(sTrim, "]")])
			}
			aM.lC.Log().Println(outLine)
		case status := <-eC.status:
			aM.lC.Log().Printf("Playbook finished (%d)", status)
			return status, nil
		}
	}
}

func (aM manager) Inventory() (Inventory, error) {
	res := Inventory{}

	// Discovered inventory sources
	inventoryPaths := aM.findInventoryPaths()
	defer inventoryPaths.Release()

	args := []string{"--list"}
	args = append(args, aM.buildInventoryArgs(inventoryPaths)...)

	eC, err := aM.exec(aM.lC.Ef().RootFolder, "ansible-inventory", args, EnvVars{})
	if err != nil {
		return res, err
	}

	// Read the output until a status code is returned
	sb := strings.Builder{}
	var finished bool
	for !finished {
		select {
		case <-eC.err:
			// drop err lines
		case outLine := <-eC.out:
			sb.WriteString(outLine)
		case status := <-eC.status:
			aM.lC.Log().Printf("Inventory done (%d)", status)
			finished = true
		}
	}

	// Parse the output
	err = res.UnmarshalJSON([]byte(sb.String()))
	if err != nil {
		return res, err
	}

	return res, nil
}

func (aM manager) buildModuleArgs(modulePaths component.MatchingPaths) []string {
	var args []string
	if modulePaths.Count() > 0 {
		pathsStrings := modulePaths.JoinAbsolutePaths(":")
		aM.lC.Log().Printf("Playbook modules directorie(s): %s", pathsStrings)
		args = append(args, "--module-path", pathsStrings)
	} else {
		aM.lC.Log().Printf("No playbook module directory")
	}
	return args
}

func (aM manager) findModulePaths() component.MatchingPaths {
	return aM.componentManager.ContainsDirectory(util.ComponentModuleFolder)
}

func (aM manager) findInventoryPaths() component.MatchingPaths {
	return aM.componentManager.ContainsDirectory(util.InventoryModuleFolder)
}

func (aM manager) buildInventoryArgs(inventoryPaths component.MatchingPaths) []string {
	var args []string
	if inventoryPaths.Count() > 0 {
		asArgs := inventoryPaths.PrefixPaths("-i")
		aM.lC.Log().Printf("Playbook inventory directorie(s): %s", inventoryPaths.JoinAbsolutePaths(":"))
		for _, v := range asArgs {
			if v == "-i" {
				continue
			}
		}
		args = append(args, asArgs...)
	} else {
		aM.lC.Log().Printf("No playbook inventory directory")
	}
	return args
}

func (aM manager) buildExtraVarsArgs(extraVars ExtraVars) []string {
	var args []string
	if extraVars.Bool {
		aM.lC.Log().Printf("Playbook extra var(s): %s", extraVars.String())
		args = append(args, "--extra-vars", extraVars.String())
	} else {
		aM.lC.Log().Printf("No playbook extra var")
	}
	return args
}

func (aM manager) exec(dir string, ex string, args []string, envVars EnvVars) (execChan, error) {
	eC := execChan{
		out:    make(chan string),
		err:    make(chan string),
		status: make(chan int),
	}

	cmd := exec.Command(ex, args...)
	cmd.Dir = dir
	cmd.Env = []string{}
	for k, v := range envVars.Content {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	errReader, err := cmd.StderrPipe()
	if err != nil {
		return eC, err
	}
	logPipe(errReader, eC.err)

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return eC, err
	}
	logPipe(outReader, eC.out)

	err = cmd.Start()
	if err != nil {
		return eC, err
	}

	go func() {
		err = cmd.Wait()
		if err != nil {
			e, ok := err.(*exec.ExitError)
			if ok {
				s := e.Sys().(syscall.WaitStatus)
				eC.status <- s.ExitStatus()
			}
		} else {
			eC.status <- 0
		}
	}()

	return eC, nil
}

// logPipe logs the given pipe, reader/closer on the given logger
func logPipe(rc io.ReadCloser, ch chan string) {
	s := bufio.NewScanner(rc)
	go func() {
		for s.Scan() {
			ch <- s.Text()
		}
	}()
}
