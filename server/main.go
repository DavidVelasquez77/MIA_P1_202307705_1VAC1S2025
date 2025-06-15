package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"server/analyzer"
	"server/console"
	"strings"
)

var outcome string

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Limpiar consola y mostrar bienvenida estética
	clearConsole()
	console.PrintWelcome()

	for {
		console.PrintPrompt()

		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "exit" {
			break
		} else if strings.HasPrefix(input, "#") {
			console.PrintInfo("Comentario ignorado")
			continue
		} else if input == "" {
			continue
		}

		// Mostrar comando que se va a ejecutar
		console.PrintCommand(input)

		msg, err := analyzer.Analyzer(input)
		if err != nil {
			console.PrintError(fmt.Sprintf("%v", err))
			outcome += fmt.Sprintf("❌ Error: %v\n", err)
		} else {
			console.PrintSuccess("Comando ejecutado correctamente")
			outcome += fmt.Sprintf("✅ %v\n", msg)
		}
		console.PrintSeparator()
	}

	// Mostrar resumen final con estilo
	clearConsole()
	console.PrintFinalSeparator()

	if outcome != "" {
		fmt.Println(outcome)
	} else {
		console.PrintInfo("No se ejecutaron comandos")
	}

	console.PrintGoodbye()
}

func clearConsole() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
