package structures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"regexp"
	utils "server/utils"
	"strings"
	"time"
)

type SuperBlock struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_inodes_count int32
	S_free_blocks_count int32
	S_mtime             float32
	S_umtime            float32
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

func (sb *SuperBlock) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.LittleEndian, sb)
	if err != nil {
		return err
	}
	return nil
}

func (sb *SuperBlock) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}
	sbSize := binary.Size(sb)
	if sbSize <= 0 {
		return fmt.Errorf("invalid superblock size: %d", sbSize)
	}
	buffer := make([]byte, sbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, sb)
	if err != nil {
		return err
	}
	return nil
}

func (sb *SuperBlock) CreateUsersFile(path string, journauling_start int64) error {
	rootInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	err := rootInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	err = sb.UpdateBitmapInode(path)
	if err != nil {
		return err
	}

	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},
		},
	}

	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	err = rootBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// Creamos el journal

	if journauling_start != 0 {
		journal := &Journal{
			J_next: -1,
			J_content: Information{
				I_operation: [10]byte{'m', 'k', 'd', 'i', 'r'},
				I_path:      [74]byte{'/'},
				I_content:   [64]byte{},
				I_date:      float32(time.Now().Unix()),
			},
		}
		err = journal.Serialize(path, journauling_start)
		if err != nil {
			return err
		}
	}

	usersText := "1,G,root\n1,U,root,root,123\n"

	err = rootInode.Deserialize(path, int64(sb.S_inode_start+0))
	if err != nil {
		return err
	}

	rootInode.I_atime = float32(time.Now().Unix())

	err = rootInode.Serialize(path, int64(sb.S_inode_start+0))
	if err != nil {
		return err
	}

	err = rootBlock.Deserialize(path, int64(sb.S_block_start+0))
	if err != nil {
		return err
	}

	rootBlock.B_content[2] = FolderContent{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count}

	err = rootBlock.Serialize(path, int64(sb.S_block_start+0))
	if err != nil {
		return err
	}
	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}
	err = sb.UpdateBitmapInode(path)
	if err != nil {
		return err
	}

	err = usersInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Crear Journal
	if journauling_start != 0 {
		journalFile := &Journal{
			J_next: -1,
			J_content: Information{
				I_operation: [10]byte{'m', 'k', 'f', 'i', 'l', 'e'},
				I_path:      [74]byte{'/', 'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'},
				I_content:   [64]byte{},
				I_date:      float32(time.Now().Unix()),
			},
		}
		// Copiamos el texto de usuarios en el journal
		copy(journalFile.J_content.I_content[:], usersText)

		// err = journalFile.Serialize(path, journauling_start)
		err = sb.AddJournal(journalFile, path, int32(journauling_start))
		if err != nil {
			return err
		}
	}

	usersBlock := &FileBlock{
		B_content: [64]byte{},
	}

	copy(usersBlock.B_content[:], usersText)

	err = usersBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// fmt.Println("\nInodo Raíz Actualizado:")
	// rootInode.Print()

	// fmt.Println("\nBloque de Carpeta Raíz Actualizado:")
	// rootBlock.Print()

	// fmt.Println("\nInodo users.txt:")
	// usersInode.Print()

	// fmt.Println("\nBloque de users.txt:")
	// usersBlock.Print()

	return nil
}

func (sb *SuperBlock) Print() {
	mountTime := time.Unix(int64(sb.S_mtime), 0)
	unmountTime := time.Unix(int64(sb.S_umtime), 0)

	fmt.Printf("Filesystem Type: %d\n", sb.S_filesystem_type)
	fmt.Printf("Inodes Count: %d\n", sb.S_inodes_count)
	fmt.Printf("Blocks Count: %d\n", sb.S_blocks_count)
	fmt.Printf("Free Inodes Count: %d\n", sb.S_free_inodes_count)
	fmt.Printf("Free Blocks Count: %d\n", sb.S_free_blocks_count)
	fmt.Printf("Mount Time: %s\n", mountTime.Format(time.RFC3339))
	fmt.Printf("Unmount Time: %s\n", unmountTime.Format(time.RFC3339))
	fmt.Printf("Mount Count: %d\n", sb.S_mnt_count)
	fmt.Printf("Magic: %d\n", sb.S_magic)
	fmt.Printf("Inode Size: %d\n", sb.S_inode_size)
	fmt.Printf("Block Size: %d\n", sb.S_block_size)
	fmt.Printf("First Inode: %d\n", sb.S_first_ino)
	fmt.Printf("First Block: %d\n", sb.S_first_blo)
	fmt.Printf("Bitmap Inode Start: %d\n", sb.S_bm_inode_start)
	fmt.Printf("Bitmap Block Start: %d\n", sb.S_bm_block_start)
	fmt.Printf("Inode Start: %d\n", sb.S_inode_start)
	fmt.Printf("Block Start: %d\n", sb.S_block_start)
}

