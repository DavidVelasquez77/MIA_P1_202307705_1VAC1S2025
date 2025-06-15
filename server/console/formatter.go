package console

import (
	"fmt"
	"strings"
)

// Códigos de color ANSI
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"
	Hidden    = "\033[8m"

	// Colores básicos
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Colores brillantes
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Colores de fondo
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

func PrintHeader(title string) {
	width := 80
	fmt.Printf("%s%s\n", Bold+BrightCyan, strings.Repeat("═", width))
	fmt.Printf("║%s%s%s║\n",
		strings.Repeat(" ", (width-len(title)-2)/2),
		Bold+BrightWhite+title+Reset+Bold+BrightCyan,
		strings.Repeat(" ", (width-len(title)-2)/2))
	fmt.Printf("%s%s\n", strings.Repeat("═", width), Reset)
}

func PrintWelcome() {
	PrintHeader("SISTEMA M.I.A - MANEJO E IMPLENTACION DE ARCHIVOS")
	fmt.Printf("%s%s🚀 ¡Bienvenido al Sistema MIA!%s\n", Bold, BrightGreen, Reset)
	fmt.Printf("%s%s💡 Escribe 'exit' para salir o '#' para comentarios%s\n", Dim, BrightYellow, Reset)
	PrintSeparator()
}

func PrintPrompt() {
	fmt.Printf("%s%s[%s%sMIA%s%s]%s%s $ %s",
		Bold, BrightBlue,
		BrightMagenta, Bold,
		Reset, Bold, BrightBlue,
		BrightGreen, Reset)
}

func PrintSuccess(message string) {
	fmt.Printf("%s%s✅ %s%s\n", Bold, BrightGreen, message, Reset)
}

func PrintError(message string) {
	fmt.Printf("%s%s❌ ERROR: %s%s\n", Bold, BrightRed, message, Reset)
}

func PrintWarning(message string) {
	fmt.Printf("%s%s⚠️  ADVERTENCIA: %s%s\n", Bold, BrightYellow, message, Reset)
}

func PrintInfo(message string) {
	fmt.Printf("%s%sℹ️  %s%s\n", Bold, BrightBlue, message, Reset)
}

func PrintCommand(command string) {
	fmt.Printf("%s%s▶️  Ejecutando: %s%s%s%s\n",
		Bold, BrightMagenta, BrightWhite, command, Reset, Reset)
}

func PrintSeparator() {
	fmt.Printf("%s%s%s%s\n", Dim, BrightCyan, strings.Repeat("─", 80), Reset)
}

func PrintFinalSeparator() {
	fmt.Printf("\n%s%s", Bold, BrightCyan)
	fmt.Printf("╔%s╗\n", strings.Repeat("═", 78))
	fmt.Printf("║%s║\n", centerText("RESUMEN DE EJECUCIÓN", 78))
	fmt.Printf("╚%s╝\n", strings.Repeat("═", 78))
	fmt.Printf("%s", Reset)
}

func PrintGoodbye() {
	fmt.Printf("\n%s%s👋 ¡Gracias por usar el Sistema M.I.A!%s\n", Bold, BrightGreen, Reset)
	fmt.Printf("%s%s🎯 Operaciones completadas exitosamente%s\n", Bold, BrightBlue, Reset)
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
}
