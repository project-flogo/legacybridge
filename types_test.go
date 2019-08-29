package legacybridge

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetComplexObjectInfo(t *testing.T) {

	cpx := `{"value":"123", "metadata":"medatat"}`
	v, metadata, ok := GetComplexObjectInfo(cpx)
	assert.True(t, ok)
	assert.Equal(t, "medatat", metadata)
	assert.Equal(t, "123", v)

	cpx = `{"value":"123"}`
	v, metadata, ok = GetComplexObjectInfo(cpx)
	assert.False(t, ok)

	cpx = `{"metadata":"123"}`
	v, metadata, ok = GetComplexObjectInfo(cpx)
	assert.False(t, ok)

	cpx = `{"value":"", "metadata":""}`
	v, metadata, ok = GetComplexObjectInfo(cpx)
	assert.True(t, ok)
	assert.Equal(t, "", metadata)
	assert.Equal(t, "", v)

}
