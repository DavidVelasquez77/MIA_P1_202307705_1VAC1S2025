package reports

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	stores "server/stores"
	"server/structures"
	utils "server/utils"
	"strconv"
	"strings"
	"time"
)

func ReportLs(path string, pathToGetInfo string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := `digraph G {
    node [shape=plaintext];
    tabla1 [label=<
        <table border="1" cellborder="1" cellspacing="0">
            <tr><td>Permisos</td><td>Owner</td><td>Grupo</td><td>Size</td><td>Fecha y Hora</td><td>Tipo</td><td>Name</td></tr>
    `

	// Ubicar el inodo desde donde todo se debe escribir
	superBlock, _, diskPath, err := stores.GetMountedPartitionSuperblock(stores.LogedIdPartition)
	if err != nil {
		return err
	}
	inodoBase, _, err := UbicarInodo(superBlock, pathToGetInfo, diskPath)
	if err != nil {
		return err
	}
	if inodoBase.I_type[0] == '1' {
		return errors.New("no se puede aplicar este reporte sobre un archivo")
	}

	// Contenido
	for i, blockIndex := range inodoBase.I_block {
		if blockIndex == -1 {
			continue
		}
		if i >= 14 {
			pointerBlock := &structures.PointerBlock{}
			err := pointerBlock.Deserialize(diskPath, int64(superBlock.S_block_start+(blockIndex*superBlock.S_block_size)))
			if err != nil {
				return err
			}
			for neoIndex := 0; neoIndex < len(pointerBlock.P_pointers); neoIndex++ {
				if pointerBlock.P_pointers[neoIndex] == -1 {
					continue
				}
				block := &structures.FolderBlock{}
				err := block.Deserialize(diskPath, int64(superBlock.S_block_start+(pointerBlock.P_pointers[neoIndex]*superBlock.S_block_size)))
				if err != nil {
					return err
				}
				for i := 2; i < len(block.B_content); i++ {
					content := block.B_content[i]
					if content.B_inodo == -1 {
						continue
					}
					temp, err := getLsString(superBlock, content.B_inodo, strings.Trim(string(content.B_name[:]), "\x00"), diskPath)
					if err != nil {
						return err
					}
					dotContent += temp
				}
			}
		} else {
			block := &structures.FolderBlock{}
			err := block.Deserialize(diskPath, int64(superBlock.S_block_start+(blockIndex*superBlock.S_block_size)))
			if err != nil {
				return err
			}
			for i := 2; i < len(block.B_content); i++ {
				content := block.B_content[i]
				if content.B_inodo == -1 {
					continue
				}
				temp, err := getLsString(superBlock, content.B_inodo, strings.Trim(string(content.B_name[:]), "\x00"), diskPath)
				if err != nil {
					return err
				}
				dotContent += temp
			}
		}
	}

	dotContent += `
	        </table>
    >];
}`
	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return err
	}
	defer dotFile.Close()

	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return err
	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil

}

func getPermissions(dato string) string {
	numero, _ := strconv.Atoi(dato)
	var result string
	switch numero {
	case 1:
		result = "--x"
	case 2:
		result = "-w-"
	case 3:
		result = "-wx"
	case 4:
		result = "r--"
	case 5:
		result = "r-x"
	case 6:
		result = "rw-"
	case 7:
		result = "rwx"
	}
	return result
}

func UbicarInodo(sb *structures.SuperBlock, dirPath string, diskPath string) (*structures.Inode, int32, error) {
	parentDirs, destDir := utils.GetParentDirectories(dirPath)
	if destDir != "" {
		parentDirs = append(parentDirs, destDir)
	}
	return getInode(sb, 0, diskPath, parentDirs)
}

