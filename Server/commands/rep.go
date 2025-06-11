package commands

import (
	"errors"
	"fmt"
	"regexp"
	"server/reports"
	"server/stores"
	"strings"
)

type REP struct {
	name         string
	path         string
	id           string
	path_file_ls string
}

func ParseRep(tokens []string) (string, error) {
	cmd := &REP{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[a-zA-Z0-9]+|-path_file_ls="[^"]+"|-path_file_ls=[^\s]+|-path="[^"]+"|-path=[^\s]+|-name=[a-zA-Z_]+`)
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
		case "-path_file_ls":
			if value == "" {
				return "", errors.New("el path_file_is no puede estar vacio")
			}
			cmd.path_file_ls = value
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacio")
			}
			cmd.path = value
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacio")
			}
			cmd.id = value
		case "-name":
			if value == "" {
				return "", errors.New("el name no puede estar vacio")
			}
			switch value {
			case "mbr":
				cmd.name = "mbr"
			case "disk":
				cmd.name = "disk"
			case "inode":
				cmd.name = "inode"
			case "block":
				cmd.name = "block"
			case "bm_inode":
				cmd.name = "bm_inode"
			case "bm_block":
				cmd.name = "bm_block"
			case "tree":
				cmd.name = "tree"
			case "sb":
				cmd.name = "sb"
			case "file":
				cmd.name = "file"
			case "ls":
				cmd.name = "ls"
			default:
				return "", fmt.Errorf("valor del nombre invalido: %s", value)
			}
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)
		}
	}
	if cmd.path == "" {
		return "", errors.New("faltan parametros requeridos: -path")
	}
	if cmd.id == "" {
		return "", errors.New("faltan parametros requeridos: -id")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parametros requeridos: -name")
	}

	err := commandRep(cmd)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("REP: el reporte %s fue generado con exito", cmd.name), nil

}

func commandRep(rep *REP) error {
	mountedMbr, mountedSb, mountedDiskPath, err := stores.GetMountedPartitionRep(rep.id)
	if err != nil {
		return err
	}

	switch rep.name {
	case "mbr":
		err = reports.ReportMBR(mountedMbr, rep.path, rep.id)
		if err != nil {
			return err
		}
	case "disk":
		err = reports.ReportDisk(mountedMbr, rep.id, rep.path, mountedDiskPath)
		if err != nil {
			return err
		}
	case "inode":
		err = reports.ReportInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			return err
		}
	case "block":
		err = reports.ReportBlock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			return err
		}
	case "bm_inode":
		err = reports.ReportBMInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			return err
		}
	case "bm_block":
		err = reports.ReportBMBlock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			return err
		}
	case "sb":
		err = reports.ReportSuperBlock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			return err
		}
	case "tree":
		err = reports.ReportTree(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			return err
		}
	case "file":
		err = reports.ReportFile(mountedSb, mountedDiskPath, rep.path, rep.path_file_ls)
		if err != nil {
			return err
		}
	case "ls":
		err = reports.ReportLs(rep.path, rep.path_file_ls)
		if err != nil {
			return err
		}
	}

	return nil
}
