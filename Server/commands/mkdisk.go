package commands

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	stores "server/stores"
	structures "server/structures"
	utils "server/utils"
	"strconv"
	"time"

	"regexp"
	"strings"
)

type MKDISK struct {
	size int
	unit string
	fit  string
	path string
}

func ParseMkdisk(tokens []string) (string, error) {
	cmd := &MKDISK{}
	letterDisk := utils.GetLetterToDisk()
	cmd.path = stores.GetPathDisk(letterDisk)
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmM]|-fit=[bBfFwW]{2}`)
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
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil {
				return "", errors.New("el tamano debe ser numero entero positivo")
			}
			cmd.size = size
		case "-unit":
			value = strings.ToUpper(value)
			if value != "K" && value != "M" {
				return "", errors.New("la unidd debe ser K o M")
			}
			cmd.unit = value
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)
		}
	}

	if cmd.size == 0 {
		return "", errors.New("faltan parametros requeridos: -size")
	}
	if cmd.unit == "" {
		cmd.unit = "K"
	}
	if cmd.fit == "" {
		cmd.fit = "FF"
	}
	err := commandMkdisk(cmd)
	if err != nil {
		return "", err
	}
	name := utils.GetNameByPath(cmd.path)
	stores.LoadedDiskPaths[name] = cmd.path
	return fmt.Sprintf("MKDISK: %s creado exitosamente", cmd.path), nil

}

func commandMkdisk(mkdisk *MKDISK) error {
	sizeBytes, err := utils.ConvertToBytes(mkdisk.size, mkdisk.unit)
	if err != nil {

		return fmt.Errorf("error creando el disco: %v", err)
	}

	err = createDisk(mkdisk, sizeBytes)
	if err != nil {

		return fmt.Errorf("error creating disk: %v", err)
	}

	err = createMBR(mkdisk, sizeBytes)
	if err != nil {
		return err
	}
	return nil

}

func createDisk(mkdisk *MKDISK, sizeBytes int) error {
	err := os.MkdirAll(filepath.Dir(mkdisk.path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creando los directorios: %v", err)
	}
	file, err := os.Create(mkdisk.path)
	if err != nil {
		return fmt.Errorf("error creando el archivo: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, 1024*1024)
	for sizeBytes > 0 {
		writeSize := len(buffer)
		if sizeBytes < writeSize {
			writeSize = sizeBytes
		}
		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return err
		}
		sizeBytes -= writeSize
	}
	return nil
}

func createMBR(mkdisk *MKDISK, sizeBytes int) error {
	var fitByte byte
	switch mkdisk.fit {
	case "FF":
		fitByte = 'F'
	case "BF":
		fitByte = 'B'
	case "WF":
		fitByte = 'W'
	default:
		return errors.New("invalido fit type")
	}

	mbr := &structures.MBR{
		Mbr_size:           int32(sizeBytes),
		Mbr_creation_date:  float32(time.Now().Unix()),
		Mbr_disk_signature: rand.Int31(),
		Mbr_disk_fit:       [1]byte{fitByte},
		Mbr_partitions: [4]structures.PARTITION{
			{Part_status: [1]byte{'N'}, Part_type: [1]byte{'N'}, Part_fit: [1]byte{'N'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
			{Part_status: [1]byte{'N'}, Part_type: [1]byte{'N'}, Part_fit: [1]byte{'N'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
			{Part_status: [1]byte{'N'}, Part_type: [1]byte{'N'}, Part_fit: [1]byte{'N'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
			{Part_status: [1]byte{'N'}, Part_type: [1]byte{'N'}, Part_fit: [1]byte{'N'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
		},
	}

	err := mbr.SerializeMBR(mkdisk.path)
	if err != nil {
		return err
	}
	return nil

}
