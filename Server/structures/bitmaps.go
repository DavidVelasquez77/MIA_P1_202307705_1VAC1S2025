package structures

import (
	"encoding/binary"
	"os"
)

func (sb *SuperBlock)CreateBitMaps(path string) error {
	file, err:= os.OpenFile(path, os.O_WRONLY| os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(int64(sb.S_bm_inode_start), 0)
	if err != nil {
		return err
	}
	
	buffer := make([]byte, sb.S_free_inodes_count)
	for i := range buffer{
		buffer[i] = '0'
	}
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	_, err = file.Seek(int64(sb.S_bm_block_start), 0)
	if err != nil {
		return err
	}
	buffer = make([]byte, sb.S_free_blocks_count)
	for i := range buffer {
		buffer[i] = 'O'
	}
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	return nil

}

func (sb *SuperBlock)UpdateBitmapInode(path string) error  {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(int64(sb.S_bm_inode_start)+int64(sb.S_inodes_count), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{'1'})
	if err != nil {
		return err
	}
	return nil
}

func (sb *SuperBlock) UpdateBitmapBlock(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(int64(sb.S_bm_block_start)+int64(sb.S_blocks_count), 0)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte{'X'})
	if err != nil {
		return err
	}

	return nil
}