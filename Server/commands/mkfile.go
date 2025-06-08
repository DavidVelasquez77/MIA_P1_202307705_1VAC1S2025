package commands

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"regexp"
	"server/stores"
	"server/structures"
	"server/utils"
	"strconv"
	"strings"
	"time"
)

type MKFILE struct {
	path string
	r    bool //true si viene el parametro
	size int
	cont string
}

func ParseMkfile(tokens []string) (string, error) {
	cmd := &MKFILE{}
	cmd.size = 0
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-size=-?\d+|-r|-path="[^"]+"|-path=[^\s]+|-cont="[^"]+"|-cont=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {

		if match == "-r" {
			cmd.r = true
			continue
		}

		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parametro invalido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil {
				return "", errors.New("el tamano debe ser numero entero positivo")
			}
			if size <= -1 {
				return "", errors.New("no puede ser un numero negativo")
			}
			cmd.size = size
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacio")
			}
			cmd.path = value
		case "-cont":
			if value == "" {
				return "", errors.New("el cont no puede estar vacio")
			}
			cmd.cont = value
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parametros requeridos: -path")
	}

	err := commandMkfile(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKFILE: %s creado exitosamente", cmd.path), nil
}

func commandMkfile(mkfile *MKFILE) error {

	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(stores.LogedIdPartition)
	if err != nil {
		return err
	}
	err = createFile(partitionPath, partitionSuperblock, mountedPartition, mkfile.path, mkfile.r, mkfile.size, mkfile.cont)
	if err != nil {
		return err
	}
	return nil
}

func createFile(diskPath string, sb *structures.SuperBlock, partition *structures.PARTITION, filePath string, createDir bool, sizeFile int, pathFileToGetInfo string) error {
	var contentToWrite string
	if sizeFile < 0 {
		return fmt.Errorf("no puede venir un size negativo")
	}
	if createDir {
		position := strings.LastIndex(filePath, "/")
		dirPath := filePath[:position]
		parentDirs, destDir := utils.GetParentDirectories(dirPath)
		err := sb.CreateFolder(diskPath, parentDirs, destDir, true)
		if err != nil {
			return err
		}
	}
	if pathFileToGetInfo != "" {
		fileContent, err := os.ReadFile(pathFileToGetInfo)
		if err != nil {
			return err
		}
		contentToWrite = string(fileContent)
		parentDirs, destDir := utils.GetParentDirectories(filePath)
		err = sb.CreateFile(diskPath, 0, parentDirs, destDir, string(fileContent), int32(len(fileContent)), false)
		if err != nil {
			return err
		}
	} else if sizeFile > 0 {
		content := getStringContent(sizeFile)
		contentToWrite = content
		parentDirs, destDir := utils.GetParentDirectories(filePath)
		err := sb.CreateFile(diskPath, 0, parentDirs, destDir, content, int32(sizeFile), false)
		if err != nil {
			return err
		}
	} else {
		contentToWrite = ""
		parentDirs, destDir := utils.GetParentDirectories(filePath)
		err := sb.CreateFile(diskPath, 0, parentDirs, destDir, "", int32(0), false)
		if err != nil {
			return err
		}
	}

	if sb.IsExt3() {
		if contentToWrite != "" {
			contentList := utils.SplitStringIntoChunks(contentToWrite)
			for _, content := range contentList {
				journalDirectory := &structures.Journal{
					J_next: -1,
					J_content: structures.Information{
						I_operation: [10]byte{'m', 'k', 'f', 'i', 'l', 'e'},
						I_path:      [74]byte{},
						I_content:   [64]byte{},
						I_date:      float32(time.Now().Unix()),
					},
				}
				copy(journalDirectory.J_content.I_path[:], filePath)
				copy(journalDirectory.J_content.I_content[:], content)
				err := sb.AddJournal(journalDirectory, diskPath, int32(partition.Part_start+int32(binary.Size(structures.SuperBlock{}))))
				if err != nil {
					return err
				}
			}
		} else {
			journalDirectory := &structures.Journal{
				J_next: -1,
				J_content: structures.Information{
					I_operation: [10]byte{'m', 'k', 'f', 'i', 'l', 'e'},
					I_path:      [74]byte{},
					I_content:   [64]byte{},
					I_date:      float32(time.Now().Unix()),
				},
			}
			copy(journalDirectory.J_content.I_path[:], filePath)
			copy(journalDirectory.J_content.I_content[:], contentToWrite)
			err := sb.AddJournal(journalDirectory, diskPath, int32(partition.Part_start+int32(binary.Size(structures.SuperBlock{}))))
			if err != nil {
				return err
			}
		}
	}

	err := sb.Serialize(diskPath, int64(partition.Part_start))
	if err != nil {
		return err
	}
	return nil
}

func getStringContent(size int) string {
	var buffer string = ""
	numeros := "0123456789"
	aux := 0
	for size != 0 {
		if aux > 9 {
			aux = 0
		}
		buffer += string(numeros[aux])
		aux++
		size--
	}
	return buffer
}
