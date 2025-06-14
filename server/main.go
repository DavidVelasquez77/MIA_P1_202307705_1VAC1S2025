package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"server/analyzer"
	"strings"
)

var outcome string

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Bienvenido!")
	for {
		fmt.Print(">>> ")

		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if input == "exit" {
			break
		} else if strings.HasPrefix(input, "#") {
			continue
		}
		msg, err := analyzer.Analyzer(input)
		if err != nil {
			outcome += fmt.Sprintf("Error: %v\n", err)
			continue
		} else {
			outcome += fmt.Sprintf("%v\n", msg)
		}
	}
	clearConsole()
	fmt.Println("================================== FIN DE EJECUCION ==================================")
	fmt.Println(outcome)

}

func clearConsole() {
	cmd := exec.Command("clear") // en Windows sería "cls"
	cmd.Stdout = os.Stdout
	cmd.Run()
}
