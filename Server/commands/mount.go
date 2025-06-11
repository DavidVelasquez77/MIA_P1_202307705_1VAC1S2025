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

type MOUNT struct {
	path string
	name string
}

func ParseMount(tokens []string) (string, error) {
	cmd := &MOUNT{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
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
			if value == "" {
				return "", errors.New("el pathno puede estar vacio")
			}
			cmd.path = value
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede estar vacio")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}
	err := commandMount(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MOUNT: %s montada exitosamente", cmd.name), nil
}
func commandMount(mount *MOUNT) error {
	var mbr structures.MBR

	err := mbr.DeserializeMBR(mount.path)
	if err != nil {
		return err
	}
	partition, indexPartition := mbr.GetPartitionByName(mount.name)
	if partition == nil {
		return errors.New("la particion no existe")
	}

	// fmt.Println("\nPartición disponible:")
	// partition.PrintPartition()

	if partition.Part_status[0] == '1' {
		return errors.New("no se puede montar una particion ya montada")
	}

	if partition.Part_type[0] == 'E' {
		return errors.New("no se puede montar una particion extendida")
	}

	idPartition, err := generatePartitionID(mount)
	if err != nil {
		return err
	}

	stores.MountedPartitions[idPartition] = mount.path
	partition.MountPartition(indexPartition, idPartition)

	// fmt.Println("\nPartición creada (modificada):")
	// partition.PrintPartition()

	mbr.Mbr_partitions[indexPartition] = *partition

	err = mbr.SerializeMBR(mount.path)
	if err != nil {
		return err
	}
	return nil
}

func generatePartitionID(mount *MOUNT) (string, error) {
	letter, partitionCorrelative, err := utils.GetLetter(mount.path)
	if err != nil {
		return "", err
	}

	idPartition := fmt.Sprintf("%s%d%s", stores.Carnet, partitionCorrelative, letter)

	return idPartition, nil

}
