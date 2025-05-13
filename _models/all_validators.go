package models

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate

	// Регулярные выражения из ваших CHECK-ограничений
	nameRegex     = regexp.MustCompile(`^[A-Za-zА-Яа-яЁё\s-]+$`)
	emailRegex    = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$`)
	notEmptyRegex = regexp.MustCompile(`\S`)
)

func init() {
	validate = validator.New()

	// Регистрируем все кастомные валидации
	_ = validate.RegisterValidation("nameFormat", validateNameFormat)
	_ = validate.RegisterValidation("emailFormat", validateEmailFormat)
	_ = validate.RegisterValidation("notEmpty", validateNotEmpty)
	_ = validate.RegisterValidation("birthDate", validateBirthDate)
	// _ = validate.RegisterValidation("duration", validateDuration)
	_ = validate.RegisterValidation("rating", validateRating)
	_ = validate.RegisterValidation("ageLimit", validateAgeLimit)
	_ = validate.RegisterValidation("revenue", validateRevenue)
	_ = validate.RegisterValidation("language", validateLanguage)
	_ = validate.RegisterValidation("capacity", validateCapacity)
	_ = validate.RegisterValidation("ticketStatus", validateTicketStatus)
	// _ = validate.RegisterValidation("adminNotBlocked", validateAdminNotBlocked)
}

// Кастомные валидаторы
func validateNameFormat(fl validator.FieldLevel) bool {
	// Соответствует: CHECK (name ~ '^[A-Za-zА-Яа-яЁё\s-]+$' AND name ~ '\S')
	return nameRegex.MatchString(fl.Field().String()) &&
		notEmptyRegex.MatchString(fl.Field().String())
}

func validateEmailFormat(fl validator.FieldLevel) bool {
	// Соответствует: CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$')
	return emailRegex.MatchString(fl.Field().String())
}

func validateNotEmpty(fl validator.FieldLevel) bool {
	// Соответствует: CHECK (description ~ '\S')
	return notEmptyRegex.MatchString(fl.Field().String())
}

func validateBirthDate(fl validator.FieldLevel) bool {
	// Соответствует: CHECK (birth_date <= CURRENT_DATE AND birth_date >= CURRENT_DATE - INTERVAL '100 years')
	birthDate := fl.Field().Interface().(time.Time)
	now := time.Now()
	hundredYearsAgo := now.AddDate(-100, 0, 0)
	return !birthDate.After(now) && !birthDate.Before(hundredYearsAgo)
}

// func validateDuration(fl validator.FieldLevel) bool {
// 	// Соответствует: CHECK (duration >= '00:00:00')
// 	durationStr := fl.Field().String()
// 	dur, err := time.Parse("15:04:05", durationStr)
// 	if err != nil {
// 		return false
// 	}
// 	return dur >= time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
// }

func validateRating(fl validator.FieldLevel) bool {
	// Соответствует: CHECK (rating >= 0 AND rating <= 10)
	rating := fl.Field().Float()
	return rating >= 0 && rating <= 10
}

func validateAgeLimit(fl validator.FieldLevel) bool {
	// Соответствует: CHECK (age_limit IN (0, 6, 12, 16, 18))
	ageLimit := fl.Field().Int()
	return ageLimit == 0 || ageLimit == 6 || ageLimit == 12 || ageLimit == 16 || ageLimit == 18
}

func validateRevenue(fl validator.FieldLevel) bool {
	// Соответствует: CHECK (box_office_revenue >= 0)
	return fl.Field().Float() >= 0
}

func validateLanguage(fl validator.FieldLevel) bool {
	// Соответствует ENUM ('English', 'Spanish', 'French', 'German', 'Italian', 'Русский')
	lang := fl.Field().String()
	switch lang {
	case "English", "Spanish", "French", "German", "Italian", "Русский":
		return true
	default:
		return false
	}
}

func validateCapacity(fl validator.FieldLevel) bool {
	// Соответствует: CHECK (capacity >= 0)
	return fl.Field().Int() >= 0
}

func validateTicketStatus(fl validator.FieldLevel) bool {
	// Соответствует ENUM ('Purchased', 'Reserved', 'Available')
	status := fl.Field().String()
	switch status {
	case "Purchased", "Reserved", "Available":
		return true
	default:
		return false
	}
}

// func validateAdminNotBlocked(sl validator.StructLevel) {
// 	// Соответствует: CHECK (NOT (is_blocked AND is_admin))
// 	user := sl.Current().Interface().(UserData)
// 	if user.IsBlocked && user.IsAdmin {
// 		sl.ReportError(user.IsBlocked, "is_blocked", "IsBlocked", "adminNotBlocked", "")
// 	}
// }