func (sb *SuperBlock) PrintInodes(path string) error {
	fmt.Println("\nInodos\n----------------")
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		fmt.Printf("\nInodo %d:\n", i)
		inode.Print()
	}

	return nil
}

func (sb *SuperBlock) PrintBlocks(path string) error {
	fmt.Println("\nBloques\n----------------")
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				continue
			}
			if inode.I_type[0] == '0' {
				block := &FolderBlock{}
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue

			} else if inode.I_type[0] == '1' {
				block := &FileBlock{}
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue
			}

		}
	}

	return nil
}

func (sb *SuperBlock) CreateFolder(path string, parentsDir []string, destDir string, flag bool) error {
	if !flag {
		return sb.createFolderInInode(path, 0, parentsDir, destDir, false)
	} else {
		return sb.createFolderInInodeWithP(path, 0, parentsDir, destDir)
	}
}

func (sb *SuperBlock) AddJournal(neoJournal *Journal, path string, offset int32) error {
	journal := &Journal{}
	err := journal.Deserialize(path, int64(offset))
	if err != nil {
		return err
	}
	if journal.J_next != -1 {
		return sb.AddJournal(neoJournal, path, journal.J_next)
	}
	position := offset + int32(binary.Size(Journal{}))
	journal.J_next = position
	err = journal.Serialize(path, int64(offset))
	if err != nil {
		return err
	}
	neoJournal.Serialize(path, int64(position))
	return nil
}

func (sb *SuperBlock) IsExt3() bool {
	return sb.S_filesystem_type == 3
}

func (sb *SuperBlock) ChmodRecursive(diskPath string, indexInode int32, permissions string, userLogedId, userGroupID int32) error {
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+indexInode*sb.S_inode_size))
	if err != nil {
		return err
	}
	outcome, err := inode.HasPermissionsChmod(userLogedId, userGroupID)
	if err != nil {
		return err
	}
	if outcome {
		copy(inode.I_perm[:], []byte(permissions))
		err = inode.Serialize(diskPath, int64(sb.S_inode_start+indexInode*sb.S_inode_size))
		if err != nil {
			return err
		}
		if inode.I_type[0] == '0' {
			for i, blockIndex := range inode.I_block {
				if blockIndex == -1 {
					continue
				}
				if i >= 14 {
					pointerBlock := &PointerBlock{}
					err := pointerBlock.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
					if err != nil {
						return err
					}
					for _, value := range pointerBlock.P_pointers {
						if value == -1 {
							continue
						}
						block := &FolderBlock{}
						err := block.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*value))
						if err != nil {
							return err
						}
						for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
							content := block.B_content[indexContent]
							if content.B_inodo == -1 {
								continue
							}
							err := sb.ChmodRecursive(diskPath, content.B_inodo, permissions, userLogedId, userGroupID)
							if err != nil {
								return err
							}
						}
					}
				} else {
					block := &FolderBlock{}
					err := block.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
					if err != nil {
						return err
					}
					for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
						content := block.B_content[indexContent]
						if content.B_inodo == -1 {
							continue
						}
						err := sb.ChmodRecursive(diskPath, content.B_inodo, permissions, userLogedId, userGroupID)
						if err != nil {
							return err
						}
					}

				}
			}
		}
	}
	return nil
}

