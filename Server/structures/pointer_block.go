package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type PointerBlock struct {
	P_pointers [16]int32
}

func (pb *PointerBlock) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.LittleEndian, pb)
	if err != nil {
		return err
	}

	return nil
}

func (pb *PointerBlock) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	pbSize := binary.Size(pb)
	if pbSize <= 0 {
		return fmt.Errorf("invalid PointerBlock size: %d", pbSize)
	}

	buffer := make([]byte, pbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, pb)
	if err != nil {
		return err
	}

	return nil
}

func (pb *PointerBlock) Print() {
	fmt.Println("----------------------")
	fmt.Println("CONTENIDO POINTERBLOCK")

	for i, value := range pb.P_pointers {
		fmt.Printf("VALOR %d: %d\n", i, value)
	}
}
