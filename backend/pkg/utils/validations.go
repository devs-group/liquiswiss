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
	validate.RegisterValidation("exitDateGTEEntryDate", validateExitDate)
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
	layout := "2006-01-02" // Assuming the date format is YYYY-MM-DD
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		return false // Invalid start date format
	}

	if endDateStr == "" {
		return true // If EndDate is empty, we allow this as it's omitempty
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		return false // Invalid end date format
	}

	// Check if EndDate is after or equal to StartDate
	return endDate.After(startDate) || endDate.Equal(startDate)
}

func validateExitDate(fl validator.FieldLevel) bool {
	entryDateField := fl.Parent().FieldByName("EntryDate")
	// exitDateStr is always a pointer
	exitDateStr := fl.Field().String()

	// entryDateField can be a pointer or not so it needs special treatment
	var entryDateStr string
	if entryDateField.Kind() == reflect.Ptr {
		// If it's a pointer, check if it's nil
		if entryDateField.IsNil() {
			// If EntryDate is not provided, ExitDate can't be correct
			return false
		}
		// Dereference the pointer to get the actual value
		entryDateStr = entryDateField.Elem().String()
	} else {
		// If it's not a pointer, get the value directly
		entryDateStr = entryDateField.String()
	}

	if entryDateStr == "" && exitDateStr == "" {
		return true
	} else if entryDateStr == "" && exitDateStr != "" {
		return false
	}

	// Parse start and end dates
	layout := "2006-01-02" // Assuming the date format is YYYY-MM-DD
	entryDate, err := time.Parse(layout, entryDateStr)
	if err != nil {
		return false // Invalid start date format
	}

	if exitDateStr == "" {
		return true // If ExitDate is empty, we allow this as it's omitempty
	}

	exitDate, err := time.Parse(layout, exitDateStr)
	if err != nil {
		return false // Invalid end date format
	}

	// Check if ExitDate is after or equal to EntryDate
	return exitDate.After(entryDate) || exitDate.Equal(entryDate)
}
