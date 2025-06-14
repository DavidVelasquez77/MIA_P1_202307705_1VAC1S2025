package commands

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	stores "server/stores"
	"server/structures"
	"strings"
	"time"
)

type RMGRP struct {
	name string
}

func ParseRmgrp(tokens []string) (string, error) {
	cmd := &RMGRP{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-name="[^"]+"|-name=[^\s]+`)
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
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede venir vacio")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)
		}
	}

	if cmd.name == "" {
		return "", errors.New("parametro obligatorio: -name")
	}
	err := CommandRmgrp(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("RMGRP: grupo %s eliminado exitosamente", cmd.name), nil

}

func CommandRmgrp(rmgrp *RMGRP) error {
	if stores.LogedIdPartition == "" {
		return errors.New("no hay sesion activa")
	}
	if stores.LogedUser != "root" {
		return errors.New("este comando solo lo puede ejecutar el usuario root")
	}
	contentUsersTxt, err := getContetnUsersTxt(stores.LogedIdPartition)
	if err != nil {
		return err
	}
	contentMatrix := getContentMatrixUsers(contentUsersTxt)
	outcome := removeGroup(rmgrp.name, contentMatrix)
	if !outcome {
		return errors.New("no existe el nombre del grupo a eliminar")
	}
	contentUsersTxt = reformUserstxt(contentMatrix)
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(stores.LogedIdPartition)
	if err != nil {
		return err
	}
	err = OverrideUserstxt(partitionSuperblock, partitionPath, contentUsersTxt)
	if err != nil {
		return err
	}

	if partitionSuperblock.IsExt3() {
		journalDirectory := &structures.Journal{
			J_next: -1,
			J_content: structures.Information{
				I_operation: [10]byte{'r', 'm', 'g', 'r', 'p'},
				I_path:      [74]byte{},
				I_content:   [64]byte{},
				I_date:      float32(time.Now().Unix()),
			},
		}
		copy(journalDirectory.J_content.I_content[:], rmgrp.name)
		err = partitionSuperblock.AddJournal(journalDirectory, partitionPath, int32(mountedPartition.Part_start+int32(binary.Size(structures.SuperBlock{}))))
		if err != nil {
			return err
		}
	}

	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return err
	}
	return nil
}

func reformUserstxt(matrix [][]string) string {
	var onlyRows []string
	for _, row := range matrix {
		onlyRows = append(onlyRows, strings.Join(row, ","))
	}
	fullContent := strings.Join(onlyRows, "\n")
	return fullContent
}

func removeGroup(nameGroup string, matrix [][]string) bool {
	for _, row := range matrix {
		if row[1] != "G" {
			continue
		}
		if row[2] == nameGroup {
			row[0] = "0"
			return true
		}
	}
	return false
}
