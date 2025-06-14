package reports

import (
	"os"
	"server/structures"
	"server/utils"
)

func ReportFile(sb *structures.SuperBlock, diskPath, path string, pathFileToGetInfo string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}
	parentDirs, destDir := utils.GetParentDirectories(pathFileToGetInfo)

	content, err := sb.ContentFromFileCat(diskPath, 0, parentDirs, destDir)
	if err != nil {
		return err
	}
	txtFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer txtFile.Close()
	_, err = txtFile.WriteString(content)
	if err != nil {
		return err
	}
	return nil

}
