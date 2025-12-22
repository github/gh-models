// Package i18n provides internationalization support for BrainSait CLI
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed locales/*.json
var localesFS embed.FS

// Locale represents a language locale
type Locale struct {
	Code       string            `json:"code"`
	Name       string            `json:"name"`
	NativeName string            `json:"native_name"`
	Direction  string            `json:"direction"` // "ltr" or "rtl"
	Messages   map[string]string `json:"messages"`
}

// I18n manages internationalization
type I18n struct {
	currentLocale string
	locales       map[string]*Locale
	fallback      string
}

// NewI18n creates a new I18n instance
func NewI18n() *I18n {
	i := &I18n{
		locales:  make(map[string]*Locale),
		fallback: "en",
	}

	// Load embedded locales
	i.loadEmbeddedLocales()

	// Detect system locale
	i.detectLocale()

	return i
}

// loadEmbeddedLocales loads all locale files from embedded filesystem
func (i *I18n) loadEmbeddedLocales() {
	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := localesFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			continue
		}

		var locale Locale
		if err := json.Unmarshal(data, &locale); err != nil {
			continue
		}

		i.locales[locale.Code] = &locale
	}
}

// detectLocale detects the system locale
func (i *I18n) detectLocale() {
	// Check environment variables
	for _, envVar := range []string{"BRAINSAIT_LANG", "LANG", "LANGUAGE", "LC_ALL", "LC_MESSAGES"} {
		if lang := os.Getenv(envVar); lang != "" {
			// Parse locale (e.g., "en_US.UTF-8" -> "en")
			lang = strings.Split(lang, ".")[0]
			lang = strings.Split(lang, "_")[0]
			lang = strings.ToLower(lang)

			if _, ok := i.locales[lang]; ok {
				i.currentLocale = lang
				return
			}
		}
	}

	// Default to English
	i.currentLocale = "en"
}

// SetLocale sets the current locale
func (i *I18n) SetLocale(code string) error {
	code = strings.ToLower(code)
	if _, ok := i.locales[code]; !ok {
		return fmt.Errorf("locale %s not found", code)
	}
	i.currentLocale = code
	return nil
}

// GetLocale returns the current locale
func (i *I18n) GetLocale() string {
	return i.currentLocale
}

// T translates a message key
func (i *I18n) T(key string, args ...interface{}) string {
	locale, ok := i.locales[i.currentLocale]
	if !ok {
		locale = i.locales[i.fallback]
	}

	if locale == nil {
		return key
	}

	msg, ok := locale.Messages[key]
	if !ok {
		// Try fallback
		if fallbackLocale, ok := i.locales[i.fallback]; ok {
			msg, ok = fallbackLocale.Messages[key]
			if !ok {
				return key
			}
		} else {
			return key
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// IsRTL returns true if current locale is right-to-left
func (i *I18n) IsRTL() bool {
	locale, ok := i.locales[i.currentLocale]
	if !ok {
		return false
	}
	return locale.Direction == "rtl"
}

// AvailableLocales returns all available locales
func (i *I18n) AvailableLocales() []Locale {
	var locales []Locale
	for _, locale := range i.locales {
		locales = append(locales, *locale)
	}
	return locales
}

// Global instance
var defaultI18n = NewI18n()

// T translates using the global instance
func T(key string, args ...interface{}) string {
	return defaultI18n.T(key, args...)
}

// SetLocale sets the locale for the global instance
func SetLocale(code string) error {
	return defaultI18n.SetLocale(code)
}

// GetLocale returns the current locale for the global instance
func GetLocale() string {
	return defaultI18n.GetLocale()
}

// IsRTL returns true if current locale is RTL
func IsRTL() bool {
	return defaultI18n.IsRTL()
}
