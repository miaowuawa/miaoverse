package validate

import "github.com/go-playground/validator/v10"

func InitialValidator(validator *validator.Validate) error {
	err := validator.RegisterValidation("isDigit", isDigit)
	err = validator.RegisterValidation("isPhone", isPhone)
	err = validator.RegisterValidation("isUUIDv4", isUUIDV4)
	if err != nil {
		return err
	}
	return nil
}
