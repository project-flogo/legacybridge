package flow

import (
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/flow/definition"
	"github.com/project-flogo/legacybridge/config"

	legacyDef "github.com/TIBCOSoftware/flogo-contrib/action/flow/definition"
	legacyData "github.com/TIBCOSoftware/flogo-lib/core/data"
)

func convertLegacyFlow(rep *legacyDef.DefinitionRep) (*definition.DefinitionRep, error) {

	if rep.RootTask != nil {
		return nil, fmt.Errorf("definition too old to be automatically converted")
	}

	newDef := &definition.DefinitionRep{}
	newDef.Name = rep.Name
	newDef.ModelID = rep.ModelID
	newDef.ExplicitReply = rep.ExplicitReply

	if rep.Metadata != nil {
		newDef.Metadata = &metadata.IOMetadata{}
		if len(rep.Metadata.Input) > 0 {
			newDef.Metadata.Input = make(map[string]data.TypedValue)
			for name, attr := range rep.Metadata.Input {
				newAttr, err := config.ConvertLegacyAttr(attr)
				if err != nil {
					return nil, err
				}
				newDef.Metadata.Input[name] = newAttr
			}
		}
		if len(rep.Metadata.Output) > 0 {
			newDef.Metadata.Output = make(map[string]data.TypedValue)
			for name, attr := range rep.Metadata.Output {
				newAttr, err := config.ConvertLegacyAttr(attr)
				if err != nil {
					return nil, err
				}
				newDef.Metadata.Output[name] = newAttr
			}
		}
	}

	if len(rep.Tasks) != 0 {

		for _, taskRep := range rep.Tasks {

			task, err := createTask(taskRep)

			if err != nil {
				return nil, err
			}
			newDef.Tasks = append(newDef.Tasks, task)
		}
	}

	if len(rep.Links) != 0 {

		for _, linkRep := range rep.Links {

			link := createLink(linkRep)
			newDef.Links = append(newDef.Links, link)
		}
	}

	if rep.ErrorHandler != nil {

		errorHandler := &definition.ErrorHandlerRep{}
		newDef.ErrorHandler = errorHandler

		if len(rep.ErrorHandler.Tasks) != 0 {

			for _, taskRep := range rep.ErrorHandler.Tasks {

				task, err := createTask(taskRep)
				if err != nil {
					return nil, err
				}
				errorHandler.Tasks = append(errorHandler.Tasks, task)
			}
		}

		if len(rep.ErrorHandler.Links) != 0 {

			for _, linkRep := range rep.ErrorHandler.Links {
				link := createLink(linkRep)
				errorHandler.Links = append(errorHandler.Links, link)
			}
		}
	}

	return newDef, nil
}

func createTask(rep *legacyDef.TaskRep) (*definition.TaskRep, error) {
	task := &definition.TaskRep{}
	task.ID = rep.ID
	task.Name = rep.Name
	task.Settings = rep.Settings
	task.Type = rep.Type

	if rep.ActivityCfgRep != nil {

		actCfg, err := createActivityConfig(rep.ActivityCfgRep)
		if err != nil {
			return nil, err
		}

		task.ActivityCfgRep = actCfg
	}

	return task, nil
}

func createActivityConfig(rep *legacyDef.ActivityConfigRep) (*definition.ActivityConfigRep, error) {

	activityCfg := &definition.ActivityConfigRep{}
	activityCfg.Settings = rep.Settings
	activityCfg.Ref = rep.Ref

	input := make(map[string]interface{})
	inputSchemas := make(map[string]*schema.Def)

	if len(rep.InputAttrs) > 0 {
		for key, value := range rep.InputAttrs {
			if co, ok := value.(*legacyData.ComplexObject); ok {
				input[key] = co.Value
				if co.Metadata != "" {
					inputSchemas[key] = &schema.Def{Type: "json", Value: co.Metadata}
				}
			} else {
				input[key] = value
			}
		}
	}

	output := make(map[string]interface{})
	outputSchemas := make(map[string]*schema.Def)

	if len(rep.OutputAttrs) > 0 {
		for key, value := range rep.OutputAttrs {
			if co, ok := value.(*legacyData.ComplexObject); ok {
				output[key] = co.Value
				if co.Metadata != "" {
					outputSchemas[key] = &schema.Def{Type: "json", Value: co.Metadata}
				}
			} else {
				output[key] = value
			}
		}
	}

	if rep.Mappings != nil {
		lm := &legacyData.IOMappings{}
		lm.Input = rep.Mappings.Input
		lm.Output = rep.Mappings.Output

		inputMappings, outputMappings, err := config.ConvertLegacyMappings(lm, definition.GetDataResolver())
		if err != nil {
			return nil, err
		}

		if len(inputMappings) > 0 {
			for key, value := range inputMappings {
				input[key] = value
			}
		}

		if len(outputMappings) > 0 {
			for key, value := range outputMappings {
				output[key] = value
			}
		}
	}

	if len(input) > 0 {
		activityCfg.Input = input
	}

	if len(output) > 0 {
		activityCfg.Output = output
	}

	if len(inputSchemas) > 0 || len(outputSchemas) > 0 {
		activityCfg.Schemas = &definition.ActivitySchemasRep{}

		if len(inputSchemas) > 0 {
			activityCfg.Schemas.Input = inputSchemas
		}

		if len(outputSchemas) > 0 {
			activityCfg.Schemas.Output = outputSchemas
		}
	}

	return activityCfg, nil
}

func createLink(linkRep *legacyDef.LinkRep) *definition.LinkRep {

	link := &definition.LinkRep{}
	link.Name = linkRep.Name
	link.Value = linkRep.Value
	link.Type = linkRep.Type
	link.ToID = linkRep.ToID
	link.FromID = linkRep.FromID

	return link
}
