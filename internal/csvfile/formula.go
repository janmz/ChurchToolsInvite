package csvfile

import "strings"

// formulaPrefix is prepended to cells that spreadsheet programs would
// interpret as a formula (CSV/formula injection, OWASP).
const formulaPrefix = "'"

// isFormulaStart reports whether a cell would start a formula in Excel or
// LibreOffice Calc.
func isFormulaStart(value string) bool {
	if value == "" {
		return false
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return true
	}
	return false
}

// escapeFormula neutralises cells that would be evaluated as a formula.
func escapeFormula(value string) string {
	if isFormulaStart(value) {
		return formulaPrefix + value
	}
	return value
}

// unescapeFormula removes the prefix added by escapeFormula on import.
func unescapeFormula(value string) string {
	if strings.HasPrefix(value, formulaPrefix) && isFormulaStart(value[len(formulaPrefix):]) {
		return value[len(formulaPrefix):]
	}
	return value
}
