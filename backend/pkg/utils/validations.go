package utils

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"time"
)

var validate *validator.Validate

func InitValidator() {
	validate = validator.New()
	// Register any custom validations here
	validate.RegisterValidation("cycleRequiredIfRepeating", cycleRequiredIfRepeating)
	validate.RegisterValidation("endDateGTEStartDate", validateEndDate)
	validate.RegisterValidation("fromDateGTEToDate", validateToDate)
	validate.RegisterAlias("allowedCycles", `oneof='daily' 'weekly' 'monthly' 'quarterly' 'biannually' 'yearly'`)
}

func GetValidator() *validator.Validate {
	return validate
}

func cycleRequiredIfRepeating(fl validator.FieldLevel) bool {
	typeField := fl.Field()
	if typeField.String() == "repeating" {
		cycleField := fl.Parent().FieldByName("Cycle")
		if cycleField.IsValid() && !cycleField.IsNil() {
			cycleValue, ok := cycleField.Interface().(*string)
			return ok && *cycleValue != ""
		}
		return false
	}
	return true
}

func validateEndDate(fl validator.FieldLevel) bool {
	startDateField := fl.Parent().FieldByName("StartDate")
	// endDateStr is always a pointer
	endDateStr := fl.Field().String()

	// startDateField can be a pointer or not so it needs special treatment
	var startDateStr string
	if startDateField.Kind() == reflect.Ptr {
		// If it's a pointer, check if it's nil
		if startDateField.IsNil() {
			// If StartDate is not provided, EndDate can't be correct
			return false
		}
		// Dereference the pointer to get the actual value
		startDateStr = startDateField.Elem().String()
	} else {
		// If it's not a pointer, get the value directly
		startDateStr = startDateField.String()
	}

	if startDateStr == "" && endDateStr == "" {
		return true
	} else if startDateStr == "" && endDateStr != "" {
		return false
	}

	// Parse start and end dates
	startDate, err := time.Parse(InternalDateFormat, startDateStr)
	if err != nil {
		return false // Invalid start date format
	}

	if endDateStr == "" {
		return true // If EndDate is empty, we allow this as it's omitempty
	}

	endDate, err := time.Parse(InternalDateFormat, endDateStr)
	if err != nil {
		return false // Invalid end date format
	}

	// Check if EndDate is after or equal to StartDate
	return endDate.After(startDate) || endDate.Equal(startDate)
}

func validateToDate(fl validator.FieldLevel) bool {
	fromDateField := fl.Parent().FieldByName("FromDate")
	// toDateStr is always a pointer
	toDateStr := fl.Field().String()

	// fromDateField can be a pointer or not so it needs special treatment
	var fromDateStr string
	if fromDateField.Kind() == reflect.Ptr {
		// If it's a pointer, check if it's nil
		if fromDateField.IsNil() {
			// If FromDate is not provided, ToDate can't be correct
			return false
		}
		// Dereference the pointer to get the actual value
		fromDateStr = fromDateField.Elem().String()
	} else {
		// If it's not a pointer, get the value directly
		fromDateStr = fromDateField.String()
	}

	if fromDateStr == "" && toDateStr == "" {
		return true
	} else if fromDateStr == "" && toDateStr != "" {
		return false
	}

	// Parse start and end dates
	fromDate, err := time.Parse(InternalDateFormat, fromDateStr)
	if err != nil {
		return false // Invalid start date format
	}

	if toDateStr == "" {
		return true // If ToDate is empty, we allow this as it's omitempty
	}

	toDate, err := time.Parse(InternalDateFormat, toDateStr)
	if err != nil {
		return false // Invalid end date format
	}

	// Check if ToDate is after or equal to FromDate
	return toDate.After(fromDate) || toDate.Equal(fromDate)
}
