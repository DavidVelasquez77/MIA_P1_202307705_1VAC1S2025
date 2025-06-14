package commands

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"regexp"
	stores "server/stores"
	structures "server/structures"
	"strings"
	"time"
)

type MKFS struct {
	id  string
	typ bool //Si viene el parametro
	fs  string
}

func ParseMkfs(tokens []string) (string, error) {
	cmd := &MKFS{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[a-zA-Z0-9]+|-type=[fFuUlL]+|-fs=[23]fs`)
	matches := re.FindAllString(args, -1)
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parametro invalido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacio")
			}
			cmd.id = value
		case "-type":
			value = strings.ToLower(value)
			if value != "full" {
				return "", errors.New("solo se acepta el tipo full")
			}
			cmd.typ = true
		case "-fs":
			if value != "2fs" && value != "3fs" {
				return "", errors.New("el sistema de archivos debe ser 2fs o 3fs")
			}

			cmd.fs = value
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)
		}
	}
	if cmd.id == "" {
		return "", errors.New("faltan parametros requeridos: -id")
	}
	if cmd.fs == "" {
		cmd.fs = "2fs"
	}
	err := commandMkfs(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKFS: %s formateado exitosamente", cmd.id), nil
}

func commandMkfs(mkfs *MKFS) error {
	mountedPartition, partitionPath, err := stores.GetMountedPartition(mkfs.id)
	if err != nil {
		return err
	}

	// fmt.Println("\nPatición montada:")
	// mountedPartition.PrintPartition()

	n := calculateN(mountedPartition, mkfs.fs)

	// fmt.Println("\nValor de n: ", n)

	superBlock := createSuperBlock(mountedPartition, n, mkfs.fs)

	// fmt.Println("\nSuperBlock:")
	// superBlock.Print()

	err = superBlock.CreateBitMaps(partitionPath)
	if err != nil {
		return err
	}

	if mkfs.fs == "3fs" {
		// Crear archivo users.txt ext3
		err = superBlock.CreateUsersFile(partitionPath, int64(mountedPartition.Part_start+int32(binary.Size(structures.SuperBlock{}))))
		if err != nil {
			return err
		}
	} else {
		// Crear archivo users.txt ext2
		err = superBlock.CreateUsersFile(partitionPath, 0)
		if err != nil {
			return err
		}
	}

	err = superBlock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return err
	}

	return nil

}

func calculateN(partition *structures.PARTITION, fs string) int32 {
	numerator := int(partition.Part_size) - binary.Size(structures.SuperBlock{})

	baseDenominator := 4 + binary.Size(structures.Inode{}) + 3*binary.Size(structures.FileBlock{})

	temp := 0
	if fs == "3fs" {
		temp = binary.Size(structures.Journal{})
	}
	// Denominador final
	denominator := baseDenominator + temp

	// Calcular n
	n := math.Floor(float64(numerator) / float64(denominator))

	return int32(n)
}

func createSuperBlock(partition *structures.PARTITION, n int32, fs string) *structures.SuperBlock {
	_, bm_inode_start, bm_block_start, inode_start, block_start := calculateStartPositions(partition, fs, n)

	var fsType int32
	if fs == "2fs" {
		fsType = 2
	} else {
		fsType = 3
	}

	superBlock := &structures.SuperBlock{
		S_filesystem_type:   fsType,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_inodes_count: int32(n),
		S_free_blocks_count: int32(n * 3),
		S_mtime:             float32(time.Now().Unix()),
		S_umtime:            float32(time.Now().Unix()),
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_size:        int32(binary.Size(structures.Inode{})),
		S_block_size:        int32(binary.Size(structures.FileBlock{})),
		S_first_ino:         inode_start,
		S_first_blo:         block_start,
		S_bm_inode_start:    bm_inode_start,
		S_bm_block_start:    bm_block_start,
		S_inode_start:       inode_start,
		S_block_start:       block_start,
	}
	return superBlock
}

func calculateStartPositions(partition *structures.PARTITION, fs string, n int32) (int32, int32, int32, int32, int32) {
	superBlockSize := int32(binary.Size(structures.SuperBlock{}))
	journalSize := int32(binary.Size(structures.Journal{}))
	inodeSize := int32(binary.Size(structures.Inode{}))

	journalStart := int32(0)
	bmInodeStart := partition.Part_start + superBlockSize
	bmBlockStart := bmInodeStart + n
	inodeStart := bmBlockStart + (3 * n)
	blockStart := inodeStart + (inodeSize * n)

	if fs == "3fs" {
		journalStart = partition.Part_start + superBlockSize
		bmInodeStart = journalStart + (journalSize * n)
		bmBlockStart = bmInodeStart + n
		inodeStart = bmBlockStart + (3 * n)
		blockStart = inodeStart + (inodeSize * n)
	}
	return journalStart, bmInodeStart, bmBlockStart, inodeStart, blockStart
}
