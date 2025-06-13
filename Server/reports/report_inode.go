package reports

import (
	"fmt"
	"os"
	"os/exec"
	structures "server/structures"
	utils "server/utils"
	"time"
)

func ReportInode(superBlock *structures.SuperBlock, diskPath, path string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := `digraph G {
        node [shape=plaintext]
		rankdir=LR;
    `

	for i := int32(0); i < superBlock.S_inodes_count; i++ {
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(superBlock.S_inode_start+(i*superBlock.S_inode_size)))
		if err != nil {
			return err
		}

		atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
		ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
		mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)

		dotContent += fmt.Sprintf(`inode%d [label=<
            <table border="0" cellborder="1" cellspacing="0">
                <tr><td colspan="2" BGCOLOR="#bbccaa"> REPORTE INODO %d </td></tr>
                <tr><td BGCOLOR="#bbccaa">i_uid</td><td>%d</td></tr>
                <tr><td BGCOLOR="#bbccaa">i_gid</td><td>%d</td></tr>
                <tr><td BGCOLOR="#bbccaa">i_size</td><td>%d</td></tr>
                <tr><td BGCOLOR="#bbccaa">i_atime</td><td>%s</td></tr>
                <tr><td BGCOLOR="#bbccaa">i_ctime</td><td>%s</td></tr>
                <tr><td BGCOLOR="#bbccaa">i_mtime</td><td>%s</td></tr>
                <tr><td BGCOLOR="#bbccaa">i_type</td><td>%c</td></tr>
                <tr><td BGCOLOR="#bbccaa">i_perm</td><td>%s</td></tr>
                <tr><td BGCOLOR="#bbccaa" colspan="2">BLOQUES DIRECTOS</td></tr>
            `, i, i, inode.I_uid, inode.I_gid, inode.I_size, atime, ctime, mtime, rune(inode.I_type[0]), string(inode.I_perm[:]))

		for j, block := range inode.I_block {
			if j > 14 {
				break
			}
			dotContent += fmt.Sprintf("<tr><td BGCOLOR=\"#bbccaa\">%d</td><td>%d</td></tr>", j+1, block)
		}

		dotContent += fmt.Sprintf(`
                <tr><td BGCOLOR="#bbccaa" colspan="2">BLOQUE INDIRECTO</td></tr>
                <tr><td BGCOLOR="#bbccaa">%d</td><td>%d</td></tr>
                <tr><td BGCOLOR="#bbccaa" colspan="2">BLOQUE INDIRECTO DOBLE</td></tr>
                <tr><td BGCOLOR="#bbccaa">%d</td><td>%d</td></tr>
                <tr><td BGCOLOR="#bbccaa" colspan="2">BLOQUE INDIRECTO TRIPLE</td></tr>
                <tr><td BGCOLOR="#bbccaa">%d</td><td>%d</td></tr>
            </table>>];
        `, 16, inode.I_block[15], 17, inode.I_block[16], 18, inode.I_block[17])

		if i < superBlock.S_inodes_count-1 {
			dotContent += fmt.Sprintf("inode%d -> inode%d;\n", i, i+1)
		}
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
