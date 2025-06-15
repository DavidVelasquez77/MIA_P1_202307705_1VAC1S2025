package analyzer
// hola
import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	commands "server/commands"
	"server/stores"
	"server/structures"
	"server/utils"
	"strings"
	"time"
)

func Analyzer(input string) (interface{}, error) {
	tokens := strings.Fields(input)

	if len(tokens) == 0 {
		return "", nil
	}

	switch strings.ToLower(tokens[0]) {
	case "mkdir":
		return commands.ParseMkdir(tokens[1:])
	case "mkdisk":
		return commands.ParseMkdisk(tokens[1:])
	case "fdisk":
		return commands.ParseFdisk(tokens[1:])
	case "mount":
		return commands.ParseMount(tokens[1:])
	case "rmdisk":
		return commands.ParseRmdisk(tokens[1:])
	case "mounted":
		var result string
		if len(stores.MountedPartitions) == 0 {
			return "No hay particiones montadas", nil
		} else {
			for key := range stores.MountedPartitions {
				result += key + ", "
			}
		}
		fmt.Println("Particiones montadas: ", result)

		return fmt.Sprintf("Particiones montadas: %s", result), nil
	case "mkfs":
		return commands.ParseMkfs(tokens[1:])
	case "cat":
		return commands.ParseCat(tokens[1:])
	case "login":
		return commands.ParseLogin(tokens[1:])
	case "logout":
		if stores.LogedIdPartition == "" {
			return nil, errors.New("no hay sesion iniciada como para hacer un logout")
		}
		stores.LogedUser = ""
		temp := stores.LogedIdPartition
		stores.LogedIdPartition = ""
		utils.LogedUserGroupID = 1
		utils.LogedUserID = 1

		sb, part, diskPath, err := stores.GetMountedPartitionSuperblock(temp)
		if err != nil {
			return nil, err
		}
		if sb.IsExt3() {
			journalDirectory := &structures.Journal{
				J_next: -1,
				J_content: structures.Information{
					I_operation: [10]byte{'l', 'o', 'g', 'o', 'u', 't'},
					I_path:      [74]byte{},
					I_content:   [64]byte{},
					I_date:      float32(time.Now().Unix()),
				},
			}
			err = sb.AddJournal(journalDirectory, diskPath, int32(part.Part_start+int32(binary.Size(structures.SuperBlock{}))))
			if err != nil {
				return nil, err
			}
		}
		return "LOGOUT", nil
	case "mkgrp":
		return commands.ParseMkgrp(tokens[1:])
	case "rmgrp":
		return commands.ParseRmgrp(tokens[1:])
	case "mkusr":
		return commands.ParseMkusr(tokens[1:])
	case "rmusr":
		return commands.ParseRmusr(tokens[1:])
	case "mkfile":
		return commands.ParseMkfile(tokens[1:])
	case "rep":
		return commands.ParseRep(tokens[1:])
	case "unmount":
		return commands.ParseUnmount(tokens[1:])
	case "find":
		return commands.ParseFind(tokens[1:])
	case "execute":
		return ParseExecute(tokens[1:])
	case "pause":
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("Apache ENTER para continuar: ")
			if !scanner.Scan() {
				break
			}
			input := scanner.Text()
			if input == "" {
				break
			}
		}
		return "PAUSE", nil
	default:
		return nil, fmt.Errorf("comando desconocido: %v", tokens[0])
	}
}
