package dto

type LanguageEnum string

const (
	LanguageEnglish LanguageEnum = "English"
	LanguageSpanish LanguageEnum = "Spanish"
	LanguageFrench  LanguageEnum = "French"
	LanguageGerman  LanguageEnum = "German"
	LanguageItalian LanguageEnum = "Italian"
	LanguageRussian LanguageEnum = "Russian"
)

type StatusEnum string

const (
	StatusAvailable StatusEnum = "Available"
	StatusReserved  StatusEnum = "Reserved"
	StatusPurchased StatusEnum = "Purchased"
)
