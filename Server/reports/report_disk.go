package reports

import (
	"fmt"
	"os"
	"os/exec"
	"server/stores"
	"server/structures"
	"server/utils"
)

var contador int32 = 0

func ReportDisk(mbr *structures.MBR, idDisk string, path string, pathDisk string) error {
	contador = 0
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := fmt.Sprintf(`digraph G{
	label = "%s";
	`, stores.GetNameDisk(idDisk))

	tamanoTotalDisco := mbr.Mbr_size

	temp := float64(153) / float64(tamanoTotalDisco)
	percentageMBR := temp * 100

	dotContent += fmt.Sprintf(`node%d[shape=record, label="%s"];
	`, getNumberNode(), "MBR")
	percentageUsed := percentageMBR

	for _, partition := range mbr.Mbr_partitions {
		var tipoParticion string
		if partition.Part_type[0] == 'P' {
			tipoParticion = "Primaria"
		} else if partition.Part_type[0] == 'E' {
			tipoParticion = "Extendida"
		} else {
			//Significa que no esta siendo utilizada esta particion
			break
		}
		percentagePartition := (float64(partition.Part_size) / float64(tamanoTotalDisco)) * 100
		percentageUsed += percentagePartition
		if partition.Part_type[0] == 'E' {
			dotContent += fmt.Sprintf(`subgraph cluster_2 {
			label="%s";
			rankdir=LR
			`, tipoParticion)

			dotContent += fmt.Sprintf(`node%d[shape=record, label="%s\n%d%%"];
			}`, getNumberNode(), "Libre", 100)
		} else {
			dotContent += fmt.Sprintf(`node%d[shape=record, label="%s\n%.1f%%"];
			`, getNumberNode(), tipoParticion, percentagePartition)
		}
	}

	dotContent += fmt.Sprintf(`node%d[shape=record, label="%s\n%.1f%%"];
	}`, getNumberNode(), "Libre", 100-percentageUsed)

	//Creacion del dot y el merquetenge
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

func getNumberNode() int32 {
	contador++
	return contador
}
