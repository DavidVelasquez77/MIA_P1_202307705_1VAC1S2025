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

type MKUSR struct {
	user     string
	password string
	group    string
}

func ParseMkusr(tokens []string) (string, error) {
	cmd := &MKUSR{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-user="[^"]+"|-user=[^\s]+|-grp="[^"]+"|-grp=[^\s]+|-pass="[^"]+"|-pass=[^\s]+`)
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
		case "-pass":
			if value == "" {
				return "", errors.New("el password no puede estar vacio")
			}
			if len(value) > 10 {
				return "", errors.New("el pass de usuario no se puede exceder de 10 caracteres")
			}
			cmd.password = value
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

	if cmd.password == "" {
		return "", errors.New("faltan parametros requeridos: -pass")
	}
	if cmd.user == "" {
		return "", errors.New("faltan parametros requeridos: -user")
	}
	if cmd.group == "" {
		return "", errors.New("faltan parametros requeridos: -grp")
	}

	err := commandMkusr(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKUSR: usuario %s creado exitosamente", cmd.user), nil
}

func commandMkusr(mkusr *MKUSR) error {
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
	outcome := !soleNameGroup(mkusr.group, contentMatrix)
	if !outcome {
		return errors.New("el grupo especificado no existe")
	}
	outcome = soleNameUser(mkusr.user, contentMatrix)
	if !outcome {
		return errors.New("nombre de usuario no disponible")
	}
	neoUserID := getNeoNumber("U", contentMatrix)
	contentUsersTxt += fmt.Sprintf("%d,U,%s,%s,%s\n", neoUserID, mkusr.group, mkusr.user, mkusr.password)
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
				I_operation: [10]byte{'m', 'k', 'u', 's', 'r'},
				I_path:      [74]byte{},
				I_content:   [64]byte{},
				I_date:      float32(time.Now().Unix()),
			},
		}
		fullContent := fmt.Sprintf("%s/%s/%s", mkusr.user, mkusr.password, mkusr.group)
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

func soleNameUser(userName string, matrix [][]string) bool {
	for _, row := range matrix {
		if row[1] != "U" {
			continue
		}
		if row[3] == userName {
			return false
		}
	}
	return true
}
