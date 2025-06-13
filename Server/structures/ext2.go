package structures

import (
	// structures "server/structures"

	"errors"
	utils "server/utils"
	"strings"
	"time"
)

func (sb *SuperBlock) createFolderInInode(path string, inodeIndex int32, parentsDir []string, destDir string, justSearchingAFile bool) error {
	inode := &Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return err
	}
	if inode.I_type[0] == '1' {
		return nil
	}
	var inodoPadre int32
	for i, blockIndex := range inode.I_block {
		numeroApuntadorIndirect := sb.S_blocks_count
		if blockIndex == -1 {
			if !justSearchingAFile && len(parentsDir) != 0 {
				return errors.New("ruta invalida, asegurese que exita la ruta antes")
			}
			// Aqui se debe validar si es la iteracion 15 en adelante para hacer lo de los apuntadores indirectos
			if i >= 15 {
				inode.I_block[i] = sb.S_blocks_count

				pointerBlock := &PointerBlock{
					P_pointers: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
				}
				err = pointerBlock.Serialize(path, int64(sb.S_first_blo))
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

				//Bloque folder
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodoPadre},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}
				err = folderBlock.Serialize(path, int64(sb.S_first_blo))
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

				inode.Serialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
				flag, err := sb.folderFromAuntadorIndirecto13(path, inodeIndex, parentsDir, destDir, justSearchingAFile, numeroApuntadorIndirect, inodoPadre)
				if err != nil {
					return err
				}
				if flag {
					return nil
				}
			} else {
				inode.I_block[i] = sb.S_blocks_count
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodoPadre},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}
				err = folderBlock.Serialize(path, int64(sb.S_first_blo))
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
				inode.Serialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
				return sb.createFolderInInode(path, inodeIndex, parentsDir, destDir, justSearchingAFile)
			}
		}

		if i >= 15 {
			flag, err := sb.folderFromAuntadorIndirecto13(path, inodeIndex, parentsDir, destDir, justSearchingAFile, blockIndex, inodoPadre)
			if err != nil {
				return err
			}
			if flag {
				return nil
			}

		}
		block := &FolderBlock{}
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return err
		}

		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]

			if len(parentsDir) != 0 {

				if content.B_inodo == -1 {
					break
				}

				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")
				if strings.EqualFold(contentName, parentDirName) {
					err := sb.createFolderInInode(path, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir, justSearchingAFile)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				contentName := strings.Trim(string(content.B_name[:]), "\x00")
				destinationName := strings.Trim(destDir, "\x00")
				if strings.EqualFold(contentName, destinationName) {
					return errors.New("ya existe un directorio con el mismo nombre")
				}
				if content.B_inodo != -1 {
					tempContent := block.B_content[1]
					inodoPadre = tempContent.B_inodo
					continue
				}

				copy(content.B_name[:], destDir)
				content.B_inodo = sb.S_inodes_count

				block.B_content[indexContent] = content

				err = block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}

				folderInode := &Inode{
					I_uid:   utils.LogedUserID,
					I_gid:   utils.LogedUserGroupID,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				err = folderInode.Serialize(path, int64(sb.S_first_ino))
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

				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: content.B_inodo},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				err = folderBlock.Serialize(path, int64(sb.S_first_blo))
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

				return nil
			}
		}

	}
	return nil
}

func (sb *SuperBlock) createFolderInInodeWithP(path string, inodeIndex int32, parentsDir []string, destDir string) error {
	inodo := &Inode{}
	err := inodo.Deserialize(path, int64(sb.S_inode_start+inodeIndex*sb.S_inode_size))
	if err != nil {
		return err
	}
	if len(parentsDir) == 0 {
		sb.createFolderInInode(path, inodeIndex, make([]string, 0), destDir, true)
		return nil
	}
	nameDir, err := utils.First(parentsDir)
	if err != nil {
		return err
	}

	flag, neoInodoToVisit, err := existTheDirectory(inodo, nameDir, path, sb)
	if err != nil {
		return err
	}
	if flag { //si existe el primer dir
		sb.createFolderInInodeWithP(path, neoInodoToVisit, utils.RemoveElement(parentsDir, 0), destDir)
	} else { //No existe el primero dir
		sb.createFolderInInode(path, inodeIndex, make([]string, 0), nameDir, true)
		sb.createFolderInInodeWithP(path, inodeIndex, parentsDir, destDir)
	}
	return nil
}

