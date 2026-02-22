package domain

type Language string

const (
	LanguageEnglish Language = "English"
	LanguageSpanish Language = "Spanish"
	LanguageFrench  Language = "French"
	LanguageGerman  Language = "German"
	LanguageItalian Language = "Italian"
	LanguageRussian Language = "Russian"
)

type Status string

const (
	StatusAvailable Status = "Available"
	StatusReserved  Status = "Reserved"
	StatusPurchased Status = "Purchased"
)
