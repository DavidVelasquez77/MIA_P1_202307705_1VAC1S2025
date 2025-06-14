package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type FileBlock struct {
	B_content [64]byte
}

func (fb *FileBlock) Serialize(path string, offset int64) error {
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

func (fb *FileBlock) Deserialize(path string, offset int64) error {
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
		return fmt.Errorf("invalid FileBlock size: %d", fbSize)
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

func (fb *FileBlock) Print() {
	fmt.Println("ESTE ES EL CONTEIDO DEL BLOQUE:", string(fb.B_content[:]))
}
