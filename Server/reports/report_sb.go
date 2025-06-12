package reports

import (
	"fmt"
	"os"
	"os/exec"
	structures "server/structures"
	utils "server/utils"
	"time"
)

func ReportSuperBlock(sb *structures.SuperBlock, diskPath, path string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := `digraph G {
        node [shape=plaintext]
		rankdir=LR;
    `
	mtime := time.Unix(int64(sb.S_mtime), 0).Format(time.RFC3339)
	umtime := time.Unix(int64(sb.S_umtime), 0).Format(time.RFC3339)
	dotContent += fmt.Sprintf(`inode [label=<
            <table border="0" cellborder="1" cellspacing="0">
                <tr><td colspan="2" BGCOLOR="#aaccbb"> REPORTE SUPERBLOCK</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_filesystem_type</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_inodes_count</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_blocks_count</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_free_inodes_count</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_free_blocks_count</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_mtime</td><td>%s</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_umtime</td><td>%s</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_mnt_count</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_magic</td><td>0xEF53</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_inode_size</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_block_size</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_first_ino</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_first_blo</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_bm_inode_start</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_bm_block_start</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_inode_start</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aaccbb">S_block_start</td><td>%d</td></tr>
				 </table>>];
            `, sb.S_filesystem_type, sb.S_inodes_count, sb.S_blocks_count, sb.S_free_inodes_count, sb.S_free_blocks_count, mtime, umtime, sb.S_mnt_count, sb.S_inode_size, sb.S_block_size, sb.S_first_ino, sb.S_first_blo, sb.S_bm_inode_start, sb.S_bm_block_start, sb.S_inode_start, sb.S_block_start)

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