func (sb *SuperBlock) ChownRecursive(diskPath string, indexInode int32, userLogedId, userGroupID int32, userIdNeoOwner int32, isTheFirtOne bool) error {
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+indexInode*sb.S_inode_size))
	if err != nil {
		return err
	}
	var outcome bool
	if !isTheFirtOne {
		outcome, err = inode.HasPermissionsChmod(userLogedId, userGroupID)
		if err != nil {
			return err
		}
	} else {
		outcome = true
	}

	if outcome {

		inode.I_uid = int32(userIdNeoOwner)
		err = inode.Serialize(diskPath, int64(sb.S_inode_start+indexInode*sb.S_inode_size))
		if err != nil {
			return err
		}
		if inode.I_type[0] == '0' {
			for i, blockIndex := range inode.I_block {
				if blockIndex == -1 {
					continue
				}
				if i >= 14 {
					pointerBlock := &PointerBlock{}
					err := pointerBlock.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
					if err != nil {
						return err
					}
					for _, value := range pointerBlock.P_pointers {
						if value == -1 {
							continue
						}
						block := &FolderBlock{}
						err := block.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*value))
						if err != nil {
							return err
						}
						for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
							content := block.B_content[indexContent]
							if content.B_inodo == -1 {
								continue
							}
							err := sb.ChownRecursive(diskPath, content.B_inodo, userLogedId, userGroupID, userIdNeoOwner, false)
							if err != nil {
								return err
							}
						}
					}
				} else {
					block := &FolderBlock{}
					err := block.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
					if err != nil {
						return err
					}
					for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
						content := block.B_content[indexContent]
						if content.B_inodo == -1 {
							continue
						}
						err := sb.ChownRecursive(diskPath, content.B_inodo, userLogedId, userGroupID, userIdNeoOwner, false)
						if err != nil {
							return err
						}
					}

				}
			}
		}
	}
	return nil
}

func (sb *SuperBlock) MoveTreePermissions(diskPath string, indexInode int32, userLogedId, userGroupID int32) error {
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+indexInode*sb.S_inode_size))
	if err != nil {
		return err
	}
	outcome, err := inode.HasPermissionsToWrite(userLogedId, userGroupID)
	if err != nil {
		return err
	}
	if !outcome {
		return errors.New("hay un inodo que no tiene los permisos adecuados para tal accion")
	}
	if inode.I_type[0] == '0' {
		for i, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				continue
			}
			if i >= 14 {
				pointerBlock := &PointerBlock{}
				err := pointerBlock.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
				if err != nil {
					return err
				}
				for _, value := range pointerBlock.P_pointers {
					if value == -1 {
						continue
					}
					block := &FolderBlock{}
					err := block.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*value))
					if err != nil {
						return err
					}
					for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
						content := block.B_content[indexContent]
						if content.B_inodo == -1 {
							continue
						}
						err := sb.MoveTreePermissions(diskPath, content.B_inodo, userLogedId, userGroupID)
						if err != nil {
							return err
						}
					}
				}
			} else {
				block := &FolderBlock{}
				err := block.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
				if err != nil {
					return err
				}
				for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
					content := block.B_content[indexContent]
					if content.B_inodo == -1 {
						continue
					}
					err := sb.MoveTreePermissions(diskPath, content.B_inodo, userLogedId, userGroupID)
					if err != nil {
						return err
					}
				}

			}
		}
	}
	return nil
}

func (sb *SuperBlock) TypeOfInode(diskPath string, indexInode int32) (int32, error) {
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size*indexInode))
	if err != nil {
		return -1, err
	}
	if inode.I_type[0] == '0' {
		return 0, nil
	} else if inode.I_type[0] == '1' {
		return 1, nil
	}
	return -1, errors.New("ha ocurrido un error inesperado al saber el tipo del inodo")
}