// Asumiendo que hayan dirs q crear
func existTheDirectory(inodo *Inode, nameDir string, path string, sb *SuperBlock) (bool, int32, error) {
	for i, blockIndex := range inodo.I_block {
		if blockIndex == -1 {
			return false, 0, nil
		}
		if i >= 15 {
			pointerBlock := &PointerBlock{}
			err := pointerBlock.Deserialize(path, int64(sb.S_block_start+(sb.S_block_size*blockIndex)))
			if err != nil {
				return false, 0, err
			}
			for _, value := range pointerBlock.P_pointers {
				if value == -1 {
					break
				}
				block := &FolderBlock{}
				err := block.Deserialize(path, int64(sb.S_block_start+(sb.S_block_size*value)))
				if err != nil {
					return false, 0, err
				}
				for i := 2; i < len(block.B_content); i++ {
					content := block.B_content[i]
					contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
					parentDirName := strings.Trim(nameDir, "\x00 ")
					if strings.EqualFold(contentName, parentDirName) {
						return true, content.B_inodo, nil
					}
				}
			}

		} else {
			block := &FolderBlock{}
			err := block.Deserialize(path, int64(sb.S_block_start+(sb.S_block_size*blockIndex)))
			if err != nil {
				return false, 0, err
			}
			for i := 2; i < len(block.B_content); i++ {
				content := block.B_content[i]
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(nameDir, "\x00 ")
				if strings.EqualFold(contentName, parentDirName) {
					return true, content.B_inodo, nil
				}
			}
		}
	}
	return false, 0, nil
}

