package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// printBodyAsTable renders JSON body as a clean table.
// Respects opts.RowIndex, opts.SearchQuery, opts.ExpandKey.
func printBodyAsTable(body, indent string, opts ResponseOptions) {
	body = strings.TrimSpace(body)
	if body == "" {
		return
	}

	// Try to parse as JSON
	var raw interface{}
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		// Not JSON — print raw
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			fmt.Printf("%s%s\n", indent, line)
		}
		return
	}

	// Handle full raw JSON body if no filtering is applied
	if opts.ShowRaw && opts.RowIndex <= 0 && opts.SearchQuery == "" && opts.ExpandKey == "" {
		printRawBody(body, indent)
		return
	}

	switch v := raw.(type) {
	case map[string]interface{}:
		// If searching/subsetting but root is object, look for arrays inside
		if (opts.SearchQuery != "" || opts.RowIndex > 0) && opts.ExpandKey == "" {
			// Find first array field
			var arrayKey string
			for k, val := range v {
				if _, ok := val.([]interface{}); ok {
					arrayKey = k
					break
				}
			}
			if arrayKey != "" {
				printObjectTable(v, indent, opts)
				return
			}
		}
		printObjectTable(v, indent, opts)
	case []interface{}:
		printArrayTable(v, indent, opts)
	default:
		fmt.Printf("%s%v\n", indent, v)
	}
}

// printArrayRow expands a single row from an array response.
func printArrayRow(arr []interface{}, indent string, opts ResponseOptions) {
	idx := opts.RowIndex - 1 // convert to 0-indexed
	if idx < 0 || idx >= len(arr) {
		Error(fmt.Sprintf("Row %d out of range (1-%d)", opts.RowIndex, len(arr)))
		return
	}

	item := arr[idx]
	dim.Printf("%s[%d of %d]\n", indent, opts.RowIndex, len(arr))
	fmt.Println()

	switch v := item.(type) {
	case map[string]interface{}:
		if opts.ShowRaw {
			printRawBody(rawJSON(v), indent)
		} else {
			printObjectTable(v, indent, opts)
		}
	default:
		fmt.Printf("%s%v\n", indent, v)
	}
}

// filterArray filters array items by a "key=value" query.
// Supports partial, case-insensitive matching on values.
func filterArray(arr []interface{}, query string) []interface{} {
	parts := strings.SplitN(query, "=", 2)

	var key, value string
	if len(parts) == 2 {
		key = strings.TrimSpace(parts[0])
		value = strings.ToLower(strings.TrimSpace(parts[1]))
	} else {
		// No key specified — search all values
		value = strings.ToLower(strings.TrimSpace(query))
	}

	var results []interface{}
	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if key != "" {
			// Search specific key
			v, exists := obj[key]
			if !exists {
				// Try case-insensitive key match
				for k, val := range obj {
					if strings.EqualFold(k, key) {
						v = val
						exists = true
						break
					}
				}
			}
			if exists && matchesValue(v, value) {
				results = append(results, item)
			}
		} else {
			// Search all values
			for _, v := range obj {
				if matchesValue(v, value) {
					results = append(results, item)
					break
				}
			}
		}
	}
	return results
}

// matchesValue checks if a value matches the search query (case-insensitive, partial).
func matchesValue(v interface{}, query string) bool {
	str := fmt.Sprintf("%v", v)
	return strings.Contains(strings.ToLower(str), query)
}

