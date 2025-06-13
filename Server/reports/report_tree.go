package reports

import (
	"fmt"
	"os"
	"os/exec"
	structures "server/structures"
	utils "server/utils"
	"strings"
	"time"
)

var node int = 0

func ReportTree(sb *structures.SuperBlock, diskPath, path string) error {
	contador = 0

	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := `digraph G {
        node [shape=plaintext]
		rankdir=LR;
	`
	// Contenido returbio
	inode := &structures.Inode{}
	err = inode.Deserialize(diskPath, int64(sb.S_inode_start+(sb.S_inode_size*0)))
	if err != nil {
		return err
	}
	temp, err := getInodeDOT(sb, inode, diskPath, 0, true, 0)
	if err != nil {
		return err
	}
	dotContent += temp
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

func getNode() int {
	node++
	return node
}

func getInodeDOT(sb *structures.SuperBlock, inode *structures.Inode, diskPath string, nodoPadre int, isTheRoot bool, numberInode int) (string, error) {
	nodoActual := getNode()
	// var tempNumbers []int
	atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
	ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
	mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)

	dotContent := fmt.Sprintf(`node%d [fillcolor="#85c1e9 " style=filled label=<
	<table border="0" cellborder="1" cellspacing="0">
		<tr><td colspan="2"> REPORTE INODO %d </td></tr>
		<tr><td >i_uid</td><td>%d</td></tr>
		<tr><td >i_gid</td><td>%d</td></tr>
		<tr><td >i_size</td><td>%d</td></tr>
		<tr><td >i_atime</td><td>%s</td></tr>
		<tr><td >i_ctime</td><td>%s</td></tr>
		<tr><td >i_mtime</td><td>%s</td></tr>
		<tr><td >i_type</td><td>%c</td></tr>
		<tr><td >i_perm</td><td>%s</td></tr>
		<tr><td  colspan="2">BLOQUES DIRECTOS</td></tr>
	`, nodoActual, numberInode, inode.I_uid, inode.I_gid, inode.I_size, atime, ctime, mtime, rune(inode.I_type[0]), string(inode.I_perm[:]))
	for j, block := range inode.I_block {
		if j > 14 {
			break
		}
		dotContent += fmt.Sprintf("<tr><td >%d</td><td>%d</td></tr>", j+1, block)
	}

	dotContent += fmt.Sprintf(`
			<tr><td  colspan="2">BLOQUE INDIRECTO</td></tr>
			<tr><td >%d</td><td>%d</td></tr>
			<tr><td  colspan="2">BLOQUE INDIRECTO DOBLE</td></tr>
			<tr><td >%d</td><td>%d</td></tr>
			<tr><td  colspan="2">BLOQUE INDIRECTO TRIPLE</td></tr>
			<tr><td >%d</td><td>%d</td></tr>
		</table>>];
	`, 16, inode.I_block[15], 17, inode.I_block[16], 18, inode.I_block[17])

	// Hacer las direcciones de todos los nodos q este kabron saca del Iblock
	if !isTheRoot {
		dotContent += fmt.Sprintf(`node%d -> node%d
		`, nodoPadre, nodoActual)
	}

	for i, value := range inode.I_block {
		if value != -1 {
			if i >= 15 {
				block := &structures.PointerBlock{}
				err := block.Deserialize(diskPath, int64(sb.S_block_start+(value*sb.S_block_size)))
				if err != nil {
					return "", err
				}
				temp, err := getPointerBlockDOT(sb, block, nodoActual, int(value), diskPath, inode.I_type[0])
				if err != nil {
					return "", err
				}
				dotContent += temp
			} else if inode.I_type[0] == '0' { //Carpetas
				block := &structures.FolderBlock{}
				err := block.Deserialize(diskPath, int64(sb.S_block_start+(value*sb.S_block_size)))
				if err != nil {
					return "", err
				}
				temp, err := getFolderBlockDOT(sb, block, diskPath, nodoActual, int(value))
				if err != nil {
					return "", err
				}
				dotContent += temp
			} else { //File
				block := &structures.FileBlock{}
				err := block.Deserialize(diskPath, int64(sb.S_block_start+(value*sb.S_block_size)))
				if err != nil {
					return "", err
				}
				temp, err := getFileBlockDOT(block, nodoActual, int(value))
				if err != nil {
					return "", err
				}
				dotContent += temp
			}
		}
	}

	return dotContent, nil
}

