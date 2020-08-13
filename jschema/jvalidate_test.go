package jschema

import (
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

func BenchmarkValidate(b *testing.B) {

	schemaLoader := gojsonschema.NewReferenceLoader("file://./schema.json")
	schema, err := gojsonschema.NewSchema(schemaLoader)

	if err != nil {
		b.Error("Error reading schema ")
	}
	for i := 0; i < b.N; i++ {
		ValidateDoc("file://./document.json", schema)
	}
}
