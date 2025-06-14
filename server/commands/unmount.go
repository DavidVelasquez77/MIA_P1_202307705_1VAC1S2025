package commands

import (
	"errors"
	"fmt"
	"regexp"
	"server/stores"
	"server/structures"
	"server/utils"
	"strings"
)

type UNMOUNT struct {
	id string
}

// Validar si esta montada
// Cambiar el valor del estado a 0

func ParseUnmount(tokens []string) (string, error) {
	cmd := &UNMOUNT{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[a-zA-Z0-9]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parametro invalido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		switch key {
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacio")
			}
			cmd.id = value
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)

		}
	}
	if cmd.id == "" {
		return "", errors.New("faltan parametros requeridos: -id")
	}
	err := CommandUnmount(cmd)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("UNMOUNT: %s desmontado exitosamente", cmd.id), nil

}

func CommandUnmount(unmount *UNMOUNT) error {
	diskPath := stores.MountedPartitions[unmount.id]
	if diskPath == "" {
		return errors.New("id de particion no montada")
	}
	mbr := &structures.MBR{}
	err := mbr.DeserializeMBR(diskPath)
	if err != nil {
		return err
	}
	partition, index, err := mbr.GetPartitionByID(unmount.id)
	if err != nil {
		return err
	}
	if partition.Part_status[0] == '0' {
		return errors.New("no se puede desmontar una particion no montada")
	}
	partition.Part_status[0] = '0'
	mbr.Mbr_partitions[index] = *partition
	err = mbr.SerializeMBR(diskPath)
	if err != nil {
		return err
	}
	utils.PathToPartitionCount[diskPath] -= 1
	delete(stores.MountedPartitions, unmount.id)
	return nil
}
