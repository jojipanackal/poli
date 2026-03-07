package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var scanner *bufio.Scanner

func init() {
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for large inputs
}

// PromptString asks the user for a single-line string input.
// If defaultVal is non-empty, it's shown and used when the user presses Enter.
func PromptString(label, defaultVal string) string {
	if defaultVal != "" {
		cyan.Printf("  %s ", label)
		dim.Printf("(%s)", defaultVal)
		fmt.Print(": ")
	} else {
		cyan.Printf("  %s: ", label)
	}

	scanner.Scan()
	val := strings.TrimSpace(scanner.Text())
	if val == "" {
		return defaultVal
	}
	return val
}

// PromptSelect shows a numbered list and returns the selected option.
func PromptSelect(label string, options []string) (int, string) {
	cyan.Printf("\n  %s\n\n", label)

	for i, opt := range options {
		fmt.Printf("    %d) %s\n", i+1, opt)
	}

	fmt.Println()
	for {
		cyan.Print("  Select: ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		var choice int
		if _, err := fmt.Sscanf(input, "%d", &choice); err == nil {
			if choice >= 1 && choice <= len(options) {
				return choice - 1, options[choice-1]
			}
		}
		Warning("Invalid selection, try again")
	}
}

// PromptMultiline collects multiple lines of input until an empty line.
func PromptMultiline(label string) string {
	cyan.Printf("  %s ", label)
	dim.Println("(empty line to finish)")

	var lines []string
	for {
		fmt.Print("    ")
		scanner.Scan()
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// PromptConfirm asks a yes/no question.
func PromptConfirm(label string) bool {
	cyan.Printf("  %s ", label)
	dim.Print("(y/N): ")

	scanner.Scan()
	input := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return input == "y" || input == "yes"
}

// PromptMethod shows a method selector.
func PromptMethod(defaultMethod string) string {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	if defaultMethod == "" {
		defaultMethod = "GET"
	}

	cyan.Printf("  Method ")
	dim.Printf("(%s)", defaultMethod)
	fmt.Print(": ")

	scanner.Scan()
	input := strings.TrimSpace(strings.ToUpper(scanner.Text()))

	if input == "" {
		return defaultMethod
	}

	// Validate
	for _, m := range methods {
		if input == m {
			return input
		}
	}

	Warning(fmt.Sprintf("Unknown method %q, using %s", input, defaultMethod))
	return defaultMethod
}

// PromptHeaders collects headers as key:value pairs.
func PromptHeaders(existing []struct{ Key, Value string }) []struct{ Key, Value string } {
	cyan.Print("  Headers ")
	dim.Println("(key:value, empty to finish)")

	var headers []struct{ Key, Value string }

	// Show existing headers first
	for _, h := range existing {
		dim.Printf("    [existing] %s: %s\n", h.Key, h.Value)
		headers = append(headers, h)
	}

	for {
		fmt.Print("    ")
		scanner.Scan()
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			Warning("Format: key:value")
			continue
		}

		headers = append(headers, struct{ Key, Value string }{
			Key:   strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
		})
	}

	return headers
}
