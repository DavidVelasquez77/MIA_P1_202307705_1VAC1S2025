package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type EBR struct { //Pesa 30B
	Part_mount [1]byte
	Part_fit   [1]byte
	Part_start int32
	Part_size  int32
	Part_next  int32
	Part_name  [16]byte
}

func (ebr *EBR) SerializeEBR(path string, offset int32) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Con esto reubicamos el puntero en el archivo binario, lo reubicamos al offset pasado en parametro
	_, err = file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.LittleEndian, ebr)
	if err != nil {
		return err
	}
	return nil
}

func (ebr *EBR) DeserializeEBRAvailable(path string, offset int32) (int32, error) {
	file, err := os.Open(path)

	if err != nil {
		return -1, err
	}
	defer file.Close()

	// Con esto reubicamos el puntero en el archivo binario, lo reubicamos al offset pasado en parametro
	_, err = file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return -1, err
	}

	ebrSize := binary.Size(ebr)

	if ebrSize <= 0 {
		return -1, fmt.Errorf("invalid EBR size %d", ebrSize)
	}

	buffer := make([]byte, ebrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return -1, err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, ebr)
	if err != nil {
		return -1, err
	}

	if ebr.Part_next == -1 {
		return offset, nil
	}

	return ebr.DeserializeEBRAvailable(path, ebr.Part_next)
}

func (ebr *EBR) DeserializeEBR(path string, offset int32) error {

	file, err := os.Open(path)

	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(int64(offset), 0)
	if err != nil {
		return err
	}

	ebrSize := binary.Size(ebr)

	if ebrSize <= 0 {
		return fmt.Errorf("invalid EBR size %d", ebrSize)
	}

	buffer := make([]byte, ebrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, ebr)
	if err != nil {
		return err
	}

	return nil
}

func (ebr *EBR) SetNeoEBR(partFit string, partStart int, partSize int, partName string) {

	ebr.Part_mount[0] = '0'

	var fitByte byte
	switch partFit {
	case "FF":
		fitByte = 'F'
	case "BF":
		fitByte = 'B'
	case "WF":
		fitByte = 'W'
	}

	ebr.Part_fit[0] = fitByte

	ebr.Part_start = int32(partStart) + int32(binary.Size(ebr))

	ebr.Part_size = int32(partSize)

	ebr.Part_next = ebr.Part_start + int32(partSize)

	copy(ebr.Part_name[:], partName)
}

func (ebr *EBR) PrintEBR() {
	fmt.Printf("Extended Partition: \n")
	fmt.Printf(" Part_mount: %c\n", ebr.Part_mount[0])
	fmt.Printf(" Part_fit: %c\n", ebr.Part_fit[0])
	fmt.Printf(" Part_start: %d\n", ebr.Part_start)
	fmt.Printf(" Part_s: %d\n", ebr.Part_size)
	fmt.Printf(" Part_next: %d\n", ebr.Part_next)
	fmt.Printf(" Part_name: %d\n", ebr.Part_name[:])
}
