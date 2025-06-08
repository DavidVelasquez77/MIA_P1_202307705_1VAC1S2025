package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type FolderBlock struct {
	B_content [4]FolderContent
}

type FolderContent struct {
	B_name [12]byte
	B_inodo int32
}

func (fb *FolderBlock)Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.LittleEndian, fb)
	if err != nil {
		return err
	}

	return nil
}

func (fb *FolderBlock) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	fbSize := binary.Size(fb)
	if fbSize <= 0 {
		return fmt.Errorf("invalid FolderBlock size: %d", fbSize)
	}

	buffer := make([]byte, fbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, fb)
	if err != nil {
		return err
	}

	return nil
}

func (fb *FolderBlock)Print()  {
	for i, content := range fb.B_content {
		name := string(content.B_name[:])
		fmt.Printf("Content %d:\n", i+1)
		fmt.Printf("  B_name: %s\n", name)
		fmt.Printf("  B_inodo: %d\n", content.B_inodo)
	}
}
