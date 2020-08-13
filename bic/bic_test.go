package bic

import (
	"io/ioutil"
	"testing"
)

func BenchmarkValidate(b *testing.B) {

	config, err := GetProducerConfigFromFile("/Users/marlopezceli/Documents/dev/meli/metrics/go_jsonschema_test/bic/config-productor.json")
	if err != nil {
		b.Error("Error reading config " + err.Error())
	}

	for i := 0; i < b.N; i++ {
		payloadContent, err := ioutil.ReadFile("/Users/marlopezceli/Documents/dev/meli/metrics/go_jsonschema_test/document.json")
		if err != nil {
			b.Error("Error reading document file")
		}

		Validate(payloadContent, config)
	}

}
