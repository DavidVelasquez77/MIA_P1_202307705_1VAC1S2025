package commands

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"server/stores"
	"strings"
)

type RMDISK struct {
	path string
}

func ParseRmdisk(tokens []string) (string, error) {
	cmd := &RMDISK{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parametro invalido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-path":
			if value == "'" {
				return "", errors.New("el path no puede venir vacio")
			}
			cmd.path = value
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parametros requeridos: -path")
	}

	err := commandRmdisk(cmd)
	if err != nil {
		return "", err
	}
	stores.DeleteMountedPartitions(cmd.path)
	return fmt.Sprintf("RMDISK: %s eliminado exitosamente", cmd.path), nil

}

func commandRmdisk(rmdisk *RMDISK) error {

	// stores.DeleteMountedPartitions(rmdisk.path)

	if !fileExists(rmdisk.path) {
		return fmt.Errorf("el archivo no existe en el path solicitado")
	}

	err := os.Remove(rmdisk.path)
	if err != nil {
		return fmt.Errorf("error al eliminar disco con path %s", rmdisk.path)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err) //true si existe
}
