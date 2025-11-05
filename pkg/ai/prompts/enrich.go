package prompts

import "fmt"

// TransactionCategorization generates a prompt for categorizing transactions
func TransactionCategorization(description string) string {
	return fmt.Sprintf(`You are a financial transaction categorizer. Analyze the following transaction description and suggest the most appropriate category.

Transaction Description: %s

Available Categories:
- Food & Dining (restaurants, groceries, coffee shops)
- Transportation (gas, parking, public transit, rideshare)
- Shopping (retail, online purchases, clothing)
- Entertainment (movies, games, subscriptions)
- Bills & Utilities (electric, water, internet, phone)
- Healthcare (doctor, pharmacy, insurance)
- Travel (hotels, flights, vacation)
- Income (salary, refunds, transfers in)
- Other

Respond in JSON format:
{
  "category": "The most appropriate category",
  "confidence": 0.0-1.0,
  "subcategory": "Optional more specific subcategory",
  "merchant": "Extracted merchant name if identifiable",
  "reasoning": "Brief explanation of why this category fits"
}

Be specific and confident in your categorization.`, description)
}

// TransactionEnrichment generates a prompt for enriching transaction data
func TransactionEnrichment(description, currentCategory string, amount float64) string {
	return fmt.Sprintf(`You are a financial data enrichment assistant. Enhance the following transaction with additional metadata.

Transaction:
- Description: %s
- Current Category: %s
- Amount: %.2f

Please provide:
1. A cleaned, normalized description
2. Extracted merchant name (if identifiable)
3. Suggested tags for budgeting/tracking
4. Better category if current one seems wrong

Respond in JSON format:
{
  "normalized_description": "Cleaned description",
  "merchant": "Merchant name",
  "suggested_category": "Better category if needed, else same as current",
  "tags": ["tag1", "tag2"],
  "notes": "Any additional useful information"
}`, description, currentCategory, amount)
}

// RecordEnrichment generates a generic enrichment prompt for any record
func RecordEnrichment(collection string, fields map[string]interface{}) string {
	prompt := fmt.Sprintf(`You are a data enrichment assistant for a %s record. Analyze the following data and suggest enrichments:

Current Fields:
`, collection)

	for key, value := range fields {
		prompt += fmt.Sprintf("- %s: %v\n", key, value)
	}

	prompt += `
Please suggest:
1. Additional metadata that could be extracted
2. Normalized/cleaned versions of fields
3. Suggested tags or categories
4. Any derived fields that would be useful

Respond in JSON format with suggestions for new fields or improvements to existing ones.
`

	return prompt
}