func (sb *SuperBlock) CopyInode0(diskPath string, indexInodeToCopy, indexInodoPadre int32) (int32, error) { // Inodo tipo folder
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size*indexInodeToCopy))
	if err != nil {
		return -1, err
	}
	outcome, err := inode.HasPermissionsToRead(utils.LogedUserID, utils.LogedUserGroupID)
	if err != nil {
		return -1, err
	}
	if !outcome {
		return -1, nil
	}
	// Creamos el inodo
	resultIndex := sb.S_inodes_count
	inodoCopia := &Inode{
		I_uid:   inode.I_uid,
		I_gid:   inode.I_gid,
		I_size:  inode.I_size,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  inode.I_type,
		I_perm:  inode.I_perm,
	}
	offsetInodoCopia := int64(sb.S_first_ino)
	err = inodoCopia.Serialize(diskPath, offsetInodoCopia)
	if err != nil {
		return -1, err
	}
	err = sb.UpdateBitmapInode(diskPath)
	if err != nil {
		return -1, err
	}
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size
	//

	for i, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}
		// Creamos un folderblock
		inodoCopia.I_block[i] = sb.S_blocks_count
		folderBlock := &FolderBlock{
			B_content: [4]FolderContent{
				{B_name: [12]byte{'.'}, B_inodo: resultIndex},
				{B_name: [12]byte{'.', '.'}, B_inodo: indexInodoPadre},
				{B_name: [12]byte{'-'}, B_inodo: -1},
				{B_name: [12]byte{'-'}, B_inodo: -1},
			},
		}
		offsetFolderBlock := int64(sb.S_first_blo)
		err = folderBlock.Serialize(diskPath, offsetFolderBlock)
		if err != nil {
			return -1, err
		}
		err = sb.UpdateBitmapBlock(diskPath)
		if err != nil {
			return -1, err
		}
		sb.S_blocks_count++
		sb.S_free_blocks_count--
		sb.S_first_blo += sb.S_block_size
		//

		block := &FolderBlock{}
		err := block.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
		if err != nil {
			return -1, err
		}
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]
			if content.B_inodo == -1 {
				continue
			}
			tipoInodo, err := sb.TypeOfInode(diskPath, content.B_inodo)
			if err != nil {
				return -1, err
			}
			var inodoAIndexar int32
			if tipoInodo == 0 {
				inodoAIndexar, err = sb.CopyInode0(diskPath, content.B_inodo, resultIndex)
				if err != nil {
					return -1, err
				}
				if inodoAIndexar == -1 {
					continue
				}
			} else {
				inodoAIndexar, err = sb.CopyInode1(diskPath, content.B_inodo)
				if err != nil {
					return -1, err
				}
				if inodoAIndexar == -1 {
					continue
				}
			}
			copy(folderBlock.B_content[indexContent].B_name[:], content.B_name[:])
			folderBlock.B_content[indexContent].B_inodo = inodoAIndexar
		}
		err = folderBlock.Serialize(diskPath, offsetFolderBlock)
		if err != nil {
			return -1, err
		}
	}
	err = inodoCopia.Serialize(diskPath, offsetInodoCopia)
	if err != nil {
		return -1, err
	}

	return resultIndex, nil
}

func (sb *SuperBlock) CopyInode1(diskPath string, indexInodeToCopy int32) (int32, error) {
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size*indexInodeToCopy))
	if err != nil {
		return -1, err
	}
	outcome, err := inode.HasPermissionsToRead(utils.LogedUserID, utils.LogedUserGroupID)
	if err != nil {
		return -1, err
	}
	if !outcome {
		return -1, nil
	}
	// Creamos el inodo
	resultIndex := sb.S_inodes_count
	inodoCopia := &Inode{
		I_uid:   inode.I_uid,
		I_gid:   inode.I_gid,
		I_size:  inode.I_size,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  inode.I_type,
		I_perm:  inode.I_perm,
	}
	offsetInodoCopia := int64(sb.S_first_ino)
	err = inodoCopia.Serialize(diskPath, offsetInodoCopia)
	if err != nil {
		return -1, err
	}
	err = sb.UpdateBitmapInode(diskPath)
	if err != nil {
		return -1, err
	}
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size
	//
	for i, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}
		inodoCopia.I_block[i] = sb.S_blocks_count
		contentBlock := &FileBlock{
			B_content: [64]byte{},
		}
		offsetFolderBlock := int64(sb.S_first_blo)
		err = contentBlock.Serialize(diskPath, offsetFolderBlock)
		if err != nil {
			return -1, err
		}
		err = sb.UpdateBitmapBlock(diskPath)
		if err != nil {
			return -1, err
		}
		sb.S_blocks_count++
		sb.S_free_blocks_count--
		sb.S_first_blo += sb.S_block_size
		//

		blockToGetInfo := &FileBlock{}
		err := blockToGetInfo.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
		if err != nil {
			return -1, err
		}
		copy(contentBlock.B_content[:], blockToGetInfo.B_content[:])
		err = contentBlock.Serialize(diskPath, offsetFolderBlock)
		if err != nil {
			return -1, err
		}
	}
	err = inodoCopia.Serialize(diskPath, offsetInodoCopia)
	if err != nil {
		return -1, err
	}

	return resultIndex, nil
}

