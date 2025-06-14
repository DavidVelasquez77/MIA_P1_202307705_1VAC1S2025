package commands

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"server/stores"
	"server/structures"
	"server/utils"
	"strings"
	"time"
)

type MKGRP struct {
	name string
}

func ParseMkgrp(tokens []string) (string, error) {
	cmd := &MKGRP{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-name="[^"]+"|-name=[^\s]+`)
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
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede venir vacio")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parametro desconocido: %s", key)
		}
	}

	if cmd.name == "" {
		return "", errors.New("parametro obligatorio: -name")
	}

	err := CommmandMkgrp(cmd)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("MKGRP: grupo %s creado exitosamente", cmd.name), nil
}

func CommmandMkgrp(mkgrp *MKGRP) error {
	if stores.LogedIdPartition == "" {
		return errors.New("no hay sesion activa")
	}
	if stores.LogedUser != "root" {
		return errors.New("este comando solo lo puede ejecutar el usuario root")
	}
	contentUsersTxt, err := getContetnUsersTxt(stores.LogedIdPartition)
	if err != nil {
		return err
	}
	contentMatrix := getContentMatrixUsers(contentUsersTxt)
	outcome := soleNameGroup(mkgrp.name, contentMatrix)
	if !outcome {
		return errors.New("el nombre de grupo ya esta siendo utilizado")
	}
	neoGroupID := getNeoNumber("G", contentMatrix)

	contentUsersTxt += fmt.Sprintf("%d,G,%s\n", neoGroupID, mkgrp.name)
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(stores.LogedIdPartition)
	if err != nil {
		return err
	}
	err = OverrideUserstxt(partitionSuperblock, partitionPath, contentUsersTxt)
	if err != nil {
		return err
	}

	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return err
	}
	if partitionSuperblock.IsExt3() {
		journalDirectory := &structures.Journal{
			J_next: -1,
			J_content: structures.Information{
				I_operation: [10]byte{'m', 'k', 'g', 'r', 'p'},
				I_path:      [74]byte{},
				I_content:   [64]byte{},
				I_date:      float32(time.Now().Unix()),
			},
		}
		copy(journalDirectory.J_content.I_content[:], mkgrp.name)
		err = partitionSuperblock.AddJournal(journalDirectory, partitionPath, int32(mountedPartition.Part_start+int32(binary.Size(structures.SuperBlock{}))))
		if err != nil {
			return err
		}
	}
	// err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	// if err != nil {
	// 	return err
	// }
	return nil
}

func soleNameGroup(nameGroup string, matrix [][]string) bool {
	for _, row := range matrix {
		if row[1] != "G" {
			continue
		}
		if row[2] == nameGroup {
			return false
		}
	}
	return true
}

func getNeoNumber(tipo string, matrix [][]string) int {
	var highestNumber int
	for _, row := range matrix {
		if row[1] != tipo {
			continue
		} else {
			highestNumber++
		}

	}
	highestNumber++
	return highestNumber
}

func OverrideUserstxt(sb *structures.SuperBlock, diskPath, content string) error {
	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size))
	if err != nil {
		return err
	}
	contentChunks := utils.SplitStringIntoChunks(content)
	for i, indexFileBlock := range inode.I_block {
		if len(contentChunks) == 0 {
			break
		}
		if indexFileBlock != -1 {
			fileBlock := &structures.FileBlock{}
			err := fileBlock.Deserialize(diskPath, int64(sb.S_block_start+(indexFileBlock*sb.S_block_size)))
			if err != nil {
				return err
			}
			copy(fileBlock.B_content[:], []byte(contentChunks[0]))
			contentChunks = utils.RemoveElement(contentChunks, 0)
			fileBlock.Serialize(diskPath, int64(sb.S_block_start+(indexFileBlock*sb.S_block_size)))
		} else {
			if len(contentChunks) == 0 {
				break
			}
			inode.I_block[i] = sb.S_blocks_count
			contentBlock := &structures.FileBlock{
				B_content: [64]byte{},
			}
			copy(contentBlock.B_content[:], []byte(contentChunks[0]))
			contentChunks = utils.RemoveElement(contentChunks, 0)
			err = contentBlock.Serialize(diskPath, int64(sb.S_first_blo))
			if err != nil {
				return err
			}

			err = sb.UpdateBitmapBlock(diskPath)
			if err != nil {
				return err
			}
			// fmt.Println("EL CONTADOR:", sb.S_blocks_count)
			sb.S_blocks_count++
			// fmt.Println("EL CONTADOR NUEVO:", sb.S_blocks_count)
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size

		}
	}
	inode.I_size = int32(len(content))
	err = inode.Serialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size))
	if err != nil {
		return err
	}
	return nil
}
