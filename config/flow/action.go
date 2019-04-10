package flow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	legacyFlow "github.com/TIBCOSoftware/flogo-contrib/action/flow"
	legacyDef "github.com/TIBCOSoftware/flogo-contrib/action/flow/definition"
	legacyAction "github.com/TIBCOSoftware/flogo-lib/core/action"
	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/legacybridge/config"
)

func init() {
	config.RegisterActionConverter("github.com/TIBCOSoftware/flogo-contrib/action/flow", ConvertLegacyFlowAction)
	config.RegisterResourceDataConverter("flow", ConvertLegacyResourceData)
}

func ConvertLegacyFlowAction(ctx *config.ConversionContext, laConfig *legacyAction.Config) (*action.Config, error) {

	newConfig := &action.Config{Id: laConfig.Id}
	newConfig.Ref = "github.com/project-flogo/flow"

	newConfig.Settings = make(map[string]interface{})
	for key, value := range laConfig.Settings {
		newConfig.Settings[key] = value
	}

	var actionData *legacyFlow.ActionData
	err := json.Unmarshal(laConfig.Data, &actionData)
	if err != nil {
		return nil, fmt.Errorf("faild to load flow action data error: %s", err.Error())
	}

	if actionData.FlowURI != "" {
		newConfig.Settings["flowURI"] = actionData.FlowURI

	} else {
		resConfig, err := createResource(actionData)
		if err != nil {
			return nil, fmt.Errorf("faild to create flow resource from flow data: %s", err.Error())
		}

		newConfig.Settings["flowURI"] = "res://" + resConfig.ID
		ctx.AddResource(resConfig)
	}

	return newConfig, nil
}

func createResource(actionData *legacyFlow.ActionData) (*resource.Config, error) {

	resourceCfg := &resource.Config{ID: "flow:" + strconv.Itoa(time.Now().Nanosecond())}

	if actionData.FlowCompressed != nil {
		//todo un-compress flow
		resourceCfg.Data = actionData.FlowCompressed
	} else if actionData.Flow != nil {
		resourceCfg.Data = actionData.Flow
	} else {
		return nil, fmt.Errorf("flow not provided for Flow Action")
	}

	return resourceCfg, nil
}

func ConvertLegacyResourceData(oldData json.RawMessage) (json.RawMessage, error) {

	var oldDef *legacyDef.DefinitionRep
	err := json.Unmarshal(oldData, &oldDef)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling old flow: %s", err.Error())
	}

	newDef, err := convertLegacyFlow(oldDef)
	if err != nil {
		return nil, fmt.Errorf("error converting flow: %s", err.Error())
	}

	bytes, err := json.Marshal(newDef)
	if err != nil {
		return nil, fmt.Errorf("error marshalling new flow: %s", err.Error())
	}

	return json.RawMessage(bytes), nil
}
