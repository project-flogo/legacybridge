package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	legacyData "github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	_ "github.com/project-flogo/core/data/expression/script"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/resolve"
	"github.com/stretchr/testify/assert"
)

func TestArrayMapper(t *testing.T) {
	oldArray := `{
    "fields": [
        {
            "from": "tstring.concat(\"this street name: \", \"ddd\")",
            "to": "$.street",
            "type": "primitive"
        },
        {
            "from": "tstring.concat(\"The zipcode is: \",$.zipcode)",
            "to": "$.zipcode",
            "type": "primitive"
        },
        {
            "from": "$.state",
            "to": "$.state",
            "type": "primitive"
        },
		{
    		"from": "$.array",
    		"to": "$.array",
            "type": "foreach",
			"fields":[
				{
           			 "from": "$.field1",
           			 "to": "$.tofield1",
           			 "type": "assign"
        		},
				{
            		"from": "$.field2",
					"to": "$.tofield2",
            		"type": "assign"
        		},
				{
            		"from": "flogo",
					"to": "$.tofield3",
            		"type": "assign"
        		}
			]

		}
    ],
    "from": "$activity[a1].field.addresses",
    "to": ".field.addresses",
    "type": "foreach"
}
`

	array, err := ParseArrayMapping(oldArray)
	assert.Nil(t, err)

	v, err := ToNewArray(array, resolve.GetBasicResolver())
	assert.Nil(t, err)

	vv, _ := json.Marshal(v)
	fmt.Println(string(vv))

	assert.Equal(t, "=$.state", v.(map[string]interface{})["@foreach($activity[a1].field.addresses)"].(map[string]interface{})["state"])
	assert.Equal(t, "=tstring.concat(\"this street name: \", \"ddd\")", v.(map[string]interface{})["@foreach($activity[a1].field.addresses)"].(map[string]interface{})["street"])
	assert.Equal(t, "=$.field1", v.(map[string]interface{})["@foreach($activity[a1].field.addresses)"].(map[string]interface{})["array"].(map[string]interface{})["@foreach($.array)"].(map[string]interface{})["tofield1"])
	assert.Equal(t, "=$.field2", v.(map[string]interface{})["@foreach($activity[a1].field.addresses)"].(map[string]interface{})["array"].(map[string]interface{})["@foreach($.array)"].(map[string]interface{})["tofield2"])
	assert.Equal(t, "flogo", v.(map[string]interface{})["@foreach($activity[a1].field.addresses)"].(map[string]interface{})["array"].(map[string]interface{})["@foreach($.array)"].(map[string]interface{})["tofield3"])

}

func TestNewArrayMapper(t *testing.T) {
	oldArray := `{
    "fields": [
        {
            "from": "tstring.concat(\"this street name: \", \"ddd\")",
            "to": "$.street",
            "type": "primitive"
        },
        {
            "from": "tstring.concat(\"The zipcode is: \",$.zipcode)",
            "to": "$.zipcode",
            "type": "primitive"
        },
        {
            "from": "$.state",
            "to": "$.state",
            "type": "primitive"
        },
		{
    		"from": "NEWARRAY",
    		"to": "$.array",
            "type": "foreach",
			"fields":[
				{
           			 "from": "$.field1",
           			 "to": "$.tofield1",
           			 "type": "assign"
        		},
				{
            		"from": "$.field2",
					"to": "$.tofield2",
            		"type": "assign"
        		},
				{
            		"from": "wangzai",
					"to": "$.tofield3",
            		"type": "assign"
        		}
			]

		}
    ],
    "from": "NEWARRAY",
    "to": ".field.addresses",
    "type": "foreach"
}
`

	array, err := ParseArrayMapping(oldArray)
	assert.Nil(t, err)

	v, err := ToNewArray(array, resolve.GetBasicResolver())
	assert.Nil(t, err)

	vv, _ := json.Marshal(v)
	fmt.Println(string(vv))

	assert.Equal(t, "=$.state", v.([]interface{})[0].(map[string]interface{})["state"])
	assert.Equal(t, "=tstring.concat(\"this street name: \", \"ddd\")", v.([]interface{})[0].(map[string]interface{})["street"])
	assert.Equal(t, "=$.field1", v.([]interface{})[0].(map[string]interface{})["array"].([]interface{})[0].(map[string]interface{})["tofield1"])
}

