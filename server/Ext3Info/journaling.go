package ext3

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	stores "server/stores"
	"server/utils"
	"strings"
)

func ReportJournaling(id, path string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}
	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := `digraph G {
    node [shape=plaintext]

    tabla1 [label=<
        <TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0">
		<TR>
                <TD><B>Command</B></TD>
                <TD><B>Path</B></TD>
                <TD><B>Content</B></TD>
                <TD><B>Date</B></TD>
		</TR>
    `
	commandList, pathList, contentList, dateList, err := getContentJournaling(id)
	if err != nil {
		return err
	}
	for i, _ := range commandList {
		dotContent += fmt.Sprintf(`
		<TR>
			<TD>%s</TD>
			<TD>%s</TD>
			<TD>%s</TD>
			<TD>%s</TD>
		</TR>
		`, strings.TrimRight(string(commandList[i]), "\x00"), strings.TrimRight(string(pathList[i]), "\x00"), strings.TrimRight(string(contentList[i]), "\x00"), strings.TrimRight(string(dateList[i]), "\x00"))
	}
	dotContent += `</TABLE>
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

func getContentJournaling(id string) ([]string, []string, []string, []string, error) {
	sb, part, diskPath, err := stores.GetMountedPartitionSuperblock(id)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	outcome := sb.IsExt3()
	if !outcome {
		return nil, nil, nil, nil, errors.New("este comando no es aplicable porque el sistema de archivos no es ext3")
	}
	return GetJournalForCommand(diskPath, part.Part_start)

}
