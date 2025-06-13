package stores

import (
	"errors"
	"path/filepath"
	"server/structures"
	"server/utils"
	"strings"
)

const Carnet string = "05" //202307705

var (
	MountedPartitions map[string]string = make(map[string]string) //ID:path
	LogedIdPartition  string            = ""
	LogedUser         string            = ""
	LoadedDiskPaths   map[string]string = make(map[string]string) //Nombre:path
)

func GetMountedPartition(id string) (*structures.PARTITION, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, "", errors.New("la particion no esta montada")
	}
	var mbr structures.MBR

	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, "", err
	}

	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, "", err
	}
	return partition, path, nil

}

func DeleteMountedPartitions(path string) {
	for key, value := range MountedPartitions {
		if value == path {
			delete(MountedPartitions, key)
			delete(utils.PathToPartitionCount, path)
			// delete(utils.PathToLetter, path)
		}
	}
	for key, value := range LoadedDiskPaths {
		if value == path {
			delete(LoadedDiskPaths, key)
		}
	}
}

func GetMountedPartitionRep(id string) (*structures.MBR, *structures.SuperBlock, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la particion no esta montada")
	}

	var mbr structures.MBR
	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}
	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, nil, "", err
	}

	var sb structures.SuperBlock

	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &mbr, &sb, path, nil

}

func GetNameDisk(idDisk string) string {
	pathDisk := MountedPartitions[idDisk]
	diskLetter := utils.GetDiskLetterFromPath(pathDisk)
	if diskLetter != "" {
		return diskLetter
	}
	// Fallback al método anterior si no se encuentra letra
	baseName := strings.TrimSuffix(filepath.Base(pathDisk), filepath.Ext(pathDisk))
	return baseName
}

func GetMountedPartitionSuperblock(id string) (*structures.SuperBlock, *structures.PARTITION, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la particion no esta montada")
	}

	var mbr structures.MBR

	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}

	partition, err := mbr.GetPartitionByID(id)
	if err != nil {
		return nil, nil, "", err
	}

	var sb structures.SuperBlock

	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &sb, partition, path, nil
}
