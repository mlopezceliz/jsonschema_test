package bic

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"regexp"
	"strings"

	"github.com/mercadolibre/go-meli-toolkit/goutils/apierrors"
	"github.com/mercadolibre/go-meli-toolkit/goutils/logger"
)

var VALIDATED_METRIC = false

func GetProducerConfigFromFile(configFilePath string) (*StructProducerConfig, error) {
	configContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	producerConfig, producerConfigError := GetProducerConfig("1", configContent)
	if producerConfigError != nil {
		return nil, producerConfigError
	}

	return producerConfig, nil
}

func GetProducerConfig(token string, configBytes []byte) (*StructProducerConfig, apierrors.ApiError) {

	config := new(StructProducerConfig)

	//Unmarshalling body into StructPayload
	if unmarshallError := json.Unmarshal(configBytes, config); unmarshallError != nil {
		logger.Error("Error unmarshalling body from feed into struct config. Body: "+string(json.RawMessage(configBytes)), unmarshallError)
		err := apierrors.NewInternalServerApiError("error unmarshalling body from feed into struct payload", unmarshallError)
		return nil, err
	}

	return config, nil
}

func Validate(payloadContent []byte, producerConfig *StructProducerConfig) (bool, error) {

	token := "1"

	//Getting a StructPayload from request body
	payload, err := GetPayloadBody(token, payloadContent)
	if err != nil {
		return false, err
	}

	_, err = validatePayloadBody(token, payload)
	if err != nil {
		return false, err
	}

	_, validationError := validatePayload(payload, producerConfig)
	if validationError != nil {
		return false, validationError
	}

	return true, nil
}

func GetPayloadBody(token string, jsonBytes []byte) (*StructPayload, apierrors.ApiError) {

	payload := new(StructPayload)

	//Unmarshalling body into StructPayload
	if unmarshallError := json.Unmarshal(jsonBytes, payload); unmarshallError != nil {
		logger.Error("Error unmarshalling body from feed into struct payload. Body: "+string(json.RawMessage(jsonBytes)), unmarshallError)
		err := apierrors.NewInternalServerApiError("error unmarshalling body from feed into struct payload", unmarshallError)
		return nil, err
	}

	return payload, nil
}

func validatePayloadBody(token string, payload *StructPayload) (*StructPayload, apierrors.ApiError) {

	//Validate that payload keys are not nil
	if payloadKeysError := payloadKeysValidation(payload.ID, payload.Entity); payloadKeysError != nil {
		logger.Error("Error validating payload keys", payloadKeysError)
		err := apierrors.NewInternalServerApiError("payload keys validation", payloadKeysError)
		return nil, err
	}

	//Validate that metrics block is not nil. It could be {} if producer do not want to save new metrics.
	if payload.Metrics == nil {
		payloadMetricsBlockError := errors.New("Metrics block can't be null. If you don't want to post any metrics, send an empty map {} in metrics instead of null")
		logger.Error("Error validating payload metrics block", payloadMetricsBlockError)
		err := apierrors.NewInternalServerApiError("payload metrics block validation", payloadMetricsBlockError)
		return nil, err
	}

	payload.ProducerToken = token
	return payload, nil
}

func validatePayload(payload *StructPayload, producerConfig *StructProducerConfig) (*StructPayload, apierrors.ApiError) {
	token := payload.ProducerToken
	if !strings.EqualFold(producerConfig.Entity, payload.Entity) {
		err := apierrors.NewUnauthorizedApiError("provided entity does not match the one in the producer configuration")
		logger.Errorf("Unauthorized provided entity [id: %v][entity: %v][configurationEntity: %v][token: %v]", err, payload.ID, payload.Entity, producerConfig.Entity, token)
		return nil, err
	}

	if !strings.EqualFold(producerConfig.Status, "enabled") {
		err := apierrors.NewUnauthorizedApiError("producer not enabled")
		logger.Errorf("Unauthorized producer [id: %v][entity: %v][token: %v]", err, payload.ID, payload.Entity, token)
		return nil, err
	}

	if producerConfig.MandatoryFields != nil {
		if mandatoryFieldsError := checkMandatoryFields(producerConfig.MandatoryFields, payload.Metrics); mandatoryFieldsError != nil {
			err := NewNotAcceptableApiError("missing a few mandatory fields", mandatoryFieldsError)
			logger.Errorf("Missing a few mandatory fields [id: %v][entity: %v][configurationEntity: %v][token: %v]", mandatoryFieldsError, payload.ID, payload.Entity, token)
			return nil, err
		}
	}

	newPayload, producerConfigError := checkProducerConfig(producerConfig, payload)
	if producerConfigError != nil {
		err := NewNotAcceptableApiError("provided metrics do not match the ones in the producer configuration", producerConfigError)
		logger.Errorf("Not acceptable provided metrics [id: %v][entity: %v][token: %v]", producerConfigError, payload.ID, payload.Entity, token)
		return nil, err
	}
	return newPayload, nil
}

