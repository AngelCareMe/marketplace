package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(s interface{}) error
	ValidateStruct(s interface{}) []ValidationError
}

type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

type customValidator struct {
	validator *validator.Validate
}

func NewValidator() Validator {
	validate := validator.New()
	
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	return &customValidator{validator: validate}
}

func (cv *customValidator) Validate(s interface{}) error {
	return cv.validator.Struct(s)
}

func (cv *customValidator) ValidateStruct(s interface{}) []ValidationError {
	err := cv.validator.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors []ValidationError
	
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErr {
			ve := ValidationError{
				Field: fieldError.Field(),
				Tag:   fieldError.Tag(),
				Value: fmt.Sprintf("%v", fieldError.Value()),
			}
			
			switch fieldError.Tag() {
			case "required":
				ve.Message = fmt.Sprintf("Field %s is required", fieldError.Field())
			case "email":
				ve.Message = fmt.Sprintf("Field %s must be a valid email", fieldError.Field())
			case "min":
				ve.Message = fmt.Sprintf("Field %s must be at least %s characters", fieldError.Field(), fieldError.Param())
			case "max":
				ve.Message = fmt.Sprintf("Field %s must be at most %s characters", fieldError.Field(), fieldError.Param())
			case "len":
				ve.Message = fmt.Sprintf("Field %s must be exactly %s characters", fieldError.Field(), fieldError.Param())
			case "oneof":
				ve.Message = fmt.Sprintf("Field %s must be one of: %s", fieldError.Field(), fieldError.Param())
			default:
				ve.Message = fmt.Sprintf("Field %s failed validation for tag %s", fieldError.Field(), fieldError.Tag())
			}
			
			validationErrors = append(validationErrors, ve)
		}
	}
	
	return validationErrors
}