func getFolderBlockDOT(sb *structures.SuperBlock, block *structures.FolderBlock, diskPath string, nodoPadre int, blockIndex int) (string, error) {
	nodoActual := getNode()
	dotContent := fmt.Sprintf(`node%d[fillcolor="#ec7063" style=filled shape=record label="Bloque Carpeta%d\nb_name : b_inodo\n`, nodoActual, blockIndex)
	for _, value := range block.B_content {
		tempString := string(value.B_name[:])
		nameTemp := strings.TrimRight(tempString, "\x00")
		dotContent += fmt.Sprintf(" %s : %d\\n", nameTemp, int32(value.B_inodo))
	}
	dotContent += `"];
	`
	dotContent += fmt.Sprintf(`node%d -> node%d
	`, nodoPadre, nodoActual)

	for i := 2; i < len(block.B_content); i++ {
		content := block.B_content[i]
		if content.B_inodo == -1 {
			continue
		}
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(sb.S_inode_size*content.B_inodo)))
		if err != nil {
			return "", err
		}
		temp, err := getInodeDOT(sb, inode, diskPath, nodoActual, false, int(content.B_inodo))
		if err != nil {
			return "", err
		}
		dotContent += temp
	}
	return dotContent, nil
}

func getFileBlockDOT(block *structures.FileBlock, nodoPadre int, blockIndex int) (string, error) {
	nodoActual := getNode()
	contentBlock := string(block.B_content[:])
	contentBlock = strings.TrimRight(contentBlock, "\x00")
	splitContent := splitEqualParts(contentBlock)
	dotContent := fmt.Sprintf(`node%d[fillcolor="#7dcea0" style=filled shape=record label="Bloque archivo %d\n%s\n%s\n%s\n%s"];
	`, nodoActual, blockIndex, splitContent[0], splitContent[1], splitContent[2], splitContent[3])
	dotContent += fmt.Sprintf(`node%d -> node%d
	`, nodoPadre, nodoActual)
	return dotContent, nil
}

func getPointerBlockDOT(sb *structures.SuperBlock, block *structures.PointerBlock, nodoPadre int, blockIndex int, diskPath string, tipoInodo byte) (string, error) {
	nodoActual := getNode()
	dotContent := fmt.Sprintf(`node%d[fillcolor="#f7dc6f" style=filled shape=record label="Bloque Apuntador%d\n`, nodoActual, blockIndex)
	for index, value := range block.P_pointers {
		if index%6 == 0 {
			dotContent += "\n"
		}
		dotContent += fmt.Sprintf(" %d,", value)
	}
	dotContent += `"];
	`
	dotContent += fmt.Sprintf(`node%d -> node%d
	`, nodoPadre, nodoActual)
	for _, indexInode := range block.P_pointers {
		if indexInode == -1 {
			break
		}
		if tipoInodo == '0' {
			block := &structures.FolderBlock{}
			err := block.Deserialize(diskPath, int64(sb.S_block_start+(indexInode*sb.S_block_size)))
			if err != nil {
				return "", err
			}
			temp, err := getFolderBlockDOT(sb, block, diskPath, nodoActual, int(indexInode))
			if err != nil {
				return "", err
			}
			dotContent += temp
		} else {
			block := &structures.FileBlock{}
			err := block.Deserialize(diskPath, int64(sb.S_block_start+(indexInode*sb.S_block_size)))
			if err != nil {
				return "", err
			}
			temp, err := getFileBlockDOT(block, nodoActual, int(indexInode))
			if err != nil {
				return "", err
			}
			dotContent += temp
		}
	}
	return dotContent, nil
}
