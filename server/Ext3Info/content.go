package ext3

import (
	"encoding/binary"
	"errors"
	"fmt"
	"server/reports"
	"server/stores"
	"server/structures"
	"server/utils"
	"strconv"
	"strings"
	"time"
)

func GetLogicPartitions(diskName string) ([]string, []string, error) {
	partitions := make([]string, 0)
	information := make([]string, 0)

	mbr := &structures.MBR{}
	path := stores.LoadedDiskPaths[diskName]
	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, err
	}
	for _, part := range mbr.Mbr_partitions {
		if part.Part_type[0] != 'E' {
			continue
		}
		return partitions, information, nil

	}
	return partitions, information, nil
}

func GetAllContentByPath(diskName, partitionName, pathToGetInfo string) ([]string, []string, []string, []string, error) {
	mbr := &structures.MBR{}
	var idPartition string
	var folderList []string
	var fileList []string
	var fileInfo []string
	var folderInfo []string

	diskPath := stores.LoadedDiskPaths[diskName]
	err := mbr.DeserializeMBR(diskPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for _, part := range mbr.Mbr_partitions {
		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
		DestinationPartName := strings.Trim(partitionName, "\x00")
		if strings.EqualFold(partName, DestinationPartName) {
			if part.Part_status[0] == '0' {
				return nil, nil, nil, nil, errors.New("particion no montada")
			}
			idPartition = strings.TrimRight(string(part.Part_id[:]), "\x00")
			break
		}
	}
	superBlock, _, _, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	inodoBase, _, err := reports.UbicarInodo(superBlock, pathToGetInfo, diskPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if inodoBase.I_type[0] == '1' {
		return nil, nil, nil, nil, errors.New("no se puede aplicar este reporte sobre un archivo")
	}

	for i, blockIndex := range inodoBase.I_block {
		if blockIndex == -1 {
			continue
		}
		if i >= 14 {
			pointerBlock := &structures.PointerBlock{}
			err := pointerBlock.Deserialize(diskPath, int64(superBlock.S_block_start+(superBlock.S_block_size*blockIndex)))
			if err != nil {
				return nil, nil, nil, nil, err
			}
			for _, value := range pointerBlock.P_pointers {
				if value == -1 {
					continue
				}
				block := &structures.FolderBlock{}
				err := block.Deserialize(diskPath, int64(superBlock.S_block_start+(value*superBlock.S_block_size)))
				if err != nil {
					return nil, nil, nil, nil, err
				}
				for i := 2; i < len(block.B_content); i++ {
					content := block.B_content[i]
					if content.B_inodo == -1 {
						continue
					}
					fileList, folderList, fileInfo, folderInfo, err = getInformationByInode(fileList, folderList, fileInfo, folderInfo, content.B_inodo, superBlock, diskPath, strings.Trim(string(content.B_name[:]), "\x00"), idPartition)
					if err != nil {
						return nil, nil, nil, nil, err
					}
				}
			}

		} else {
			block := &structures.FolderBlock{}
			err := block.Deserialize(diskPath, int64(superBlock.S_block_start+(blockIndex*superBlock.S_block_size)))
			if err != nil {
				return nil, nil, nil, nil, err
			}
			for i := 2; i < len(block.B_content); i++ {
				content := block.B_content[i]
				if content.B_inodo == -1 {
					continue
				}
				fileList, folderList, fileInfo, folderInfo, err = getInformationByInode(fileList, folderList, fileInfo, folderInfo, content.B_inodo, superBlock, diskPath, strings.Trim(string(content.B_name[:]), "\x00"), idPartition)
				if err != nil {
					return nil, nil, nil, nil, err
				}
			}
		}

	}
	return fileList, folderList, fileInfo, folderInfo, nil
}

func getInformationByInode(fileList, folderList, fileInfo, folderInfo []string, inodeIndex int32, sb *structures.SuperBlock, diskPath, contentName, idPartition string) ([]string, []string, []string, []string, error) {
	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(sb.S_inode_size*inodeIndex)))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	date := fmt.Sprintf("Date: %s", time.Unix(int64(inode.I_ctime), 0).Format("2006-01-02"))
	permissions := fmt.Sprintf("Permissions: %s ", string(inode.I_perm[:]))
	userID := inode.I_uid
	groupID := inode.I_gid
	userName, err := getOwnerByID(userID, idPartition)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	user := fmt.Sprintf("User: %s ", userName)
	groupName, err := getGroupById(groupID, idPartition)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	group := fmt.Sprintf("Group: %s ", groupName)
	if inode.I_type[0] == '0' {
		folderList = append(folderList, contentName)
		folderInfo = append(folderInfo, user+group+permissions+date)
	} else if inode.I_type[0] == '1' {
		fileList = append(fileList, contentName)
		size := fmt.Sprintf("Size: %d ", inode.I_size)
		fileInfo = append(fileInfo, user+group+permissions+size+date)
	}

	return fileList, folderList, fileInfo, folderInfo, nil
}

