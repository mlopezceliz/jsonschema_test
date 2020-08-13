package jschema

import (
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func Validate() (bool, error) {
	return validateDocSch("file://./document.json", "file://./jschema/schema.json")
}

func ValidateBytes(doc []byte, schema *gojsonschema.Schema) (bool, error) {
	documentLoader := gojsonschema.NewBytesLoader(doc)
	return validateLoader(documentLoader, schema)
}

func ValidateDoc(doc string, schema *gojsonschema.Schema) (bool, error) {

	documentLoader := gojsonschema.NewReferenceLoader(doc)
	return validateLoader(documentLoader, schema)
}

func validateLoader(loader gojsonschema.JSONLoader, schema *gojsonschema.Schema) (bool, error) {

	result, err := schema.Validate(loader)
	if err != nil {
		return false, errors.New("Error reading document " + err.Error())
	}

	if result.Valid() {
		return true, nil
	} else {
		fmt.Printf("The document is not valid. see errors :\n")
		errorStr := ""
		for _, desc := range result.Errors() {
			errorStr = errorStr + fmt.Sprintf("- %s\n", desc)
		}
		return false, errors.New(errorStr)
	}
}

func validateDocSch(doc string, sch string) (bool, error) {

	schemaLoader := gojsonschema.NewReferenceLoader(sch)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return false, errors.New("Error reading schema ")
	}

	return ValidateDoc(doc, schema)
}
