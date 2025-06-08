package analyzer

import (
	"errors"
	"fmt"
	commands "server/commands"
	"server/stores"
	"server/utils"
	"strings"
)

func Analyzer(input string) (interface{}, error) {
	tokens := strings.Fields(input)

	if len(tokens) == 0 {
		return nil, errors.New("no hay ningun comando")
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
		return fmt.Sprintf("particiones montadas: %s", result), nil
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
		stores.LogedIdPartition = ""
		utils.LogedUserGroupID = 1
		utils.LogedUserID = 1
		return "LOGOUT", nil
	case "mkgrp":
		return commands.ParseMkgrp(tokens[1:])
	case "rmgrp":
		return commands.ParseRmgrp(tokens[1:])
	case "mkusr":
		return commands.ParseMkusr(tokens[1:])
	case "rmusr":
		return commands.ParseRmusr(tokens[1:])
	case "chgrp":
		return commands.ParseChgrp(tokens[1:])
	case "mkfile":
		return commands.ParseMkfile(tokens[1:])
	case "rep":
		return commands.ParseRep(tokens[1:])
	default:
		return nil, fmt.Errorf("comando desconocido: %v", tokens[0])
	}
}
