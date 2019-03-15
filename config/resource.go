package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/project-flogo/core/app/resource"

	legacyResource "github.com/TIBCOSoftware/flogo-lib/app/resource"
)

type ResourceDataConverter func(legacyData json.RawMessage) (json.RawMessage, error)

var resourceDataConverters = make(map[string]ResourceDataConverter)

func RegisterResourceDataConverter(resourceType string, converter ResourceDataConverter) {
	resourceDataConverters[resourceType] = converter
}

func ConvertLegacyResource(lrConfig *legacyResource.Config) (*resource.Config, error) {

	newConfig := &resource.Config{ID: lrConfig.ID}

	idInfo := strings.Split(lrConfig.ID, ":")
	if len(idInfo) < 2 {
		return nil, fmt.Errorf("invalid resource id: %s", lrConfig.ID)
	}

	resType := idInfo[0]

	if converter, ok := resourceDataConverters[resType]; ok {
		newData, err := converter(lrConfig.Data)
		if err != nil {
			return nil, fmt.Errorf("error converting '%s': %s", lrConfig.ID, err.Error())
		}
		newConfig.Data = newData
	} else {
		// todo log a warn that no resource convert for resType registered?
		newConfig.Data = lrConfig.Data
	}

	return newConfig, nil
}