func payloadKeysValidation(id string, entity string) error {
	if id == "" && entity == "" {
		return fmt.Errorf("payload does not contain valid id and entity [id: %v][entity: %v]", id, entity)
	} else if id == "" {
		return fmt.Errorf("payload does not contain valid id [id: %v]", id)
	} else if entity == "" {
		return fmt.Errorf("payload does not contain valid entity [entity: %v]", entity)
	} else {
		return nil
	}
}

func checkProducerConfig(config *StructProducerConfig, payload *StructPayload) (*StructPayload, error) {
	newPayload := new(StructPayload)
	newPayload.ID = payload.ID
	newPayload.Entity = payload.Entity
	newPayload.ProducerToken = payload.ProducerToken

	payloadMetrics := payload.Metrics
	configMetrics := config.AllowedMetrics

	for key, metricsBlock := range payloadMetrics {
		var pathMetric []string              //Always create the path root of current metrics block
		pathMetric = append(pathMetric, key) //Appends the first level of the block before calling metrics block validation method

		pathError, err := validateMetricBlock(metricsBlock, configMetrics, &pathMetric, "", nil) //key and value are necessary for recursion inside the validateMetricBlock method, so they are nil in this case

		if err != nil { //In case validateMetricBlock method returns error, pathError is used for return exact error point at path
			concatPathError := ""
			for _, path := range *pathError {
				concatPathError = concatPathError + path + "."
			}
			concatPathError = concatPathError[:len(concatPathError)-1]
			error := fmt.Errorf("%v at %v", err.Error(), concatPathError)
			return nil, error
		}
	}

	newPayload.Metrics = payload.Metrics

	return newPayload, nil
}

func validateMetricBlock(metricsBlock interface{}, configMetrics interface{}, pathMetric *[]string, key string, value interface{}) (*[]string, error) {
	subLevelBlock, subLevelIsMap := metricsBlock.(map[string]interface{}) //Validates if an interface{} is a map
	var validatingError error
	var pathError *[]string
	if subLevelIsMap { //If an interface is a map, then iterates it looking for metric leaf before recursion on validateMetricBlock
		for subLevelKey, subLevelValue := range subLevelBlock {
			*pathMetric = append(*pathMetric, subLevelKey) //Saves the next level of the block in path
			pathError, validatingError = validateMetricBlock(subLevelBlock[subLevelKey], configMetrics, pathMetric, subLevelKey, subLevelValue)
			if validatingError != nil { //If a error occurs in a recursive call, keeps original cause in all recursive calls
				return pathError, validatingError
			}
			*pathMetric = (*pathMetric)[:len(*pathMetric)-1] //Once metric leaf has been validated, returns to the previous point of the path
		}
	}

	_, ValueIsMap := value.(map[string]interface{}) //Validates that value is a metric leaf before calling validatePathAndTypeOfLeafMetric

	if !ValueIsMap && key != "" { //In parallel, validates that key is not missing to avoid a not metric leaf
		VALIDATED_METRIC = false                                                                                       //Uses a global variable to save the state of metric leaf validation
		validatedMetric, pathError, error := validatePathAndTypeOfLeafMetric(configMetrics, pathMetric, 0, key, value) //pathPosition is necessary for recursion inside the validatePathAndTypeOfLeafMetric method, so its value must be 0 in this case
		if validatedMetric {
			return nil, nil
		}
		if error != nil {
			return pathError, error
		}
	}
	if validatingError == nil {
		return nil, nil
	} else {
		return pathError, validatingError
	}
}