func TestPathToObject(t *testing.T) {
	path := []string{"data", "field", "value"}
	obj, err := constructObjectFromPath(path, "1234", make(map[string]interface{}))
	assert.Nil(t, err)
	v, _ := json.Marshal(obj)
	fmt.Println(string(v))
	assert.Equal(t, "1234", obj.(map[string]interface{})["data"].(map[string]interface{})["field"].(map[string]interface{})["value"])
}

func TestPathToObjectArray(t *testing.T) {
	path := []string{"data[2]", "field[0]", "value"}

	obj, err := constructObjectFromPath(path, "1234", make(map[string]interface{}))
	assert.Nil(t, err)
	v, _ := json.Marshal(obj)
	fmt.Println(string(v))
	assert.Equal(t, "1234", obj.(map[string]interface{})["data"].([]interface{})[2].(map[string]interface{})["field"].([]interface{})[0].(map[string]interface{})["value"])
}

func TestMultiplePathToObject(t *testing.T) {
	path := []string{"data", "field", "value"}

	path2 := []string{"data", "field2", "value"}

	obj, err := constructObjectFromPath(path, "1234", make(map[string]interface{}))
	assert.Nil(t, err)

	obj, err = constructObjectFromPath(path2, "1234", obj)

	v, _ := json.Marshal(obj)
	fmt.Println(string(v))
	assert.Equal(t, "1234", obj.(map[string]interface{})["data"].(map[string]interface{})["field"].(map[string]interface{})["value"])
}

func TestMultiplePathToObjectArray(t *testing.T) {
	path := []string{"data[2]", "field[0]", "value"}
	path2 := []string{"data[4]", "field[0]", "value"}

	obj, err := constructObjectFromPath(path, "1234", make(map[string]interface{}))
	assert.Nil(t, err)

	obj, err = constructObjectFromPath(path2, "1234", obj)
	assert.Nil(t, err)

	v, _ := json.Marshal(obj)
	fmt.Println(string(v))
	assert.Equal(t, "1234", obj.(map[string]interface{})["data"].([]interface{})[2].(map[string]interface{})["field"].([]interface{})[0].(map[string]interface{})["value"])
	assert.Equal(t, "1234", obj.(map[string]interface{})["data"].([]interface{})[4].(map[string]interface{})["field"].([]interface{})[0].(map[string]interface{})["value"])

}

func TestMultiplePathToObjectArray2(t *testing.T) {
	path := []string{"data", "field", "value[0]"}
	path2 := []string{"data", "field", "value[4]"}

	obj, err := constructObjectFromPath(path, 22, make(map[string]interface{}))
	assert.Nil(t, err)

	obj, err = constructObjectFromPath(path2, 33, obj)
	assert.Nil(t, err)

	v, _ := json.Marshal(obj)
	fmt.Println(string(v))
	assert.Equal(t, 22, obj.(map[string]interface{})["data"].(map[string]interface{})["field"].(map[string]interface{})["value"].([]interface{})[0])
	assert.Equal(t, 33, obj.(map[string]interface{})["data"].(map[string]interface{})["field"].(map[string]interface{})["value"].([]interface{})[4])

}

func TestConvertMappingValue(t *testing.T) {
	mappings := `{
        "input": [
         {
          "mapTo": "input1",
          "type": "expression",
          "value": "$.body.id"
         },
         {
          "mapTo": "input2",
          "type": "expression",
          "value": "$.body.name"
         }
        ],
        "output": [
         {
          "mapTo": "code",
          "type": "expression",
          "value": 200
         },
         {
          "mapTo": "data.return",
          "type": "expression",
          "value": "$.res"
         }
        ]
       }`

	mapping := &legacyData.IOMappings{}

	err := json.Unmarshal([]byte(mappings), mapping)
	assert.Nil(t, err)

	input, output, err := ConvertLegacyMappings(mapping, resolve.GetBasicResolver())
	assert.Nil(t, err)

	v, _ := json.Marshal(input)
	fmt.Println("input:", string(v))

	v2, _ := json.Marshal(output)
	fmt.Println("output:", string(v2))
	assert.Equal(t, "=$.body.id", input["input1"])
	assert.Equal(t, "=$.body.name", input["input2"])

}

