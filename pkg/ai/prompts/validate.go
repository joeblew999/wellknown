package prompts

import "fmt"

// TransactionValidation generates a prompt for validating banking transactions
func TransactionValidation(amount float64, description, category, currency string) string {
	return fmt.Sprintf(`You are a banking transaction validator. Analyze the following transaction and determine if it appears valid and consistent.

Transaction Details:
- Amount: %.2f %s
- Description: %s
- Category: %s

Please analyze:
1. Does the description match the category?
2. Is the amount reasonable for this type of transaction?
3. Are there any suspicious patterns (e.g., unusual characters, inconsistent formatting)?
4. Does the description appear legitimate?

Respond in JSON format:
{
  "valid": true/false,
  "confidence": 0.0-1.0,
  "reason": "Brief explanation",
  "suggestions": ["Optional suggestions for corrections"]
}

Be strict but fair. Flag anything that looks suspicious or inconsistent.`, amount, currency, description, category)
}

// RecordValidation generates a generic validation prompt for any PocketBase record
func RecordValidation(collection string, fields map[string]interface{}) string {
	prompt := fmt.Sprintf(`You are a data validator for a %s record. Analyze the following fields for consistency and validity:

Fields:
`, collection)

	for key, value := range fields {
		prompt += fmt.Sprintf("- %s: %v\n", key, value)
	}

	prompt += `
Please check:
1. Field consistency (do values make sense together?)
2. Data format validity
3. Potential data quality issues
4. Any suspicious patterns

Respond in JSON format:
{
  "valid": true/false,
  "confidence": 0.0-1.0,
  "issues": ["List of issues found"],
  "suggestions": ["Suggestions for corrections"]
}
`

	return prompt
}
