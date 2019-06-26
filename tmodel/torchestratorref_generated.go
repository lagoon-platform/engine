package engine

import(
	"github.com/ekara-platform/model"
)

//*****************************************************************************
//
// This file is autogenerated by "go generate .". Do not modify.
//
//*****************************************************************************

// ----------------------------------------------------
// TOrchestratorRef is a read only reference on the orchestrator
// ----------------------------------------------------
type TOrchestratorRef interface {
    //Orchestrator returns the orchestrator managing a node
    Orchestrator() (TOrchestrator, error)
	
}

// ----------------------------------------------------
// Implementation(s) of TOrchestratorRef  
// ----------------------------------------------------

//TOrchestratorRefOnOrchestratorRefHolder is the struct containing the OrchestratorRef in otder to implement TOrchestratorRef  
type TOrchestratorRefOnOrchestratorRefHolder struct {
	h 	model.OrchestratorRef
}

//CreateTOrchestratorRefForOrchestratorRef returns a holder implementing TOrchestratorRef
func CreateTOrchestratorRefForOrchestratorRef(o model.OrchestratorRef) TOrchestratorRefOnOrchestratorRefHolder {
	return TOrchestratorRefOnOrchestratorRefHolder{
		h: o,
	}
}

//Orchestrator returns the orchestrator managing a node
func (r TOrchestratorRefOnOrchestratorRefHolder) Orchestrator() (TOrchestrator, error){
	    v, err := r.h.Resolve()
    return CreateTOrchestratorForOrchestrator(v), err
}