func TestConvertMappingValue2(t *testing.T) {
	mappings := `{
        "input": [
         {
          "mapTo": "input1.a.b",
          "type": "expression",
          "value": "$.body.id"
         },
         {
          "mapTo": "input1.a.c",
          "type": "expression",
          "value": "$.body.name"
         }
        ],
        "output": [
         {
          "mapTo": "code",
          "type": "expression",
          "value": 200
         },
         {
          "mapTo": "data.return",
          "type": "expression",
          "value": "$.res"
         }
        ]
       }`

	mapping := &legacyData.IOMappings{}

	err := json.Unmarshal([]byte(mappings), mapping)
	assert.Nil(t, err)

	input, output, err := ConvertLegacyMappings(mapping, resolve.GetBasicResolver())
	assert.Nil(t, err)

	v, _ := json.Marshal(input)
	fmt.Println("input:", string(v))

	v2, _ := json.Marshal(output)
	fmt.Println("output:", string(v2))

	assert.Equal(t, "=$.body.id", input["input1"].(*mapper.ObjectMapping).Mapping.(map[string]interface{})["a"].(map[string]interface{})["b"])
	assert.Equal(t, "=$.body.name", input["input1"].(*mapper.ObjectMapping).Mapping.(map[string]interface{})["a"].(map[string]interface{})["c"])

}

func TestConvertMappingValue3(t *testing.T) {
	mappings := `{
        "input": [
         {
          "mapTo": "input1[0].a.b",
          "type": "expression",
          "value": "$.body.id"
         },
         {
          "mapTo": "data[0]",
          "type": "expression",
          "value": "$.body.name"
         },
		{
          "mapTo": "data[1]",
          "type": "expression",
          "value": "$.body.name"
         }
        ],
        "output": [
         {
          "mapTo": "code",
          "type": "expression",
          "value": 200
         },
         {
          "mapTo": "data.return",
          "type": "expression",
          "value": "$.res"
         }
        ]
       }`

	mapping := &legacyData.IOMappings{}

	err := json.Unmarshal([]byte(mappings), mapping)
	assert.Nil(t, err)

	input, output, err := ConvertLegacyMappings(mapping, resolve.GetBasicResolver())
	assert.Nil(t, err)

	v, _ := json.Marshal(input)
	fmt.Println("input:", string(v))

	v2, _ := json.Marshal(output)
	fmt.Println("output:", string(v2))
	assert.Equal(t, "=$.body.name", input["data"].(*mapper.ObjectMapping).Mapping.([]interface{})[0])
	assert.Equal(t, "=$.body.name", input["data"].(*mapper.ObjectMapping).Mapping.([]interface{})[1])
	assert.Equal(t, "=$.body.id", input["input1"].(*mapper.ObjectMapping).Mapping.([]interface{})[0].(map[string]interface{})["a"].(map[string]interface{})["b"])

	//output
	assert.Equal(t, "=$.res", output["data"].(*mapper.ObjectMapping).Mapping.(map[string]interface{})["return"])
	assert.Equal(t, float64(200), output["code"])

}

func TestConvertMappingValue4(t *testing.T) {
	mappings := `{
        "input": [
         {
          "mapTo": "data[0]",
          "type": "expression",
          "value": "$.body.name"
         },
		{
          "mapTo": "data[1]",
          "type": "expression",
          "value": "$.body.ddd"
         }
        ]
       }`

	mapping := &legacyData.IOMappings{}

	err := json.Unmarshal([]byte(mappings), mapping)
	assert.Nil(t, err)

	input, output, err := ConvertLegacyMappings(mapping, resolve.GetBasicResolver())
	assert.Nil(t, err)

	v, _ := json.Marshal(input)
	fmt.Println("input:", string(v))

	v2, _ := json.Marshal(output)
	fmt.Println("output:", string(v2))
	assert.Equal(t, "=$.body.name", input["data"].(*mapper.ObjectMapping).Mapping.([]interface{})[0])
	assert.Equal(t, "=$.body.ddd", input["data"].(*mapper.ObjectMapping).Mapping.([]interface{})[1])
}