func GetContentFromFile(diskName, partitionName, pathToGetInfo string) (string, error) {
	mbr := &structures.MBR{}
	var result string
	var idPartition string
	diskPath := stores.LoadedDiskPaths[diskName]
	err := mbr.DeserializeMBR(diskPath)
	if err != nil {
		return "", err
	}
	// Se obtiene el id
	for _, part := range mbr.Mbr_partitions {
		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
		DestinationPartName := strings.Trim(partitionName, "\x00")
		if strings.EqualFold(partName, DestinationPartName) {
			if part.Part_status[0] == '0' {
				return "", errors.New("particion no montada")
			}
			idPartition = strings.TrimRight(string(part.Part_id[:]), "\x00")
			break
		}
	}
	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "", err
	}
	parentDirs, destDir := utils.GetParentDirectories(pathToGetInfo)
	content, err := partitionSuperblock.ContentFromFile(partitionPath, 0, parentDirs, destDir)
	if err != nil {
		return "", err
	}
	result += content
	return result, nil
}

func getOwnerByID(id int32, idPartition string) (string, error) {
	contentUsersTxt, err := reports.GetContetnUsersTxt(idPartition)
	if err != nil {
		return "", err
	}
	strId := strconv.Itoa(int(id))
	contentMatrix := reports.GetContentMatrixUsers(contentUsersTxt)
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

func getGroupById(id int32, idPartition string) (string, error) {
	contentUsersTxt, err := reports.GetContetnUsersTxt(idPartition)
	if err != nil {
		return "", err
	}
	strId := strconv.Itoa(int(id))
	contentMatrix := reports.GetContentMatrixUsers(contentUsersTxt)
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

func GetJournal(diskName, partitionName string) ([]string, []string, []string, []string, error) {
	mbr := &structures.MBR{}
	var partitionStart int32
	diskPath := stores.LoadedDiskPaths[diskName]
	err := mbr.DeserializeMBR(diskPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for _, part := range mbr.Mbr_partitions {
		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
		DestinationPartName := strings.Trim(partitionName, "\x00")
		if strings.EqualFold(partName, DestinationPartName) {
			if part.Part_status[0] == '0' {
				return nil, nil, nil, nil, errors.New("particion no montada")
			}
			partitionStart = part.Part_start
			break
		}
	}
	journal := &structures.Journal{}
	err = journal.Deserialize(diskPath, int64(partitionStart+int32(binary.Size(structures.SuperBlock{}))))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	commandList, pathList, contentList, dateList, err := getInformationJournals(journal, diskPath, make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return commandList, pathList, contentList, dateList, nil
}

func GetJournalForCommand(diskPath string, partitionStart int32) ([]string, []string, []string, []string, error) {
	mbr := &structures.MBR{}
	err := mbr.DeserializeMBR(diskPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	journal := &structures.Journal{}
	err = journal.Deserialize(diskPath, int64(partitionStart+int32(binary.Size(structures.SuperBlock{}))))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	commandList, pathList, contentList, dateList, err := getInformationJournals(journal, diskPath, make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return commandList, pathList, contentList, dateList, nil
}

func getInformationJournals(journal *structures.Journal, diskPath string, commandList, pathList, contentList, dateList []string) ([]string, []string, []string, []string, error) {
	command := string(journal.J_content.I_operation[:])
	path := string(journal.J_content.I_path[:])
	content := string(journal.J_content.I_content[:])
	date := time.Unix(int64(journal.J_content.I_date), 0).Format("2006-01-02")

	commandList = append(commandList, command)
	pathList = append(pathList, path)
	contentList = append(contentList, content)
	dateList = append(dateList, date)

	if journal.J_next == -1 {
		return commandList, pathList, contentList, dateList, nil
	}

	neoJournal := &structures.Journal{}
	err := neoJournal.Deserialize(diskPath, int64(journal.J_next))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return getInformationJournals(neoJournal, diskPath, commandList, pathList, contentList, dateList)
}

func IsExt3(diskName, partitionName string) (bool, error) {
	mbr := &structures.MBR{}
	var idPartition string
	diskPath := stores.LoadedDiskPaths[diskName]
	err := mbr.DeserializeMBR(diskPath)
	if err != nil {
		return false, err
	}
	// Se obtiene el id
	for _, part := range mbr.Mbr_partitions {
		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
		DestinationPartName := strings.Trim(partitionName, "\x00")
		if strings.EqualFold(partName, DestinationPartName) {
			if part.Part_status[0] == '0' {
				return false, errors.New("particion no montada")
			}
			idPartition = strings.TrimRight(string(part.Part_id[:]), "\x00")
			break
		}
	}
	partitionSuperblock, _, _, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return false, err
	}

	if partitionSuperblock.S_filesystem_type == 2 {
		return false, nil
	} else {
		return true, nil
	}
}