func getInode(sb *structures.SuperBlock, inodeIndex int32, diskPath string, parentsDir []string) (*structures.Inode, int32, error) {
	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return nil, 0, err
	}
	if len(parentsDir) == 0 {
		return inode, inodeIndex, nil
	}
	for i, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}
		if i >= 14 {
			inderctNode := &structures.PointerBlock{}
			err := inderctNode.Deserialize(diskPath, int64(sb.S_block_start+(sb.S_block_size*blockIndex)))
			if err != nil {
				return nil, 0, err
			}
			for _, value := range inderctNode.P_pointers {
				if value == -1 {
					continue
				}
				block := &structures.FolderBlock{}
				err := block.Deserialize(diskPath, int64(sb.S_block_start+(value*sb.S_block_size)))
				if err != nil {
					return nil, 0, err
				}
				for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
					content := block.B_content[indexContent]
					if content.B_inodo == -1 {
						continue
					}
					parentDir, err := utils.First(parentsDir)
					if err != nil {
						return nil, 0, err
					}

					contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
					parentDirName := strings.Trim(parentDir, "\x00 ")
					if strings.EqualFold(contentName, parentDirName) {
						return getInode(sb, content.B_inodo, diskPath, utils.RemoveElement(parentsDir, 0))
					}
				}
			}

		} else {
			block := &structures.FolderBlock{}
			err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return nil, 0, err
			}
			for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
				content := block.B_content[indexContent]
				if content.B_inodo == -1 {
					continue
				}
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return nil, 0, err
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")
				if strings.EqualFold(contentName, parentDirName) {
					return getInode(sb, content.B_inodo, diskPath, utils.RemoveElement(parentsDir, 0))
				}
			}
		}

	}
	return nil, 0, errors.New("no existe la ruta especificada")
}

func getLsString(sb *structures.SuperBlock, inodeIndex int32, nombre string, diskPath string) (string, error) {

	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return "", err
	}
	var permissions string
	tempPermisions := string(inode.I_perm[:])
	for i := 0; i < 3; i++ {
		permissions += getPermissions(string(tempPermisions[i])) + " "
	}
	owner, err := getOwnerByID(inode.I_uid)
	if err != nil {
		return "", err
	}
	group, err := getGroupById(inode.I_gid)
	if err != nil {
		return "", err
	}
	var tipoInodo string
	if inode.I_type[0] == '0' {
		tipoInodo = "Carpeta"
	} else {
		tipoInodo = "Archivo"
	}
	mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)
	dotContent := fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%s</td><td>%s</td><td>%s</td></tr>`, permissions, owner, group, inode.I_size, mtime, tipoInodo, nombre)
	return dotContent, nil
}

func getOwnerByID(id int32) (string, error) {
	contentUsersTxt, err := GetContetnUsersTxt(stores.LogedIdPartition)
	if err != nil {
		return "", err
	}
	strId := strconv.Itoa(int(id))
	contentMatrix := GetContentMatrixUsers(contentUsersTxt)
	for _, row := range contentMatrix {
		if row[0] != strId {
			continue
		}
		if row[1] != "U" {
			continue
		}
		return row[3], nil
	}
	return "", errors.New("no se encontro el usuario")
}

func getGroupById(id int32) (string, error) {
	contentUsersTxt, err := GetContetnUsersTxt(stores.LogedIdPartition)
	if err != nil {
		return "", err
	}
	strId := strconv.Itoa(int(id))
	contentMatrix := GetContentMatrixUsers(contentUsersTxt)
	for _, row := range contentMatrix {
		if row[0] != strId {
			continue
		}
		if row[1] != "G" {
			continue
		}
		return row[2], nil
	}
	return "", errors.New("no se encontro el usuario")
}

func GetContetnUsersTxt(idPartition string) (string, error) {
	var result string
	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "", err
	}
	parentDirs, destDir := utils.GetParentDirectories("/users.txt")
	content, err := partitionSuperblock.ContentFromFile(partitionPath, 0, parentDirs, destDir)
	if err != nil {
		return "", err
	}
	result += content
	return result, nil
}

func GetContentMatrixUsers(contentUsers string) [][]string {
	contentSplitedByEnters := splitContent(contentUsers, "\n")
	contentSplitedByEnters = contentSplitedByEnters[:len(contentSplitedByEnters)-1]
	var contentMatrix [][]string
	for _, value := range contentSplitedByEnters {
		contentMatrix = append(contentMatrix, splitContent(value, ","))
	}
	return contentMatrix
}

func splitContent(content string, parameter string) []string {
	result := strings.Split(content, parameter)
	return result
}