func (sb *SuperBlock) CreateFile(diskPath string, inodeIndex int32, parentsDir []string, destDir string, fileContent string, size int32, justSearchingAFile bool) error {
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return err
	}
	if inode.I_type[0] == '1' {
		return nil
	}
	var inodoPadre int32
	for i, blockIndex := range inode.I_block {
		numeroApuntadorIndirect := sb.S_blocks_count
		if blockIndex == -1 {
			if !justSearchingAFile && len(parentsDir) != 0 {
				return errors.New("ruta invalida, asegurese que exita la ruta antes")
			}
			// Aqui se debe validar si es la iteracion 15 en adelante para hacer lo de los apuntadores indirectos
			if i >= 15 {
				inode.I_block[i] = sb.S_blocks_count

				pointerBlock := &PointerBlock{
					P_pointers: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
				}
				err = pointerBlock.Serialize(diskPath, int64(sb.S_first_blo))
				if err != nil {
					return err
				}

				err = sb.UpdateBitmapBlock(diskPath)
				if err != nil {
					return err
				}
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				//Bloque folder
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodoPadre},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}
				err = folderBlock.Serialize(diskPath, int64(sb.S_first_blo))
				if err != nil {
					return err
				}

				err = sb.UpdateBitmapBlock(diskPath)
				if err != nil {
					return err
				}
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				inode.Serialize(diskPath, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
				flag, err := sb.folderFromAuntadorIndirecto13(diskPath, inodeIndex, parentsDir, destDir, justSearchingAFile, numeroApuntadorIndirect, inodoPadre)
				if err != nil {
					return err
				}
				if flag {
					return nil
				}
			} else {
				inode.I_block[i] = sb.S_blocks_count
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodoPadre},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}
				err = folderBlock.Serialize(diskPath, int64(sb.S_first_blo))
				if err != nil {
					return err
				}

				err = sb.UpdateBitmapBlock(diskPath)
				if err != nil {
					return err
				}
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size
				inode.Serialize(diskPath, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
				return sb.CreateFile(diskPath, inodeIndex, parentsDir, destDir, fileContent, size, justSearchingAFile)
			}
		}

		if i >= 15 {
			pointerBlock := &PointerBlock{}
			err := pointerBlock.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return err
			}
			for _, neoBlockIndex := range pointerBlock.P_pointers {
				if neoBlockIndex == -1 {
					break
				}
				block := &FolderBlock{}
				err := block.Deserialize(diskPath, int64(sb.S_block_start+(neoBlockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}
				for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
					content := block.B_content[indexContent]
					if len(parentsDir) != 0 {
						if content.B_inodo == -1 {
							break
						}
						parentDir, err := utils.First(parentsDir)
						if err != nil {
							return err
						}
						contentName := strings.Trim(string(content.B_name[:]), "\x00")
						parentDirName := strings.Trim(parentDir, "\x00")
						if strings.EqualFold(contentName, parentDirName) {
							err := sb.CreateFile(diskPath, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir, fileContent, size, justSearchingAFile)
							if err != nil {
								return err
							}
							return nil
						}
					} else {
						contentName := strings.Trim(string(content.B_name[:]), "\x00")
						destinationName := strings.Trim(destDir, "\x00")
						if strings.EqualFold(contentName, destinationName) {
							return errors.New("ya existe un file con el mismo nombre")
						}
						if content.B_inodo != -1 {
							tempContent := block.B_content[1]
							inodoPadre = tempContent.B_inodo
							continue
						}
						copy(content.B_name[:], destDir)
						content.B_inodo = sb.S_inodes_count
						block.B_content[indexContent] = content
						err = block.Serialize(diskPath, int64(sb.S_block_start+(neoBlockIndex*sb.S_block_size)))
						if err != nil {
							return err
						}
						folderInode := &Inode{
							I_uid:   utils.LogedUserID,
							I_gid:   utils.LogedUserGroupID,
							I_size:  int32(len(fileContent)),
							I_atime: float32(time.Now().Unix()),
							I_ctime: float32(time.Now().Unix()),
							I_mtime: float32(time.Now().Unix()),
							I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
							I_type:  [1]byte{'1'},
							I_perm:  [3]byte{'6', '6', '4'},
						}
						contentChunks := utils.SplitStringIntoChunks(fileContent)
						// for i := range contentChunks {
						// 	folderInode.I_block[i] = sb.S_blocks_count + int32(i)

						// }
						offsetInodo := sb.S_first_ino
						err = folderInode.Serialize(diskPath, int64(offsetInodo))
						if err != nil {
							return err
						}
						err = sb.UpdateBitmapInode(diskPath)
						if err != nil {
							return err
						}

						sb.S_inodes_count++
						sb.S_free_inodes_count--
						sb.S_first_ino += sb.S_inode_size
						// Flag de repetir el llenado
						flagToCreateNeoBlockPointer := true
						tempCont := 15
						// Index del bloque pointer
						var indexBlockPointer int32
						tempBlockPointer := &PointerBlock{}
						for i, content := range contentChunks {
							if i >= 15 { //Aputnadores indirectos
								if flagToCreateNeoBlockPointer {
									for id, value := range folderInode.I_block {
										if value == -1 {
											folderInode.I_block[id] = sb.S_blocks_count
											break
										}
									}
									indexBlockPointer = folderInode.I_block[tempCont]
									pointerBlock := &PointerBlock{
										P_pointers: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
									}
									tempBlockPointer = pointerBlock
									flagToCreateNeoBlockPointer = false
									err = sb.UpdateBitmapBlock(diskPath)
									if err != nil {
										return err
									}
									sb.S_blocks_count++
									sb.S_free_blocks_count--
									sb.S_first_blo += sb.S_block_size
								}
								flag, err := stuffTheBlock(sb, tempBlockPointer, content, diskPath)
								if err != nil {
									return err
								}
								err = tempBlockPointer.Serialize(diskPath, int64(sb.S_block_start+(indexBlockPointer*sb.S_block_size)))
								if err != nil {
									return err
								}
								if !flag {
									tempCont++
									flagToCreateNeoBlockPointer = true
									indexBlockPointer = folderInode.I_block[tempCont]
								}
							} else {
								folderInode.I_block[i] = sb.S_blocks_count
								contentBlock := &FileBlock{
									B_content: [64]byte{},
								}
								copy(contentBlock.B_content[:], content)
								err = contentBlock.Serialize(diskPath, int64(sb.S_first_blo))
								if err != nil {
									return err
								}

								err = sb.UpdateBitmapBlock(diskPath)
								if err != nil {
									return err
								}
								sb.S_blocks_count++
								sb.S_free_blocks_count--
								sb.S_first_blo += sb.S_block_size
							}
						}
						err = folderInode.Serialize(diskPath, int64(offsetInodo))
						if err != nil {
							return err
						}
						return nil
					}
				}
			}

		}
		block := &FolderBlock{}
		err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return err
		}
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]
			if len(parentsDir) != 0 {
				if content.B_inodo == -1 {
					break
				}
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}
				contentName := strings.Trim(string(content.B_name[:]), "\x00")
				parentDirName := strings.Trim(parentDir, "\x00")
				if strings.EqualFold(contentName, parentDirName) {
					err := sb.CreateFile(diskPath, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir, fileContent, size, justSearchingAFile)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				contentName := strings.Trim(string(content.B_name[:]), "\x00")
				destinationName := strings.Trim(destDir, "\x00")
				if strings.EqualFold(contentName, destinationName) {
					return errors.New("ya existe un file con el mismo nombre")
				}
				if content.B_inodo != -1 {
					tempContent := block.B_content[1]
					inodoPadre = tempContent.B_inodo
					continue
				}
				copy(content.B_name[:], destDir)
				content.B_inodo = sb.S_inodes_count
				block.B_content[indexContent] = content
				err = block.Serialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}
				folderInode := &Inode{
					I_uid:   utils.LogedUserID,
					I_gid:   utils.LogedUserGroupID,
					I_size:  int32(len(fileContent)),
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'1'},
					I_perm:  [3]byte{'6', '6', '4'},
				}
				contentChunks := utils.SplitStringIntoChunks(fileContent)
				// for i := range contentChunks {
				// 	folderInode.I_block[i] = sb.S_blocks_count + int32(i)

				// }
				offsetInodo := sb.S_first_ino
				err = folderInode.Serialize(diskPath, int64(offsetInodo))
				if err != nil {
					return err
				}
				err = sb.UpdateBitmapInode(diskPath)
				if err != nil {
					return err
				}

				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size
				// Flag de repetir el llenado
				flagToCreateNeoBlockPointer := true
				tempCont := 15
				// Index del bloque pointer
				var indexBlockPointer int32
				tempBlockPointer := &PointerBlock{}
				for i, content := range contentChunks {
					if i >= 15 { //Aputnadores indirectos
						if flagToCreateNeoBlockPointer {
							for id, value := range folderInode.I_block {
								if value == -1 {
									folderInode.I_block[id] = sb.S_blocks_count
									break
								}
							}
							indexBlockPointer = folderInode.I_block[tempCont]
							pointerBlock := &PointerBlock{
								P_pointers: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
							}
							tempBlockPointer = pointerBlock
							flagToCreateNeoBlockPointer = false
							err = sb.UpdateBitmapBlock(diskPath)
							if err != nil {
								return err
							}
							sb.S_blocks_count++
							sb.S_free_blocks_count--
							sb.S_first_blo += sb.S_block_size
						}
						flag, err := stuffTheBlock(sb, tempBlockPointer, content, diskPath)
						if err != nil {
							return err
						}
						err = tempBlockPointer.Serialize(diskPath, int64(sb.S_block_start+(indexBlockPointer*sb.S_block_size)))
						if err != nil {
							return err
						}
						if !flag {
							tempCont++
							flagToCreateNeoBlockPointer = true
							indexBlockPointer = folderInode.I_block[tempCont]
						}
					} else {
						folderInode.I_block[i] = sb.S_blocks_count
						contentBlock := &FileBlock{
							B_content: [64]byte{},
						}
						copy(contentBlock.B_content[:], content)
						err = contentBlock.Serialize(diskPath, int64(sb.S_first_blo))
						if err != nil {
							return err
						}

						err = sb.UpdateBitmapBlock(diskPath)
						if err != nil {
							return err
						}
						sb.S_blocks_count++
						sb.S_free_blocks_count--
						sb.S_first_blo += sb.S_block_size
					}
				}
				err = folderInode.Serialize(diskPath, int64(offsetInodo))
				if err != nil {
					return err
				}
				return nil
			}
		}
	}
	return nil
}

