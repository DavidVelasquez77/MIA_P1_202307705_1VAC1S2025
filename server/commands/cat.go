package commands

import (
	"errors"
	"fmt"
	"regexp"
	stores "server/stores"
	utils "server/utils"
	"strconv"
	"strings"
)
//hola
type CAT struct {
	files map[int]string
}

func ParseCat(tokens []string) (string, error) {
	cmd := &CAT{}
	cmd.files = make(map[int]string)

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-file\d+="[^"]+"|-file\d+=[^\s]+`)
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
		commandKey := regexp.MustCompile("[a-zA-Z-]+")
		numberKey := regexp.MustCompile("[0-9]+")

		letters := commandKey.FindString(key)
		number := numberKey.FindString(key)

		switch letters {
		case "-file":
			numberFile, err := strconv.Atoi(number)
			if err != nil || numberFile <= 0 {
				return "", errors.New("el numero de fileN debe ser mayor a 0")
			}
			if value == "" {
				return "", errors.New("el fileN no puede estar vacio")
			}
			cmd.files[numberFile] = value
		}
	}

	if len(cmd.files) == 0 {
		return "", errors.New("falta al menos un parametro requerido: -fileN")
	}

	// Logica de Cat
	content, err := commandCat(cmd)
	if err != nil {
		return "", err
	}
	fmt.Println(content)

	return content, nil

}

func commandCat(cat *CAT) (string, error) {
	// Tomar en cuenta que el idPartition correspondara al id actual en el q este el usuario
	var result string
	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(stores.LogedIdPartition)
	if err != nil {
		return "", err
	}
	for _, pathToGetInfo := range cat.files {
		parentDirs, destDir := utils.GetParentDirectories(pathToGetInfo)

		content, err := partitionSuperblock.ContentFromFileCat(partitionPath, 0, parentDirs, destDir)
		if err != nil {
			return "", err
		}
		result += content + "\n"
	}
	return result, nil
}