func TestConvertMappingWithArrayMapping(t *testing.T) {
	mappings := `{
        "input": [
         {
          "mapTo": "input1[0].a.b",
          "type": "expression",
          "value": "$.body.id"
         },
		 {
				"mapto":"input1[1].c",
				"type":"array",
				"value":"{\r\n    \"fields\": [\r\n        {\r\n            \"from\": \"tstring.concat(\\\"this street name: \\\", \\\"ddd\\\")\",\r\n            \"to\": \"$.street\",\r\n            \"type\": \"primitive\"\r\n        },\r\n        {\r\n            \"from\": \"tstring.concat(\\\"The zipcode is: \\\",$.zipcode)\",\r\n            \"to\": \"$.zipcode\",\r\n            \"type\": \"primitive\"\r\n        },\r\n        {\r\n            \"from\": \"$.state\",\r\n            \"to\": \"$.state\",\r\n            \"type\": \"primitive\"\r\n        },\r\n\t\t{\r\n    \t\t\"from\": \"$.array\",\r\n    \t\t\"to\": \"$.array\",\r\n            \"type\": \"foreach\",\r\n\t\t\t\"fields\":[\r\n\t\t\t\t{\r\n           \t\t\t \"from\": \"$.field1\",\r\n           \t\t\t \"to\": \"$.tofield1\",\r\n           \t\t\t \"type\": \"assign\"\r\n        \t\t},\r\n\t\t\t\t{\r\n            \t\t\"from\": \"$.field2\",\r\n\t\t\t\t\t\"to\": \"$.tofield2\",\r\n            \t\t\"type\": \"assign\"\r\n        \t\t},\r\n\t\t\t\t{\r\n            \t\t\"from\": \"wangzai\",\r\n\t\t\t\t\t\"to\": \"$.tofield3\",\r\n            \t\t\"type\": \"assign\"\r\n        \t\t}\r\n\t\t\t]\r\n\r\n\t\t}\r\n    ],\r\n    \"from\": \"$activity[a1].field.addresses\",\r\n    \"to\": \".field.addresses\",\r\n    \"type\": \"foreach\"\r\n}"
         },
         {
          "mapTo": "data[0]",
          "type": "expression",
          "value": "$.body.name"
         },
		{
          "mapTo": "data[1]",
          "type": "expression",
          "value": "$.body.name"
         }
        ],
        "output": [
         {
          "mapTo": "code",
          "type": "expression",
          "value": 200
         },
         {
          "mapTo": "data.return",
          "type": "expression",
          "value": "$.res"
         }
        ]
       }`

	mapping := &legacyData.IOMappings{}

	err := json.Unmarshal([]byte(mappings), mapping)
	assert.Nil(t, err)

	input, output, err := ConvertLegacyMappings(mapping, resolve.GetBasicResolver())
	assert.Nil(t, err)

	v, _ := json.Marshal(input)
	fmt.Println("input:", string(v))

	v2, _ := json.Marshal(output)
	fmt.Println("output:", string(v2))
	assert.Equal(t, "=$.body.name", input["data"].(*mapper.ObjectMapping).Mapping.([]interface{})[0])
	assert.Equal(t, "=$.body.name", input["data"].(*mapper.ObjectMapping).Mapping.([]interface{})[1])
	assert.Equal(t, "=$.body.id", input["input1"].(*mapper.ObjectMapping).Mapping.([]interface{})[0].(map[string]interface{})["a"].(map[string]interface{})["b"])
	assert.Equal(t, "=$.state", input["input1"].(*mapper.ObjectMapping).Mapping.([]interface{})[1].(map[string]interface{})["c"].(map[string]interface{})["@foreach($activity[a1].field.addresses)"].(map[string]interface{})["state"])
	//output
	assert.Equal(t, "=$.res", output["data"].(*mapper.ObjectMapping).Mapping.(map[string]interface{})["return"])
	assert.Equal(t, float64(200), output["code"])

}

func TestConvertLiteralObject(t *testing.T) {
	mappings := `{
        "input": [
         {
          "mapTo": "field1.id",
          "type": "object",
          "value": {"id":"id2", "name":"name2"}
         }
		]
       }`

	mapping := &legacyData.IOMappings{}

	err := json.Unmarshal([]byte(mappings), mapping)
	assert.Nil(t, err)

	input, _, err := ConvertLegacyMappings(mapping, resolve.GetBasicResolver())
	assert.Nil(t, err)

	v, _ := json.Marshal(input)
	fmt.Println("input:", string(v))

	assert.Equal(t, "id2", input["field1"].(*mapper.ObjectMapping).Mapping.(map[string]interface{})["id"].(map[string]interface{})["id"])

}

func init() {
	_ = function.Register(&fnConcat{})
	function.SetPackageAlias(reflect.ValueOf(fnConcat{}).Type().PkgPath(), "tstring")
	function.ResolveAliases()
}

type fnConcat struct {
}

func (fnConcat) Name() string {
	return "concat"
}

func (fnConcat) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeString}, true
}

func (fnConcat) Eval(params ...interface{}) (interface{}, error) {
	if len(params) >= 2 {
		var buffer bytes.Buffer

		for _, v := range params {
			buffer.WriteString(v.(string))
		}
		return buffer.String(), nil
	}

	return "", fmt.Errorf("fnConcat function must have at least two arguments")
}