func stuffTheBlock(sb *SuperBlock, pointerBlock *PointerBlock, content string, diskPath string) (bool, error) {
	for i, indexInode := range pointerBlock.P_pointers {

		if indexInode != -1 {
			continue
		}
		pointerBlock.P_pointers[i] = sb.S_blocks_count
		contentBlock := &FileBlock{
			B_content: [64]byte{},
		}
		copy(contentBlock.B_content[:], content)
		err := contentBlock.Serialize(diskPath, int64(sb.S_first_blo))
		if err != nil {
			return false, err
		}

		err = sb.UpdateBitmapBlock(diskPath)
		if err != nil {
			return false, err
		}
		sb.S_blocks_count++
		sb.S_free_blocks_count--
		sb.S_first_blo += sb.S_block_size
		if i == len(pointerBlock.P_pointers)-1 {
			return false, nil
		} else {
			return true, nil
		}
	}
	return false, errors.New("no entro al forr de llenado por apuntadores indirectos")
}

func (sb *SuperBlock) ContentFromFile(diskPath string, inodeIndex int32, parentsDir []string, destDir string) (string, error) {
	inode := &Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return "", err
	}
	if inode.I_type[0] == '1' {
		return "", errors.New("se entro a un Inodo tipo file en reportFile")
	}
	for _, blockIndex := range inode.I_block {

		if blockIndex == -1 {
			return "", errors.New("error en el path solicitado para extraer informacion de un archivo")
		}
		block := &FolderBlock{}
		err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return "", err
		}
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]
			if len(parentsDir) != 0 {
				if content.B_inodo == -1 {
					break
				}
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return "", err
				}
				contentName := strings.Trim(string(content.B_name[:]), "\x00")
				parentDirName := strings.Trim(parentDir, "\x00")
				if strings.EqualFold(contentName, parentDirName) {
					content, err := sb.ContentFromFile(diskPath, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return "", err
					}
					return content, nil
				}
			} else {
				contentName := strings.Trim(string(content.B_name[:]), "\x00")
				destinationName := strings.Trim(destDir, "\x00")
				if strings.EqualFold(contentName, destinationName) {
					// Son iguales
					inodoFile := &Inode{}
					inodoFile.Deserialize(diskPath, int64(sb.S_inode_start+(content.B_inodo*sb.S_inode_size)))
					var content string
					for iTe, value := range inodoFile.I_block {
						if value == -1 {
							break
						}
						if iTe >= 15 {
							pointerBlock := &PointerBlock{}
							err := pointerBlock.Deserialize(diskPath, int64(sb.S_block_start+(sb.S_block_size*value)))
							if err != nil {
								return "", err
							}
							for _, indexContentBlock := range pointerBlock.P_pointers {
								if indexContentBlock == -1 {
									break
								}
								blockContentFile := &FileBlock{}
								blockContentFile.Deserialize(diskPath, int64(sb.S_block_start+(sb.S_block_size*indexContentBlock)))
								contentBlock := string(blockContentFile.B_content[:])
								contentBlock = strings.TrimRight(contentBlock, "\x00")
								content += contentBlock
							}
						} else {
							blockContentFile := &FileBlock{}
							blockContentFile.Deserialize(diskPath, int64(sb.S_block_start+(sb.S_block_size*value)))
							contentBlock := string(blockContentFile.B_content[:])
							contentBlock = strings.TrimRight(contentBlock, "\x00")
							content += contentBlock
						}
					}
					return content, nil
				}
				if content.B_inodo != -1 {
					continue
				}
			}
		}
	}
	return "", errors.New("se ha producido un error en reportFile")
}

