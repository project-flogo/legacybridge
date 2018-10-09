package legacybridge

import (
	"errors"

	olddata "github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/project-flogo/core/data"
)

// ToTypeEnum get the data type that corresponds to the specified name
func ToNewTypeFromLegacy(legacyType olddata.Type) (data.Type, error) {

	switch legacyType {
	case olddata.TypeAny:
		return data.TypeAny, nil
	case olddata.TypeString:
		return data.TypeString, nil
	case olddata.TypeInteger:
		return data.TypeInt, nil
	case olddata.TypeLong:
		return data.TypeInt64, nil
	case olddata.TypeDouble:
		return data.TypeFloat64, nil
	case olddata.TypeBoolean:
		return data.TypeBool, nil
	case olddata.TypeObject:
		return data.TypeObject, nil
	case olddata.TypeParams:
		return data.TypeParams, nil
	case olddata.TypeArray:
		return data.TypeArray, nil
	case olddata.TypeComplexObject:
		return data.TypeComplexObject, nil
	default:
		return 0, errors.New("unknown type: " + legacyType.String())
	}
}

// ToTypeEnum get the data type that corresponds to the specified name
func ToLegacyFromNewType(dataType data.Type) (olddata.Type, error) {

	switch dataType {
	case data.TypeAny:
		return olddata.TypeAny, nil
	case data.TypeString:
		return olddata.TypeString, nil
	case data.TypeInt:
		return olddata.TypeInteger, nil
	case data.TypeInt64:
		return olddata.TypeLong, nil
	case data.TypeFloat64:
		return olddata.TypeDouble, nil
	case data.TypeBool:
		return olddata.TypeBoolean, nil
	case data.TypeObject:
		return olddata.TypeObject, nil
	case data.TypeParams:
		return olddata.TypeParams, nil
	case data.TypeArray:
		return olddata.TypeArray, nil
	case data.TypeComplexObject:
		return olddata.TypeComplexObject, nil
	default:
		return 0, errors.New("unknown type: " + dataType.String())
	}
}
