package config

import (
	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/core/trigger"
	"github.com/project-flogo/legacybridge"

	legacyAction "github.com/TIBCOSoftware/flogo-lib/core/action"
	legacyTrigger "github.com/TIBCOSoftware/flogo-lib/core/trigger"
)

type ActionConverter func(ctx *ConversionContext, laConfig *legacyAction.Config) (*action.Config, error)

var actionConverters = make(map[string]ActionConverter)

func RegisterActionConverter(actionRef string, converter ActionConverter) {
	actionConverters[actionRef] = converter
}

func ConvertLegacyAction(ctx *ConversionContext, laConfig *legacyAction.Config) (*action.Config, error) {

	newConfig := &action.Config{Id: laConfig.Id}

	newConfig.Settings = make(map[string]interface{})
	for key, value := range laConfig.Settings {
		newConfig.Settings[key] = value
	}

	if converter, ok := actionConverters[laConfig.Ref]; ok {
		return converter(ctx, laConfig)
	} else {
		newConfig := &action.Config{Id: laConfig.Id}
		newConfig.Ref = laConfig.Ref

		newConfig.Settings = make(map[string]interface{})
		for key, value := range laConfig.Settings {
			newConfig.Settings[key] = value
		}

		return newConfig, nil
	}
}

func convertLegacyTrigger(ctx *ConversionContext, ltConfig *legacyTrigger.Config) (*trigger.Config, error) {

	newConfig := &trigger.Config{}
	newConfig.Id = ltConfig.Id
	newConfig.Ref = ltConfig.Ref

	if len(ltConfig.Settings) > 0 {
		newConfig.Settings = make(map[string]interface{})
		for key, value := range ltConfig.Settings {
			newConfig.Settings[key] = value
		}
	}

	//todo should we ignore ltConfig.Output?

	if len(ltConfig.Handlers) > 0 {

		for _, legacyHandler := range ltConfig.Handlers {

			newHandler, err := convertLegacyHandler(ctx, legacyHandler)
			if err != nil {
				return nil, err
			}

			newConfig.Handlers = append(newConfig.Handlers, newHandler)
		}
	}

	return newConfig, nil
}

func convertLegacyHandler(ctx *ConversionContext, ltHandlerConfig *legacyTrigger.HandlerConfig) (*trigger.HandlerConfig, error) {

	newConfig := &trigger.HandlerConfig{}
	newConfig.Name = ltHandlerConfig.Name

	if len(ltHandlerConfig.Settings) > 0 {
		newConfig.Settings = make(map[string]interface{})
		for key, value := range ltHandlerConfig.Settings {
			newConfig.Settings[key] = value
		}
	}

	outSchemas := make(map[string]interface{})

	_, schemas := ConvertValues(ltHandlerConfig.Output)
	if len(schemas) > 0 {
		outSchemas = schemas
	}

	//handle old deprecated outputs
	_, oschemas := ConvertValues(ltHandlerConfig.Outputs)
	if len(oschemas) > 0 {
		for k, v := range oschemas {
			outSchemas[k] = v
		}
	}

	if len(outSchemas) > 0 {
		newConfig.OutputSchemas = outSchemas
	}

	if ltHandlerConfig.ActionId != "" {

		/*
			"actionId": "test_http_error",
			"actionMappings": {
				"input": [],
				"output": []
			}
		*/

		newAction := &action.Config{Id: ltHandlerConfig.ActionId}
		newActionCfg := &trigger.ActionConfig{Config: newAction}

		input, output, err := ConvertLegacyMappings(ltHandlerConfig.Action.Mappings, resolve.GetBasicResolver())
		if err != nil {
			return nil, err
		}

		newActionCfg.Input = input
		newActionCfg.Output = output

		newConfig.Actions = append(newConfig.Actions, newActionCfg)
	} else {

		newAction, err := ConvertLegacyAction(ctx, ltHandlerConfig.Action.Config)
		if err != nil {
			return nil, err
		}

		newActionCfg := &trigger.ActionConfig{Config: newAction}

		input, output, err := ConvertLegacyMappings(ltHandlerConfig.Action.Mappings, resolve.GetBasicResolver())
		if err != nil {
			return nil, err
		}

		newActionCfg.Input = input
		newActionCfg.Output = output

		newConfig.Actions = append(newConfig.Actions, newActionCfg)
	}

	return newConfig, nil
}

func ConvertValues(oldValues map[string]interface{}) (map[string]interface{}, map[string]interface{}) {

	newVals := make(map[string]interface{})
	newSchemas := make(map[string]interface{})

	if len(oldValues) > 0 {
		for name, value := range oldValues {
			newVals[name] = value

			if value != nil {
				// cannot rely on activity metadata, since we don't know what is imported,
				// so we guess based on value
				v, s, ok := legacybridge.GetComplexObjectInfo(value)

				if ok {
					if v != "" && v != "{}" {
						newVals[name] = v
					} else {
						//Empty value, remove it so we don't create a output mapper
						delete(newVals, name)
					}
					if s != "" {
						newSchemas[name] = &schema.Def{Type: "json", Value: s}
					}
				}
			}
		}
	}

	return newVals, newSchemas
}