func (sb *SuperBlock) folderFromAuntadorIndirecto13(diskPath string, inodeIndex int32, parentsDir []string, destDir string, justSearchingAFile bool, numApuntadorIndirecto int32, inodoPadre int32) (bool, error) {
	pointerBlock := &PointerBlock{}
	err := pointerBlock.Deserialize(diskPath, int64(sb.S_block_start+(sb.S_block_size*numApuntadorIndirecto)))
	if err != nil {
		return false, err
	}

	for i, blockIndex := range pointerBlock.P_pointers {
		if blockIndex == -1 {
			// Aqui se debe validar si es la iteracion 15 en adelante para hacer lo de los apuntadores indirectos
			if !justSearchingAFile && len(parentsDir) != 0 {
				return false, errors.New("ruta invalida, asegurese que exita la ruta antes")
			}

			pointerBlock.P_pointers[i] = sb.S_blocks_count
			folderBlock := &FolderBlock{
				B_content: [4]FolderContent{
					{B_name: [12]byte{'.'}, B_inodo: inodeIndex},
					{B_name: [12]byte{'.', '.'}, B_inodo: inodoPadre},
					{B_name: [12]byte{'-'}, B_inodo: -1},
					{B_name: [12]byte{'-'}, B_inodo: -1},
				},
			}
			err = folderBlock.Serialize(diskPath, int64(sb.S_first_blo))
			if err != nil {
				return false, err
			}

			err = sb.UpdateBitmapBlock(diskPath)
			if err != nil {
				return false, err
			}
			sb.S_blocks_count++
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size

			err = pointerBlock.Serialize(diskPath, int64(sb.S_block_start+(sb.S_block_size*numApuntadorIndirecto)))
			if err != nil {
				return false, err
			}

			return sb.folderFromAuntadorIndirecto13(diskPath, inodeIndex, parentsDir, destDir, justSearchingAFile, numApuntadorIndirecto, inodoPadre)
		}

		// Si es la iteracion 15,
		block := &FolderBlock{}
		err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return false, err
		}

		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]

			if len(parentsDir) != 0 {

				if content.B_inodo == -1 {
					break
				}

				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return false, err
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")
				if strings.EqualFold(contentName, parentDirName) {
					err := sb.createFolderInInode(diskPath, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir, justSearchingAFile)
					if err != nil {
						return false, err
					}
					return true, nil
				}
			} else {

				contentName := strings.Trim(string(content.B_name[:]), "\x00")
				destinationName := strings.Trim(destDir, "\x00")
				if strings.EqualFold(contentName, destinationName) {
					return false, errors.New("ya existe un directorio con el mismo nombre")
				}

				if content.B_inodo != -1 {
					continue
				}

				copy(content.B_name[:], destDir)
				content.B_inodo = sb.S_inodes_count

				block.B_content[indexContent] = content

				err = block.Serialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return false, err
				}

				folderInode := &Inode{
					I_uid:   utils.LogedUserID,
					I_gid:   utils.LogedUserGroupID,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				err = folderInode.Serialize(diskPath, int64(sb.S_first_ino))
				if err != nil {
					return false, err
				}

				err = sb.UpdateBitmapInode(diskPath)
				if err != nil {
					return false, err
				}

				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: content.B_inodo},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				err = folderBlock.Serialize(diskPath, int64(sb.S_first_blo))
				if err != nil {
					return false, err
				}

				err = sb.UpdateBitmapBlock(diskPath)
				if err != nil {
					return false, err
				}
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				err = pointerBlock.Serialize(diskPath, int64(sb.S_inode_start+(numApuntadorIndirecto*sb.S_inode_size)))
				if err != nil {
					return false, err
				}
				return true, nil
			}
		}
	}
	return false, nil
}