// printObjectTable renders a JSON object as a two-column key/value table.
func printObjectTable(obj map[string]interface{}, indent string, opts ResponseOptions) {
	// If expanding a specific key, show just that key's content
	if opts.ExpandKey != "" {
		val, ok := obj[opts.ExpandKey]
		if !ok {
			// Try case-insensitive
			for k, v := range obj {
				if strings.EqualFold(k, opts.ExpandKey) {
					val = v
					ok = true
					break
				}
			}
		}
		if !ok {
			fmt.Printf("%s", indent)
			red.Printf("Key %q not found\n", opts.ExpandKey)
			return
		}

		fmt.Printf("%s", indent)
		cyan.Printf("%s:\n", opts.ExpandKey)
		fmt.Println()

		switch nested := val.(type) {
		case map[string]interface{}:
			printObjectTable(nested, indent, ResponseOptions{RequestName: opts.RequestName})
		case []interface{}:
			printArrayTable(nested, indent, opts)
		default:
			fmt.Printf("%s%v\n", indent, nested)
		}
		return
	}

	// Auto-expand search/row if provided and root is object
	if opts.SearchQuery != "" || opts.RowIndex > 0 {
		var arrayKey string
		// Pick first plausible array (usually 'items', 'data', or the only array)
		for _, k := range []string{"providers", "items", "data", "results", "list"} {
			if _, ok := obj[k].([]interface{}); ok {
				arrayKey = k
				break
			}
		}
		// If none of those, just first array
		if arrayKey == "" {
			for k, v := range obj {
				if _, ok := v.([]interface{}); ok {
					arrayKey = k
					break
				}
			}
		}

		if arrayKey != "" {
			opts.ExpandKey = arrayKey
			printObjectTable(obj, indent, opts)
			return
		}
	}

	// Collect keys
	var keys []string
	for k := range obj {
		keys = append(keys, k)
	}

	// Find max key width for alignment
	maxKeyLen := 0
	for _, k := range keys {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}
	if maxKeyLen > 24 {
		maxKeyLen = 24
	}

	// Print separator
	tableWidth := maxKeyLen + 4 + 50
	fmt.Printf("%s%s\n", indent, dim.Sprint(strings.Repeat("─", tableWidth)))

	for _, k := range keys {
		v := obj[k]
		paddedKey := fmt.Sprintf("%-*s", maxKeyLen, k)

		switch nested := v.(type) {
		case map[string]interface{}:
			count := len(nested)
			fmt.Printf("%s", indent)
			bold.Printf("%s", paddedKey)
			fmt.Print("    ")
			dim.Printf("{object: %d keys}", count)
			fmt.Print("  ")
			dim.Printf("→ poli last \"%s\" --expand %s\n", opts.RequestName, k)

		case []interface{}:
			count := len(nested)
			fmt.Printf("%s", indent)
			bold.Printf("%s", paddedKey)
			fmt.Print("    ")
			dim.Printf("[array: %d items]", count)
			fmt.Print("  ")
			dim.Printf("→ poli last \"%s\" --expand %s\n", opts.RequestName, k)

		default:
			strVal := formatValue(v)
			// Truncate long values
			if utf8.RuneCountInString(strVal) > 60 {
				strVal = string([]rune(strVal)[:57]) + "..."
			}
			fmt.Printf("%s", indent)
			bold.Printf("%s", paddedKey)
			fmt.Printf("    %s\n", strVal)
		}
	}

	fmt.Printf("%s%s\n", indent, dim.Sprint(strings.Repeat("─", tableWidth)))
}

// printArrayTable renders a JSON array as a multi-column table.
func printArrayTable(arr []interface{}, indent string, opts ResponseOptions) {
	if len(arr) == 0 {
		fmt.Printf("%s%s\n", indent, dim.Sprint("(empty array)"))
		return
	}

	// Handle --row: expand a single item from the array
	if opts.RowIndex > 0 {
		printArrayRow(arr, indent, opts)
		return
	}

	// Handle --search: filter array items
	if opts.SearchQuery != "" {
		filtered := filterArray(arr, opts.SearchQuery)
		if len(filtered) == 0 {
			Warning(fmt.Sprintf("No items matching %q", opts.SearchQuery))
			return
		}
		Info(fmt.Sprintf("Found %d matching items", len(filtered)))
		fmt.Println()
		if opts.ShowRaw {
			for i, item := range filtered {
				dim.Printf("%s[Match %d]\n", indent, i+1)
				printRawBody(rawJSON(item), indent)
				fmt.Println()
			}
			return
		}
		arr = filtered
	}

	if opts.ShowRaw {
		printRawBody(rawJSON(arr), indent)
		return
	}

	// Check if it's an array of objects with consistent keys
	if isObjectArray(arr) {
		printObjectArrayTable(arr, indent, opts.RequestName)
		return
	}

	// Simple array — just list items
	for i, item := range arr {
		switch v := item.(type) {
		case map[string]interface{}:
			fmt.Printf("%s", indent)
			dim.Printf("[%d] ", i)
			dim.Printf("{object: %d keys}\n", len(v))
		case []interface{}:
			fmt.Printf("%s", indent)
			dim.Printf("[%d] ", i)
			dim.Printf("[array: %d items]\n", len(v))
		default:
			fmt.Printf("%s", indent)
			dim.Printf("[%d] ", i)
			fmt.Printf("%s\n", formatValue(item))
		}
	}
}

