package commands

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	stores "server/stores"
	structures "server/structures"
	utils "server/utils"
	"strings"
	"time"
)

type MKDIR struct {
	path string
	p    bool
}

func ParseMkdir(tokens []string) (string, error) {
	cmd := &MKDIR{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-r`)
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
		case "-r":
			cmd.p = true
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	err := CommandMkdir(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKDIR: Directorio %s creado correctamente.", cmd.path), nil
}

// Aquí debería de estar logeado un usuario, por lo cual el usuario debería tener consigo el id de la partición
// En este caso el ID va a estar quemado
// var idPartition = "361A"

func CommandMkdir(mkdir *MKDIR) error {
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(stores.LogedIdPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	err = createDirectory(mkdir.path, partitionSuperblock, partitionPath, mountedPartition, mkdir.p)
	if err != nil {
		err = fmt.Errorf("error al crear el directorio: %w", err)
	}

	return err
}

func createDirectory(dirPath string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.PARTITION, flag bool) error {

	parentDirs, destDir := utils.GetParentDirectories(dirPath)

	err := sb.CreateFolder(partitionPath, parentDirs, destDir, flag)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}
	if sb.IsExt3() {
		journalDirectory := &structures.Journal{
			J_next: -1,
			J_content: structures.Information{
				I_operation: [10]byte{'m', 'k', 'd', 'i', 'r'},
				I_path:      [74]byte{},
				I_content:   [64]byte{},
				I_date:      float32(time.Now().Unix()),
			},
		}
		copy(journalDirectory.J_content.I_path[:], dirPath)
		err = sb.AddJournal(journalDirectory, partitionPath, int32(mountedPartition.Part_start+int32(binary.Size(structures.SuperBlock{}))))
		if err != nil {
			return err
		}
	}

	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
