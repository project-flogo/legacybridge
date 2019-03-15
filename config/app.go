package config

import (
	"encoding/json"

	"github.com/project-flogo/core/app"

	legacyApp "github.com/TIBCOSoftware/flogo-lib/app"
)

func ConvertLegacyJson(legacyJson string) (string, error) {

	laConfig, err := legacyApp.LoadConfig(legacyJson)
	if err != nil {
		return "", err
	}

	newConfig, err := ConvertLegacyAppConfig(laConfig)
	if err != nil {
		return "", err
	}

	newBytes, err := json.MarshalIndent(newConfig, "", "    ")
	if err != nil {
		return "", err
	}

	return string(newBytes), nil
}

func ConvertLegacyAppConfig(laConfig *legacyApp.Config) (*app.Config, error) {

	newConfig := &app.Config{}

	newConfig.Name = laConfig.Name
	newConfig.Type = laConfig.Type
	newConfig.Version = laConfig.Version
	newConfig.Description = laConfig.Description
	newConfig.AppModel = "1.1.0"

	//collect imports

	ctx := &ConversionContext{}

	//properties
	if len(laConfig.Properties) > 0 {
		for _, oldProp := range laConfig.Properties {
			newAttr, err := ConvertLegacyAttr(oldProp)
			if err != nil {
				return nil, err
			}
			newConfig.Properties = append(newConfig.Properties, newAttr)
		}
	}

	//channels
	if len(laConfig.Channels) > 0 {
		for _, channel := range laConfig.Channels {
			newConfig.Channels = append(newConfig.Channels, channel)
		}
	}

	//actions
	if len(laConfig.Actions) > 0 {
		for _, laConfig := range laConfig.Actions {
			newActionConfig, err := ConvertLegacyAction(ctx, laConfig)
			if err != nil {
				return nil, err
			}

			newConfig.Actions = append(newConfig.Actions, newActionConfig)
		}
	}

	if len(laConfig.Triggers) > 0 {
		for _, ltConfig := range laConfig.Triggers {
			newTriggerConfig, err := convertLegacyTrigger(ctx, ltConfig)
			if err != nil {
				return nil, err
			}

			newConfig.Triggers = append(newConfig.Triggers, newTriggerConfig)
		}
	}

	//resources
	if len(laConfig.Resources) > 0 {
		for _, lrConfig := range laConfig.Resources {
			newResourceConfig, err := ConvertLegacyResource(lrConfig)
			if err != nil {
				return nil, err
			}

			newConfig.Resources = append(newConfig.Resources, newResourceConfig)
		}
	}

	if ctx.resources != nil {
		for _, res := range ctx.resources {
			newConfig.Resources = append(newConfig.Resources, res)
		}
	}

	return newConfig, nil
}
