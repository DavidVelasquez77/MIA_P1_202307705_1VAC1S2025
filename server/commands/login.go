package commands

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	stores "server/stores"
	"server/structures"
	utils "server/utils"
	"strconv"
	"strings"
	"time"
)

type LOGIN struct {
	User     string
	Password string
	Id       string
}

func ParseLogin(tokens []string) (string, error) {
	cmd := &LOGIN{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[a-zA-Z0-9]+|-user="[^"]+"|-user=[^\s]+|-pass="[^"]+"|-pass=[^\s]+`)
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
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacio")
			}
			cmd.Id = value
		case "-pass":
			if value == "" {
				return "", errors.New("el password no puede estar vacio")
			}
			cmd.Password = value
		case "-user":
			if value == "" {
				return "", errors.New("el user no puede estar vacio")
			}
			cmd.User = value
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)
		}
	}
	if cmd.Password == "" {
		return "", errors.New("faltan parametros requeridos: -pass")
	}
	if cmd.User == "" {
		return "", errors.New("faltan parametros requeridos: -user")
	}
	if cmd.Id == "" {
		return "", errors.New("faltan parametros requeridos: -id")
	}

	err := CommandLogin(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("LOGIN: %s logeado exitosamente", cmd.User), nil

}

func CommandLogin(login *LOGIN) error {
	if stores.LogedIdPartition != "" {
		return errors.New("se debe realizar un logout antes de un login")
	}
	contentUsersTxt, err := getContetnUsersTxt(login.Id)
	if err != nil {
		return err
	}
	contentMatrix := getContentMatrixUsers(contentUsersTxt)

	credentials := validateInformation(login.User, login.Password, contentMatrix)
	if !credentials {
		return errors.New("credenciales invalidas en el login")
	}
	stores.LogedIdPartition = login.Id
	stores.LogedUser = login.User
	err = setUpIDs(login.User, contentMatrix)
	if err != nil {
		return err
	}
	sb, part, diskPath, err := stores.GetMountedPartitionSuperblock(stores.LogedIdPartition)
	if err != nil {
		return err
	}
	if sb.IsExt3() {
		journalDirectory := &structures.Journal{
			J_next: -1,
			J_content: structures.Information{
				I_operation: [10]byte{'l', 'o', 'g', 'i', 'n'},
				I_path:      [74]byte{},
				I_content:   [64]byte{},
				I_date:      float32(time.Now().Unix()),
			},
		}
		fullContent := fmt.Sprintf("%s/%s/%s", login.Id, login.User, login.Password)
		copy(journalDirectory.J_content.I_content[:], fullContent)
		err = sb.AddJournal(journalDirectory, diskPath, int32(part.Part_start+int32(binary.Size(structures.SuperBlock{}))))
		if err != nil {
			return err
		}
	}
	return nil
}

func getContetnUsersTxt(idPartition string) (string, error) {
	var result string
	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "", err
	}
	parentDirs, destDir := utils.GetParentDirectories("/users.txt")
	content, err := partitionSuperblock.ContentFromFile(partitionPath, 0, parentDirs, destDir)
	if err != nil {
		return "", err
	}
	result += content
	return result, nil
}

func splitContent(content string, parameter string) []string {
	result := strings.Split(content, parameter)
	return result
}

func validateInformation(user, password string, matrix [][]string) bool {
	for _, row := range matrix {
		if row[1] != "U" {
			continue
		}
		if row[3] == user && row[4] == password {
			return true
		}
	}
	return false
}

func getContentMatrixUsers(contentUsers string) [][]string {
	contentSplitedByEnters := splitContent(contentUsers, "\n")
	contentSplitedByEnters = contentSplitedByEnters[:len(contentSplitedByEnters)-1]
	var contentMatrix [][]string
	for _, value := range contentSplitedByEnters {
		contentMatrix = append(contentMatrix, splitContent(value, ","))
	}
	return contentMatrix
}

func setUpIDs(userName string, matrix [][]string) error {
	var nameGroup string
	for _, row := range matrix {
		if row[1] != "U" {
			continue
		}
		if row[3] == userName {
			num, err := strconv.Atoi(row[0])
			if err != nil {
				return err
			}
			utils.LogedUserID = int32(num)

			nameGroup = row[2]
			break
		}
	}
	for _, row := range matrix {
		if row[1] != "G" {
			continue
		}
		if row[2] == nameGroup {
			num, err := strconv.Atoi(row[0])
			if err != nil {
				return err
			}
			utils.LogedUserGroupID = int32(num)
			break
		}
	}
	return nil
}
