package reports

import (
	"fmt"
	"os"
	"os/exec"
	structures "server/structures"
	utils "server/utils"
	"strings"
)

var contadorRB int32 = 0

func ReportBlock(superBlock *structures.SuperBlock, diskPath, path string) error {

	contadorRB = 0
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}
	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := `strict digraph G {
        node [shape=plaintext]
		rankdir=LR;
		`
	for i := int32(0); i < superBlock.S_inodes_count; i++ {
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(superBlock.S_inode_start+(superBlock.S_inode_size*i)))
		if err != nil {
			return err
		}
		var content string
		if i < superBlock.S_inodes_count-1 {
			content, err = getStringBlock(inode, diskPath, superBlock, false)
		} else {
			content, err = getStringBlock(inode, diskPath, superBlock, true)
		}
		if err != nil {
			return err
		}
		dotContent += content
	}
	dotContent += "}"
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

func getStringBlock(inode *structures.Inode, diskPath string, sb *structures.SuperBlock, isTheLast bool) (string, error) {
	dotContent := ""
	for i, blockIndex := range inode.I_block {

		if blockIndex == -1 {
			continue
		}
		fValue := getNumber()
		// Aqui diferenciar si es de tipo 0 o 1 el Inodo
		if i >= 14 {
			pointerBlock := structures.PointerBlock{}
			err := pointerBlock.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return "", err
			}
			dotContent += fmt.Sprintf(`node%d[shape=record label="Bloque Apuntador%d\n`, fValue, blockIndex)
			for index, value := range pointerBlock.P_pointers {
				if index%6 == 0 {
					dotContent += "\n"
				}
				dotContent += fmt.Sprintf(" %d,", value)
			}
			dotContent += `"];
			`
			for iTe, neoIndexBlock := range pointerBlock.P_pointers {
				if neoIndexBlock == -1 {
					continue
				}
				neoFValue := getNumber()
				if inode.I_type[0] == '0' {
					block := &structures.FolderBlock{}
					err := block.Deserialize(diskPath, int64(sb.S_block_start+(neoIndexBlock*sb.S_block_size)))
					if err != nil {
						return "", err
					}
					dotContent += fmt.Sprintf(`node%d[shape=record label="Bloque Carpeta%d\nb_name : b_inodo\n`, neoFValue, neoIndexBlock)
					for _, value := range block.B_content {
						tempString := string(value.B_name[:])
						nameTemp := strings.TrimRight(tempString, "\x00")
						dotContent += fmt.Sprintf(" %s : %d\\n", nameTemp, int32(value.B_inodo))
					}
					dotContent += `"];
					`
					for neoInodeIndex := 2; i < len(block.B_content); i++ {

						content := block.B_content[neoInodeIndex]
						if content.B_inodo == -1 {
							continue
						}
						inode := &structures.Inode{}
						err = inode.Deserialize(diskPath, int64(sb.S_inode_start+(sb.S_inode_size*content.B_inodo)))
						if err != nil {
							return "", err
						}
						temp, err := getStringBlock(inode, diskPath, sb, false)
						if err != nil {
							return "", err
						}
						dotContent += temp
					}
					if !isTheLast {
						if i < len(inode.I_block)-1 {
							dotContent += fmt.Sprintf("node%d -> node%d;\n", neoFValue, neoFValue+1)
						}
					} else {
						if i < len(inode.I_block)-1 && inode.I_block[i+1] != -1 {
							dotContent += fmt.Sprintf("node%d -> node%d;\n", neoFValue, neoFValue+1)
						}
					}
				} else if inode.I_type[0] == '1' {
					if iTe == 0 {
						dotContent += fmt.Sprintf("node%d -> node%d;\n", fValue, fValue+1)
					}
					block := &structures.FileBlock{}
					err := block.Deserialize(diskPath, int64(sb.S_block_start+(neoIndexBlock*sb.S_block_size)))
					if err != nil {
						return "", err
					}
					contentBlock := string(block.B_content[:])
					contentBlock = strings.TrimRight(contentBlock, "\x00")
					splitContent := splitEqualParts(contentBlock)
					dotContent += fmt.Sprintf(`node%d[shape=record label="Bloque archivo %d\n%s\n%s\n%s\n%s"];
					`, neoFValue, neoIndexBlock, splitContent[0], splitContent[1], splitContent[2], splitContent[3])
					dotContent += fmt.Sprintf("node%d -> node%d;\n", neoFValue, neoFValue+1)
				}
			}
		} else if inode.I_type[0] == '0' {
			block := &structures.FolderBlock{}
			err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return "", err
			}
			dotContent += fmt.Sprintf(`node%d[shape=record label="Bloque Carpeta%d\nb_name : b_inodo\n`, fValue, blockIndex)
			for _, value := range block.B_content {
				tempString := string(value.B_name[:])
				nameTemp := strings.TrimRight(tempString, "\x00")
				dotContent += fmt.Sprintf(" %s : %d\\n", nameTemp, int32(value.B_inodo))
			}
			dotContent += `"];
				`
		} else if inode.I_type[0] == '1' {

			block := &structures.FileBlock{}
			err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {

				return "", err
			}

			contentBlock := string(block.B_content[:])
			contentBlock = strings.TrimRight(contentBlock, "\x00")
			splitContent := splitEqualParts(contentBlock)
			dotContent += fmt.Sprintf(`node%d[shape=record label="Bloque archivo %d\n%s\n%s\n%s\n%s"];
					`, fValue, blockIndex, splitContent[0], splitContent[1], splitContent[2], splitContent[3])

		}
		if !isTheLast {
			if i < len(inode.I_block)-1 {
				dotContent += fmt.Sprintf("node%d -> node%d;\n", fValue, fValue+1)
			}
		} else {
			if i < len(inode.I_block)-1 && inode.I_block[i+1] != -1 {
				dotContent += fmt.Sprintf("node%d -> node%d;\n", fValue, fValue+1)
			}
		}

	}
	return dotContent, nil
}

func splitEqualParts(s string) []string {
	if len(s) > 16 {
		n := len(s) / 4
		return []string{s[:n], s[n : 2*n], s[2*n : 3*n], s[3*n:]}
	}
	return []string{s, "", "", ""}
}

func getNumber() int32 {
	contadorRB++
	return contadorRB
}
