package analyzer
//hola de nuevo
import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type EXECUTE struct {
	path string
}

func ParseExecute(tokens []string) (string, error) {
	cmd := &EXECUTE{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+`)
	matches := re.FindAllString(args, -1)

	if len(matches) != len(tokens) {
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])

		switch key {
		case "-path":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.path = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	outcomeCmd, err := commandExecute(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("EXECUTE: ejecutado correctamente.\n%s", outcomeCmd), nil

}

func commandExecute(exec *EXECUTE) (string, error) {
	commands, err := getCommands(exec.path)
	if err != nil {
		return "", err
	}
	var outcome string
	for _, cmd := range commands {
		if cmd == "exit" {
			break
		} else if strings.HasPrefix(cmd, "#") {
			continue
		}
		msg, err := Analyzer(cmd)
		if err != nil {
			outcome += fmt.Sprintf("Error: %v\n", err)
			continue
		} else {
			outcome += fmt.Sprintf("%v\n", msg)
		}
	}
	return outcome, nil

}

func getCommands(path string) ([]string, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return make([]string, 0), err
	}
	return strings.Split(string(fileContent), "\n"), nil
}
