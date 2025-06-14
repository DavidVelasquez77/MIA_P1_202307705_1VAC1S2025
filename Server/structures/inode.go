package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Inode struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime float32
	I_ctime float32
	I_mtime float32
	I_block [15]int32
	I_type  [1]byte
	I_perm  [3]byte
}

func (inode *Inode) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}
	err = binary.Write(file, binary.LittleEndian, inode)
	if err != nil {
		return err
	}
	return nil
}

func (inode *Inode) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	inodeSize := binary.Size(inode)
	if inodeSize <= 0 {
		return fmt.Errorf("invalid Inode size: %d", inodeSize)
	}

	buffer := make([]byte, inodeSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, inode)
	if err != nil {
		return err
	}

	return nil
}

func (inode *Inode) Print() {
	atime := time.Unix(int64(inode.I_atime), 0)
	ctime := time.Unix(int64(inode.I_ctime), 0)
	mtime := time.Unix(int64(inode.I_mtime), 0)

	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_size)
	fmt.Printf("I_atime: %s\n", atime.Format(time.RFC3339))
	fmt.Printf("I_ctime: %s\n", ctime.Format(time.RFC3339))
	fmt.Printf("I_mtime: %s\n", mtime.Format(time.RFC3339))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Printf("I_type: %s\n", string(inode.I_type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.I_perm[:]))
}

func (inode *Inode) HasPermissionsToWrite(userID, groupID int32) (bool, error) {
	ownerUserId := inode.I_uid
	ownerGroupId := inode.I_gid
	permissions := string(inode.I_perm[:])

	if ownerUserId == userID || userID == 1 {
		permUser, err := strconv.Atoi(string(permissions[0]))
		if err != nil {
			return false, err
		}
		if permUser == 2 || permUser == 3 || permUser == 6 || permUser == 7 {
			return true, nil
		}
	}
	if ownerGroupId == groupID || userID == 1 {
		permGroup, err := strconv.Atoi(string(permissions[1]))
		if err != nil {
			return false, err
		}
		if permGroup == 2 || permGroup == 3 || permGroup == 6 || permGroup == 7 {
			return true, nil
		}
	}
	permOther, err := strconv.Atoi(string(permissions[2]))
	if err != nil {
		return false, err
	}
	if permOther == 2 || permOther == 3 || permOther == 6 || permOther == 7 {
		return true, nil
	}
	return false, nil
}

func (inode *Inode) HasPermissionsChmod(userID, groupID int32) (bool, error) {
	ownerUserId := inode.I_uid
	if ownerUserId == userID || userID == 1 {
		return true, nil
	}
	return false, nil
}
func (inode *Inode) HasPermissionsToRead(userID, groupID int32) (bool, error) {
	ownerUserId := inode.I_uid
	ownerGroupId := inode.I_gid
	permissions := string(inode.I_perm[:])

	if ownerUserId == userID || userID == 1 {
		permUser, err := strconv.Atoi(string(permissions[0]))
		if err != nil {
			return false, err
		}
		if permUser >= 4 {
			return true, nil
		}
	}
	if ownerGroupId == groupID || userID == 1 {
		permGroup, err := strconv.Atoi(string(permissions[1]))
		if err != nil {
			return false, err
		}
		if permGroup >= 4 {
			return true, nil
		}
	}
	permOther, err := strconv.Atoi(string(permissions[2]))
	if err != nil {
		return false, err
	}
	if permOther >= 4 {
		return true, nil
	}
	return false, nil
}

/*
read: 4
write: 2
execute: 1
D r w x
0 0 0 0
1 0 0 1
2 0 1 0
3 0 1 1
4 1 0 0
5 1 0 1
6 1 1 0
7 1 1 1
*/
