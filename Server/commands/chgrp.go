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

type CHGRP struct {
	user  string
	group string
}

func ParseChgrp(tokens []string) (string, error) {
	cmd := &CHGRP{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-user="[^"]+"|-user=[^\s]+|-grp="[^"]+"|-grp=[^\s]+`)
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
		case "-user":
			if value == "" {
				return "", errors.New("el user no puede estar vacio")
			}
			if len(value) > 10 {
				return "", errors.New("el user de usuario no se puede exceder de 10 caracteres")
			}
			cmd.user = value
		case "-grp":
			if value == "" {
				return "", errors.New("el grp no puede estar vacio")
			}
			if len(value) > 10 {
				return "", errors.New("el group de usuario no se puede exceder de 10 caracteres")
			}
			cmd.group = value
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)
		}
	}

	if cmd.user == "" {
		return "", errors.New("faltan parametros requeridos: -user")
	}
	if cmd.group == "" {
		return "", errors.New("faltan parametros requeridos: -grp")
	}

	err := commandChgrp(cmd)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("CHGRP: se ha cambiado al usuario %s al grupo %s exitosamente", cmd.user, cmd.group), nil
}

func commandChgrp(chgrp *CHGRP) error {
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

	outcome := validUser(chgrp.user, contentMatrix)
	if !outcome {
		return errors.New("el usuario no existe o esta elimnado")
	}
	outcome = validGroup(chgrp.group, contentMatrix)
	if !outcome {
		return errors.New("el grupo no existe o esta eliminado")
	}
	outcome = changeGroup(chgrp.user, chgrp.group, contentMatrix)
	if !outcome {
		return errors.New("se produjo un error al cambiar de grupo al usuario")
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
				I_operation: [10]byte{'c', 'h', 'g', 'r', 'p'},
				I_path:      [74]byte{},
				I_content:   [64]byte{},
				I_date:      float32(time.Now().Unix()),
			},
		}
		fullContent := fmt.Sprintf("%s/%s", chgrp.user, chgrp.group)
		copy(journalDirectory.J_content.I_content[:], fullContent)
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

func validUser(userName string, matrix [][]string) bool {
	for _, row := range matrix {
		if row[1] != "U" {
			continue
		}
		if row[0] == "0" {
			continue
		}
		if row[3] == userName {
			return true
		}
	}
	return false
}

func validGroup(groupName string, matrix [][]string) bool {
	for _, row := range matrix {
		if row[1] != "G" {
			continue
		}
		if row[0] == "0" {
			continue
		}
		if row[2] == groupName {
			return true
		}
	}
	return false
}

func changeGroup(userName, groupName string, matrix [][]string) bool {
	for _, row := range matrix {
		if row[1] != "U" {
			continue
		}
		if row[3] == userName {
			row[2] = groupName
			return true
		}
	}
	return false
}