// printObjectArrayTable renders an array of objects as a proper multi-column table.
func printObjectArrayTable(arr []interface{}, indent, requestName string) {
	// Collect all keys from all objects
	keySet := make(map[string]bool)
	var keys []string
	for _, item := range arr {
		if obj, ok := item.(map[string]interface{}); ok {
			for k := range obj {
				if !keySet[k] {
					keySet[k] = true
					keys = append(keys, k)
				}
			}
		}
	}

	// Calculate column widths
	colWidths := make(map[string]int)
	for _, k := range keys {
		colWidths[k] = len(k)
	}

	maxItems := len(arr)
	if maxItems > 10 {
		maxItems = 10
	}

	for i := 0; i < maxItems; i++ {
		if obj, ok := arr[i].(map[string]interface{}); ok {
			for _, k := range keys {
				val := formatValue(obj[k])
				if len(val) > colWidths[k] {
					colWidths[k] = len(val)
				}
			}
		}
	}

	// Cap column widths
	for k, w := range colWidths {
		if w > 30 {
			colWidths[k] = 30
		}
	}

	// Filter out columns that are objects/arrays for the table view
	var tableCols []string
	for _, k := range keys {
		scalarCount := 0
		for i := 0; i < maxItems; i++ {
			if obj, ok := arr[i].(map[string]interface{}); ok {
				v := obj[k]
				switch v.(type) {
				case map[string]interface{}, []interface{}:
					// complex type
				default:
					scalarCount++
				}
			}
		}
		if scalarCount > maxItems/2 {
			tableCols = append(tableCols, k)
		}
	}

	if len(tableCols) == 0 {
		tableCols = keys
	}

	// Print header
	var headerParts []string
	for _, k := range tableCols {
		headerParts = append(headerParts, fmt.Sprintf("%-*s", colWidths[k], k))
	}
	fmt.Printf("%s%s\n", indent, bold.Sprint(strings.Join(headerParts, "  ")))

	// Separator
	var sepParts []string
	for _, k := range tableCols {
		sepParts = append(sepParts, strings.Repeat("─", colWidths[k]))
	}
	fmt.Printf("%s%s\n", indent, dim.Sprint(strings.Join(sepParts, "──")))

	// Rows
	for i := 0; i < maxItems; i++ {
		if obj, ok := arr[i].(map[string]interface{}); ok {
			var rowParts []string
			for _, k := range tableCols {
				val := formatValue(obj[k])
				if utf8.RuneCountInString(val) > colWidths[k] {
					val = string([]rune(val)[:colWidths[k]-3]) + "..."
				}
				rowParts = append(rowParts, fmt.Sprintf("%-*s", colWidths[k], val))
			}
			fmt.Printf("%s%s\n", indent, strings.Join(rowParts, "  "))
		}
	}

	remaining := len(arr) - maxItems
	if remaining > 0 {
		fmt.Printf("%s%s\n", indent, dim.Sprintf("... and %d more items", remaining))
		fmt.Printf("%s%s\n", indent, dim.Sprintf("→ poli last \"%s\" --row N    to expand a row", requestName))
		fmt.Printf("%s%s\n", indent, dim.Sprintf("→ poli last \"%s\" --search key=value", requestName))
	}
}

func isObjectArray(arr []interface{}) bool {
	if len(arr) == 0 {
		return false
	}
	objCount := 0
	for _, item := range arr {
		if _, ok := item.(map[string]interface{}); ok {
			objCount++
		}
	}
	return objCount == len(arr)
}

func formatValue(v interface{}) string {
	if v == nil {
		return dim.Sprint("null")
	}
	switch val := v.(type) {
	case string:
		// Replace newlines with spaces to keep table rows on one line
		return strings.ReplaceAll(strings.ReplaceAll(val, "\n", " "), "\r", "")
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case bool:
		if val {
			return green.Sprint("true")
		}
		return red.Sprint("false")
	case json.Number:
		return val.String()
	case map[string]interface{}:
		return dim.Sprintf("{object: %d keys}", len(val))
	case []interface{}:
		return dim.Sprintf("[array: %d]", len(val))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func rawJSON(v interface{}) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("  ", "  ")
	enc.Encode(v)
	return strings.TrimSpace(buf.String())
}