func (sb *SuperBlock) RemoveInodo1(diskPath string, indexInode int32) (bool, error) { //devuelve true si se puede eliminar. Devuelve false si hay que preservar el tata
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size*indexInode))
	if err != nil {
		return false, err
	}
	outcome, err := inode.HasPermissionsToWrite(utils.LogedUserID, utils.LogedUserGroupID)
	if err != nil {
		return false, err
	}
	return outcome, nil
}

func (sb *SuperBlock) RemoveInodo0(diskPath string, indexInode int32) (bool, error) {

	resultRemoval := true
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size*indexInode))
	if err != nil {
		return false, err
	}
	outcome, err := inode.HasPermissionsToWrite(utils.LogedUserID, utils.LogedUserGroupID)
	if err != nil {
		return false, err
	}
	if !outcome {
		return false, nil
	}
	for i, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}
		block := &FolderBlock{}
		err := block.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
		if err != nil {
			return false, err
		}
		row := []bool{true, true}
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]
			if content.B_inodo == -1 {
				continue
			}
			tipoInodo, err := sb.TypeOfInode(diskPath, content.B_inodo)
			if err != nil {
				return false, err
			}
			if tipoInodo == 0 {
				row[indexContent-2], err = sb.RemoveInodo0(diskPath, content.B_inodo)

				if err != nil {
					return false, err
				}
			} else {
				row[indexContent-2], err = sb.RemoveInodo1(diskPath, content.B_inodo)
				if err != nil {
					return false, err
				}
			}
			if row[indexContent-2] {
				for j := range content.B_name {
					content.B_name[j] = 0
				}
				copy(content.B_name[:], []byte("-"))
				content.B_inodo = -1
				block.B_content[indexContent] = content
			}
		}
		err = block.Serialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return false, err
		}
		if row[0] && row[1] {
			inode.I_block[i] = -1
		} else if !row[0] || !row[1] {
			resultRemoval = false
		}
	}
	err = inode.Serialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size*indexInode))
	if err != nil {
		return false, err
	}

	return resultRemoval, nil
}

func (sb *SuperBlock) CommandFind(diskPath string, indexInode int32, level int, regex string) (string, error) {
	buffer := ""
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size*indexInode))
	if err != nil {
		return "", err
	}
	outcome, err := inode.HasPermissionsToRead(utils.LogedUserID, utils.LogedUserGroupID)
	if err != nil {
		return "", err
	}
	if !outcome {
		return "", nil
	}
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}
		block := &FolderBlock{}
		err := block.Deserialize(diskPath, int64(sb.S_block_start+sb.S_block_size*blockIndex))
		if err != nil {
			return "", err
		}
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]
			if content.B_inodo == -1 {
				continue
			}
			flag, err := sb.HasPermissionToCommandFind(diskPath, content.B_inodo)
			if err != nil {
				return "", err
			}

			if !flag {
				continue
			}
			tipoInodo, err := sb.TypeOfInode(diskPath, content.B_inodo)
			if err != nil {
				return "", err
			}
			if tipoInodo == 0 {
				// o que coincide o que viene algo
				contentName := strings.Trim(string(content.B_name[:]), "\x00")
				resultado, err := sb.CommandFind(diskPath, content.B_inodo, level+1, regex)
				if err != nil {
					return "", err
				}
				if resultado != "" {
					buffer += strings.Repeat("   ", level) + contentName + "\n" + resultado
					continue
				}
				re := regexp.MustCompile(regex)
				if re.MatchString(contentName) {
					buffer += strings.Repeat("   ", level) + contentName + "\n"
				}
			} else {
				contentName := strings.Trim(string(content.B_name[:]), "\x00")
				re := regexp.MustCompile(regex)
				if re.MatchString(contentName) {
					buffer += strings.Repeat("   ", level) + contentName + "\n"
				}
			}
		}
	}
	return buffer, nil
}

func (sb *SuperBlock) HasPermissionToCommandFind(diskPath string, indexInode int32) (bool, error) {
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+sb.S_inode_size*indexInode))
	if err != nil {
		return false, err
	}
	outcome, err := inode.HasPermissionsToRead(utils.LogedUserID, utils.LogedUserGroupID)
	if err != nil {
		return false, err
	}
	return outcome, nil
}
