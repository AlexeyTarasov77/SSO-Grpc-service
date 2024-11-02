package validator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

const EmptyErrors = "{}"

func camelToSnake(s string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(s, "${1}_${2}")

	return strings.ToLower(snake)
}

func getFieldName(obj any, origFieldName string) (fieldName string) {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	field, found := t.FieldByName(origFieldName)
	if !found {
		panic(fmt.Sprintf("Field %s not found in type %s", origFieldName, t.Name()))
	}
	if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
		jsonName := strings.Split(tag, ",")[0]
		if jsonName != "" {
			fieldName = jsonName
		}
	} else {
		fieldName = camelToSnake(origFieldName)
	}
	return
}


func GetErrorMsgForField(obj any, err validator.FieldError) (errorMsg string) {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	field, found := t.FieldByName(err.StructField())
	if !found {
		panic(fmt.Sprintf("Field %s not found in type %s", err.StructField(), t.Name()))
	}
	errorMsg = field.Tag.Get("errorMsg")
	if errorMsg == "" {
		switch err.Tag() {
		case "required":
			errorMsg = "This field is required"
		case "max":
			errorMsg = fmt.Sprintf("The maximum value is %s", err.Param())
		case "min":
			errorMsg = fmt.Sprintf("The minimum value is %s", err.Param())
		case "gte":
			errorMsg = fmt.Sprintf("Value should be greater than or equal to %s", err.Param())
		case "lte":
			errorMsg = fmt.Sprintf("Value should be less than or equal to %s", err.Param())
		case "lt":
			errorMsg = fmt.Sprintf("Value should be less than %s", err.Param())
		case "gt":
			errorMsg = fmt.Sprintf("Value should be greater than %s", err.Param())
		case "eqfield", "eq":
			errorMsg = fmt.Sprintf("Value should be equal to %s", err.Param())
		case "nefield", "ne":
			errorMsg = fmt.Sprintf("Value should not be equal to %s", err.Param())
		case "oneof":
			errorMsg = fmt.Sprintf("Value should be one of %s", err.Param())
		case "nooneof":
			errorMsg = fmt.Sprintf("Value should not be one of %s", err.Param())
		case "len":
			errorMsg = fmt.Sprintf("Length should be equal to %s", err.Param())
		case "unique":
			errorMsg = "Value must not contain duplicate values"
		default:
			errorMsg = "This field is invalid"
		}
	}
	return
}

func Validate(obj any, rules map[string]string) string {
	objType := reflect.TypeOf(obj)
	fieldErrors := make(map[string]string)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}
	newObj := reflect.New(objType).Elem().Interface()
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterStructValidationMapRules(rules, newObj)
	if err := validate.Struct(obj); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			fieldErrors[getFieldName(obj, e.StructField())] = GetErrorMsgForField(obj, e)
		}
	}
	// return strings.Join(fieldErrors, ", ")
	jsonErrors, _ := json.Marshal(fieldErrors)
	return string(jsonErrors)
}
