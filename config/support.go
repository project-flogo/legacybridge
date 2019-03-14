package config

import (
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/legacybridge"

	legacyData "github.com/TIBCOSoftware/flogo-lib/core/data"
)

type ConversionContext struct {
	resources []*resource.Config
}

func (cc *ConversionContext) AddResource(res *resource.Config) {
	cc.resources = append(cc.resources, res)
}

func (cc *ConversionContext) AddSchema() {

}

func (cc *ConversionContext) AddImport() {

}

func ConvertLegacyAttr(legacyAttr *legacyData.Attribute) (*data.Attribute, error) {

	newType, _ := legacybridge.ToNewTypeFromLegacy(legacyAttr.Type())
	newVal := legacyAttr.Value()
	var newSchema schema.Schema

	//special handling for ComplexObjects
	if legacyAttr.Type() == legacyData.TypeComplexObject && legacyAttr.Value() != nil {

		if cVal, ok := legacyAttr.Value().(*legacyData.ComplexObject); ok {

			newVal = cVal.Value

			if cVal.Metadata != "" {
				//has schema
				def := &schema.Def{Type: "json", Value: cVal.Metadata}
				newSchema = &legacySchema{def}
			}
		}
	}

	return data.NewAttributeWithSchema(legacyAttr.Name(), newType, newVal, newSchema), nil
}

//To make sure json.marshal can serialize
type legacySchema struct {
	*schema.Def
}

func (s *legacySchema) Type() string {
	return s.Def.Type
}

func (s *legacySchema) Value() string {
	return s.Def.Value
}

func (*legacySchema) Validate(data interface{}) error {
	return nil
}
