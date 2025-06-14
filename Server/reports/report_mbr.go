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

func ReportMBR(mbr *structures.MBR, path string, idDisk string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := fmt.Sprintf(`digraph G {
        node [shape=plaintext]
        tabla [label=<
            <table border="0" cellborder="1" cellspacing="0">
                <tr><td colspan="2" BGCOLOR="#aabbcc"> REPORTE MBR </td></tr>
                <tr><td BGCOLOR="#aabbcc">mbr_tamano</td><td>%d</td></tr>
                <tr><td BGCOLOR="#aabbcc">mrb_fecha_creacion</td><td>%s</td></tr>
                <tr><td BGCOLOR="#aabbcc">mbr_disk_signature</td><td>%d</td></tr>
            `, mbr.Mbr_size, time.Unix(int64(mbr.Mbr_creation_date), 0), mbr.Mbr_disk_signature)

	for i, part := range mbr.Mbr_partitions {

		if part.Part_type[0] == 'N' {
			continue
		}

		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
		partStatus := rune(part.Part_status[0])
		partType := rune(part.Part_type[0])
		partFit := rune(part.Part_fit[0])

		dotContent += fmt.Sprintf(`
				<tr><td colspan="2" BGCOLOR="#ccbbaa"> PARTICIÓN %d </td></tr>
				<tr><td BGCOLOR="#ccbbaa">part_status</td><td>%c</td></tr>
				<tr><td BGCOLOR="#ccbbaa">part_type</td><td>%c</td></tr>
				<tr><td BGCOLOR="#ccbbaa">part_fit</td><td>%c</td></tr>
				<tr><td BGCOLOR="#ccbbaa">part_start</td><td>%d</td></tr>
				<tr><td BGCOLOR="#ccbbaa">part_size</td><td>%d</td></tr>
				<tr><td BGCOLOR="#ccbbaa">part_name</td><td>%s</td></tr>
			`, i+1, partStatus, partType, partFit, part.Part_start, part.Part_size, partName)

		if part.Part_type[0] == 'E' {
			dotContent += ""
		}
	}

	dotContent += "</table>>] }"
	file, err := os.Create(dotFileName)

	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando Graphviz: %v", err)
	}

	return nil
}