func validatePathAndTypeOfLeafMetric(configMetricsBlock interface{}, path *[]string, pathPosition int, keyMetric string, valueMetric interface{}) (bool, *[]string, error) {
	subLevelConfigBlock, ok := configMetricsBlock.(map[string]interface{})
	var err error

	if ok && valueMetric != nil {
		for subLevelConfigKey, subLevelConfigValue := range subLevelConfigBlock {
			if pathPosition < len(*path) {
				if (*path)[pathPosition] == subLevelConfigKey {
					pathPosition++
					validatedMetric, pathError, recursiveError := validatePathAndTypeOfLeafMetric(subLevelConfigValue, path, pathPosition, keyMetric, valueMetric)
					err = recursiveError
					if validatedMetric {
						return VALIDATED_METRIC, nil, nil
					} else {
						return VALIDATED_METRIC, pathError, err
					}
				}
			} else {
				pathError := *path
				err = fmt.Errorf("invalid metric level")
				logger.Error("invalid metric level", err)
				return VALIDATED_METRIC, &pathError, err
			}
		}
	} else {
		if valueMetric == nil {
			VALIDATED_METRIC = true
		} else {
			err = checkLeavesTypes(valueMetric, configMetricsBlock, keyMetric)
			if err != nil {
				pathError := *path
				return VALIDATED_METRIC, &pathError, err
			} else {
				VALIDATED_METRIC = true
			}
		}
	}
	if VALIDATED_METRIC {
		return VALIDATED_METRIC, nil, nil
	} else {
		var pathError []string
		if pathPosition >= len(*path) {
			pathError = *path
		} else {
			pathError = (*path)[:pathPosition+1]
		}
		err = fmt.Errorf("invalid metric name")
		logger.Error("invalid metric name", err)
		return VALIDATED_METRIC, &pathError, err
	}
}

func checkLeavesTypes(metricValue interface{}, typeConfigMetric interface{}, keyMetric string) error {
	stringType := fmt.Sprintf("%v", typeConfigMetric)
	if metricTypeChecker(metricValue, stringType) {
		return nil
	} else {
		logger.Debugf("field '%v' with different data type, sent value: '%v'", keyMetric, metricValue)
		return fmt.Errorf("field '%v' with different data type, sent value: %v", keyMetric, metricValue)
	}
}

func checkMandatoryFields(mandatoryFields *[]string, metrics map[string]interface{}) error {
	flattenedPaths := make(map[string]interface{})
	flattenMetricsMap(metrics, flattenedPaths, "", "")
	validatorMap := make(map[string]bool)
	for _, mandatoryFieldPath := range *mandatoryFields {
		validated := false
		for path, pathValue := range flattenedPaths {
			_, validatedPath := validatorMap[path]
			if !validatedPath {
				if strings.EqualFold(mandatoryFieldPath, path) {
					if pathValue != nil {
						validated = true
						validatorMap[path] = validated
						break
					}
				}
			}

		}
		if !validated {
			return fmt.Errorf("missing mandatory field: %v", mandatoryFieldPath)
		}
	}
	return nil
}

func flattenMetricsMap(metrics map[string]interface{}, flattenedPaths map[string]interface{}, root string, path string) {
	for key, value := range metrics {
		valueMap, isMap := value.(map[string]interface{})
		if !isMap {
			if strings.EqualFold("", path) {
				flattenedPaths[key] = value
			} else {
				pathToSave := path + "." + key
				flattenedPaths[pathToSave] = value
			}
		} else {
			if strings.EqualFold("", path) {
				path = key
			} else {
				path = path + "." + key
			}
			subRoot := key
			flattenMetricsMap(valueMap, flattenedPaths, subRoot, path)
			path = root
		}
	}
}

func metricTypeChecker(metricValue interface{}, t string) bool {
	switch metricValue.(type) {
	case int:
		if t == "number" {
			return true
		} else if t == "boolean_number" {
			if metricValue == 1 || metricValue == 0 {
				return true
			}
		}
	case float64:
		if t == "number" {
			return true
		} else if t == "boolean_number" {
			if metricValue == 1.0 || metricValue == 0.0 {
				return true
			}
		}
	case string:
		if t == "string" {
			return true
		} else if t == "date" {
			return dateFormatChecker(fmt.Sprintf("%v", metricValue))
		} else if t == "time" {
			return timeFormatChecker(fmt.Sprintf("%v", metricValue))
		} else if t == "datetime" {
			return dateTimeFormatChecker(fmt.Sprintf("%v", metricValue))
		}
	case bool:
		if t == "bool" || t == "boolean" {
			return true
		}
	case []interface{}:
		if t == "array" {
			return true
		}
	}

	return false //No se pudo identificar el tipo o tipo erroneo
}

func dateFormatChecker(s string) bool {
	re := regexp.MustCompile("((19|20)..)-(0[1-9]|1[012])-(0[1-9]|1[0-9]|2[0-9]|3[01])")
	return re.MatchString(s)
}

func timeFormatChecker(s string) bool {
	re := regexp.MustCompile("(0[0-9]|1[0-9]|2[0-3]):(0[0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9]):(0[0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])")
	return re.MatchString(s)
}

func dateTimeFormatChecker(s string) bool {
	re := regexp.MustCompile("((19|20)..)-(0[1-9]|1[012])-(0[1-9]|1[0-9]|2[0-9]|3[01])T(0[0-9]|1[0-9]|2[0-3]):(0[0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9]):(0[0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])(.+)")
	return re.MatchString(s)
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
