package commands

import (
	"errors"
	"fmt"
	"regexp"
	"server/reports"
	"server/stores"
	utils "server/utils"
	"strings"
)

type FIND struct {
	path string
	name string
}

// \.    .*             .{1}
func ParseFind(tokens []string) (string, error) {
	cmd := &FIND{}
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
				return "", errors.New("el path no puede estar vacio")
			}
			cmd.path = value
		case "-name":
			if value == "" {
				return "", errors.New("el name no puede estar vacio")
			}
			value = strings.ReplaceAll(value, ".", "\\.")
			value = strings.ReplaceAll(value, "*", ".+")
			value = strings.ReplaceAll(value, "?", ".{1}")
			cmd.name = "^" + value + "$"
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

	result, err := commandFind(cmd)
	if err != nil {
		return "", err
	}
	fmt.Println(result)

	return fmt.Sprintf("FIND: %s\n%s ", cmd.path, result), nil
}

func commandFind(find *FIND) (string, error) {
	sb, _, diskPath, err := stores.GetMountedPartitionSuperblock(stores.LogedIdPartition)
	if err != nil {
		return "", err
	}

	inodoBase, offsetToSerialize, err := reports.UbicarInodo(sb, find.path, diskPath)
	if err != nil {
		return "", err
	}
	outcome, err := inodoBase.HasPermissionsToRead(utils.LogedUserID, utils.LogedUserGroupID)
	if err != nil {
		return "", err
	}
	if !outcome {
		return "", errors.New("accion prohibida por falta de permisos")
	}
	tipoInodo, err := sb.TypeOfInode(diskPath, offsetToSerialize)
	if err != nil {
		return "", err
	}
	if tipoInodo == 1 {
		return "", errors.New("este comando solo es aplicable a carpetas no a archivos")
	}

	return sb.CommandFind(diskPath, offsetToSerialize, 1, find.name)
}
