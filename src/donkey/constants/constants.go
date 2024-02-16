package constants

import "fmt"

const (
	colorRed   = "\033[31m"
	colorBlue  = "\033[34m"
	colorReset = "\033[0m"
)

const LangName = "donkey"

const ReplPrompt = "\u001b[33m💡 >> \u001b[0m"

const ParserErrorPrompt = "🚨 parser errors:\n"
const CompilerErrorPrompt = "🚨 compiler errors:\n"

func Blue(text string, args ...interface{}) string {
	return fmt.Sprintf(colorBlue+text+colorReset, args...)
}

func Red(text string, args ...interface{}) string {
	return fmt.Sprintf(colorRed+text+colorReset, args...)
}
