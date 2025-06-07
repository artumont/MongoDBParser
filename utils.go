package mongoparser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Parses JSON-like strings with JavaScript syntax
func (p *Parser) parseJSONLikeString(input string, target interface{}) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("empty input")
	}

	// Convert JavaScript-style object notation to valid JSON
	// Handle simple cases first
	input = p.normalizeJavaScriptObject(input)

	// Try to unmarshal as JSON
	return json.Unmarshal([]byte(input), target)
}

// Normalizes JavaScript object notation to JSON
func (p *Parser) normalizeJavaScriptObject(input string) string {
	// Handle simple cases for MongoDB operations
	// Convert single quotes to double quotes first
	input = strings.ReplaceAll(input, "'", `"`)

	// Remove trailing commas that are invalid in JSON
	input = p.removeTrailingCommas(input)

	// Simple approach: use regex-like pattern matching for common MongoDB syntax
	// Handle patterns like { key: value } -> { "key": value }

	// For simple cases like index specifications
	if strings.Contains(input, "{") && strings.Contains(input, ":") {
		// This is likely a simple object, try to add quotes around unquoted keys
		return p.addQuotesToKeys(input)
	}

	return input
}

// Adds quotes around unquoted object keys
func (p *Parser) addQuotesToKeys(input string) string {
	result := ""
	inQuotes := false
	i := 0

	for i < len(input) {
		char := input[i]

		if char == '"' {
			inQuotes = !inQuotes
			result += string(char)
			i++
			continue
		}

		if inQuotes {
			result += string(char)
			i++
			continue
		}

		// Look for unquoted identifiers followed by colon
		if isAlphaStart(rune(char)) {
			// Find the end of the identifier
			keyStart := i
			for i < len(input) && (isAlphaNum(rune(input[i])) || input[i] == '_') {
				i++
			}
			key := input[keyStart:i]

			// Skip whitespace
			for i < len(input) && (input[i] == ' ' || input[i] == '\t') {
				i++
			}

			// Check if followed by colon
			if i < len(input) && input[i] == ':' {
				// This is an unquoted key, add quotes
				result += `"` + key + `"`
			} else {
				// Not a key, just add the identifier as is
				result += key
			}
		} else {
			result += string(char)
			i++
		}
	}

	return result
}

// Helper function for character checking
func isAlphaStart(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' || char == '$'
}

// Helper function for character checking
func isAlphaNum(char rune) bool {
	return isAlphaStart(char) || (char >= '0' && char <= '9')
}

// Splits arguments respecting nested objects
func (p *Parser) splitArguments(argsString string) []string {
	var args []string
	var current strings.Builder
	braceLevel := 0
	inQuotes := false
	var quoteChar rune

	for _, char := range argsString {
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			}
			current.WriteRune(char)
		case '{':
			if !inQuotes {
				braceLevel++
			}
			current.WriteRune(char)
		case '}':
			if !inQuotes {
				braceLevel--
			}
			current.WriteRune(char)
		case ',':
			if !inQuotes && braceLevel == 0 {
				args = append(args, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the last argument
	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}

	return args
}

// Removes trailing commas from JavaScript objects to make them valid JSON
func (p *Parser) removeTrailingCommas(input string) string {
	var result strings.Builder
	inQuotes := false
	var quoteChar rune

	for i, char := range input {
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			}
			result.WriteRune(char)
		case ',':
			if inQuotes {
				result.WriteRune(char)
			} else {
				// Look ahead to see if this comma is trailing
				j := i + 1
				for j < len(input) && (input[j] == ' ' || input[j] == '\t' || input[j] == '\n' || input[j] == '\r') {
					j++
				}
				// If the next non-whitespace character is } or ], this is a trailing comma
				if j < len(input) && (input[j] == '}' || input[j] == ']') {
					// Skip the trailing comma
					continue
				} else {
					result.WriteRune(char)
				}
			}
		default:
			result.WriteRune(char)
		}
	}

	return result.String()
}
