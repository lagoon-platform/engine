package engine

import (
	"testing"

	"github.com/ekara-platform/engine/util"

	"github.com/ekara-platform/engine/model"

	"github.com/stretchr/testify/assert"
)

func TestFetchOrderedGlobalHookCreateBefore(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  create:
    before:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  create:
    before:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Create.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Create.Before), 2)
		assert.Equal(t, env.Hooks.Create.Before[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Create.Before[1].Prefix, "hook2Prefix")
	}

}

func TestFetchOrderedGlobalHookCreateAfter(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  create:
    after:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  create:
    after:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Create.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Create.After), 2)
		assert.Equal(t, env.Hooks.Create.After[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Create.After[1].Prefix, "hook2Prefix")
	}

}

func TestFetchOrderedGlobalHookInstallBefore(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  install:
    before:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  install:
    before:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Install.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Install.Before), 2)
		assert.Equal(t, env.Hooks.Install.Before[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Install.Before[1].Prefix, "hook2Prefix")
	}

}

func TestFetchOrderedGlobalHookInstallAfter(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  install:
    after:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  install:
    after:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Install.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Install.After), 2)
		assert.Equal(t, env.Hooks.Install.After[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Install.After[1].Prefix, "hook2Prefix")
	}

}

func TestFetchOrderedGlobalHookDeployBefore(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  deploy:
    before:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  deploy:
    before:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Deploy.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Deploy.Before), 2)
		assert.Equal(t, env.Hooks.Deploy.Before[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Deploy.Before[1].Prefix, "hook2Prefix")
	}

}

func TestFetchOrderedGlobalHookDeployAfter(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  deploy:
    after:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  deploy:
    after:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Deploy.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Deploy.After), 2)
		assert.Equal(t, env.Hooks.Deploy.After[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Deploy.After[1].Prefix, "hook2Prefix")
	}

}

func TestFetchOrderedGlobalHookDeleteBefore(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  delete:
    before:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  delete:
    before:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Destroy.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Destroy.Before), 2)
		assert.Equal(t, env.Hooks.Destroy.Before[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Destroy.Before[1].Prefix, "hook2Prefix")
	}

}

func TestFetchOrderedGlobalHookDestroyAfter(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  delete:
    after:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  delete:
    after:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Destroy.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Destroy.After), 2)
		assert.Equal(t, env.Hooks.Destroy.After[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Destroy.After[1].Prefix, "hook2Prefix")
	}

}

func TestFetchOrderedGlobalHookInitBefore(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  init:
    before:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  init:
    before:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Init.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Init.Before), 2)
		assert.Equal(t, env.Hooks.Init.Before[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Init.Before[1].Prefix, "hook2Prefix")
	}

}

func TestFetchOrderedGlobalHookInitAfter(t *testing.T) {
	tester := util.CreateComponentTester(t, model.CreateEmptyParameters())
	defer tester.Clean()

	tester.CreateDirEmptyDesc("comp1")
	tester.CreateDirEmptyDesc("comp2")
	tester.CreateDirEmptyDesc("comp3")
	tester.CreateDirEmptyDesc("comp4")

	repParent1 := tester.CreateDir("parent1")
	repDesc := tester.CreateDir("descriptor")

	parent1Content := `
ekara:
  components:
    comp4:
      repository: comp4
tasks:
  hook1:
    component: comp4 
    playbook: dummy.yaml
hooks:
  init:
    after:
      - task: hook1  
        prefix: hook1Prefix
`
	repParent1.WriteCommit("ekara.yaml", parent1Content)

	descContent := `
name: ekaraDemoVar
qualifier: dev

ekara:
  parent:
    repository: parent1
  components:
    comp1:
      repository: comp1
    comp2:
      repository: comp2
    comp3:
      repository: comp3

          
orchestrator:
  component: comp1
providers:
  p1:
    component: comp2
tasks:
  hook2:
    component: comp3
    playbook: dummy.yaml

nodes:
  node1:
    instances: 1
    provider:
      name: p1
hooks:
  init:
    after:
      - task: hook2
        prefix: hook2Prefix  
`
	repDesc.WriteCommit("ekara.yaml", descContent)

	tester.Init(repDesc.AsRepository("master"))
	env := tester.Env()
	assert.NotNil(t, env)

	tester.AssertComponentsExactly(model.MainComponentId, model.MainComponentId+model.ParentComponentSuffix, "comp1", "comp2", "comp3", "comp4")

	assert.Exactly(t, []string{"comp4", model.MainComponentId + model.ParentComponentSuffix, "comp1", "comp2", "comp3", model.MainComponentId}, tester.ComponentManager().ComponentOrder())

	if assert.True(t, env.Hooks.Init.HasTasks()) {
		assert.Equal(t, len(env.Hooks.Init.After), 2)
		assert.Equal(t, env.Hooks.Init.After[0].Prefix, "hook1Prefix")
		assert.Equal(t, env.Hooks.Init.After[1].Prefix, "hook2Prefix")
	}

}
