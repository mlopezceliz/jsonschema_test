package main

import (
	"fmt"
	"io/ioutil"

	"github.com/mercadolibre/jsonschema_test/bic"
	"github.com/mercadolibre/jsonschema_test/jschema"
	"github.com/mercadolibre/jsonschema_test/stopwatch"

	"github.com/xeipuuv/gojsonschema"
)

var count = 1000

var fileRelativePath = "./document.json"

func main() {

	fmt.Printf("Test reading and validate de case file %s %v times \n", fileRelativePath, count)

	testJsonSchema()

	testBic()

}

func testJsonSchema() {
	schemaLoader := gojsonschema.NewReferenceLoader("file://./jschema/schema.json")
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		fmt.Println("Error reading schema file")
	}

	watch := stopwatch.Start()

	for i := 0; i < count; i++ {

		payloadContent, err := ioutil.ReadFile(fileRelativePath)
		if err != nil {
			fmt.Println("Error reading document file")
		}

		valid, err := jschema.ValidateBytes(payloadContent, schema)
		if valid {
			//fmt.Println("Valid document")
			//Unmarshall del file para contemplar este tiempo tambien.
			_, err = bic.GetPayloadBody("1", payloadContent)
			if err != nil {
				fmt.Printf("Invalid unmarsharll , error: %s \n", err)
			}
		} else {
			fmt.Printf("Invalid document, error: %s \n", err)
		}
	}
	watch.Stop()
	fmt.Printf("Json Schema Tiempo total: %v \n", watch.Milliseconds())
}

// BIC bi-consumers-ship
func testBic() {
	config, err := bic.GetProducerConfigFromFile("./bic/config-productor.json")
	if err != nil {
		fmt.Println("Error reading config " + err.Error())
	}

	watch := stopwatch.Start()
	for i := 0; i < count; i++ {
		payloadContent, err := ioutil.ReadFile(fileRelativePath)
		if err != nil {
			fmt.Println("Error reading document file")
		}

		valid, err := bic.Validate(payloadContent, config)
		if valid {
			//fmt.Println("Valid document")
		} else {
			fmt.Printf("Invalid document, error: %s \n", err)
		}
	}
	watch.Stop()
	fmt.Printf("BIC Tiempo total: %v \n", watch.Milliseconds())
}
