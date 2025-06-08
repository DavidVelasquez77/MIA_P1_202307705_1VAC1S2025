package structures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type MBR struct {
	Mbr_size           int32   //4Bytes
	Mbr_creation_date  float32 //4Bytes
	Mbr_disk_signature int32   //4Bytes
	Mbr_disk_fit       [1]byte //4Bytes
	Mbr_partitions     [4]PARTITION
}

func (mbr *MBR) SerializeMBR(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}
	return nil

}

func (mbr *MBR) DeserializeMBR(path string) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}
	defer file.Close()

	mbrsize := binary.Size(mbr)

	if mbrsize <= 0 {
		return fmt.Errorf("invalid MBR size %d", mbrsize)
	}

	buffer := make([]byte, mbrsize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

func (mbr *MBR) GetFirstAvailablePartition() (*PARTITION, int, int) {
	offset := binary.Size(mbr)

	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		if mbr.Mbr_partitions[i].Part_start == -1 {
			return &mbr.Mbr_partitions[i], offset, i
		} else {
			offset += int(mbr.Mbr_partitions[i].Part_size)
		}
	}
	return nil, -1, -1
}

func (mbr *MBR) GetPartitionByName(name string) (*PARTITION, int) {
	for i, partition := range mbr.Mbr_partitions {
		partitionName := strings.Trim(string(partition.Part_name[:]), "\x00 ")
		inputName := strings.Trim(name, "\x00")
		if strings.EqualFold(partitionName, inputName) {
			return &partition, i
		}
	}

	return nil, -1
}

func (mbr *MBR) IsThereExtendedPartition() bool {
	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		if rune(mbr.Mbr_partitions[i].Part_type[0]) == 'E' {
			return true
		}
	}
	return false
}

func (mbr *MBR) GetOffsetFirstEBR() (int32, int32, error) {
	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		if rune(mbr.Mbr_partitions[i].Part_type[0]) == 'E' {
			return mbr.Mbr_partitions[i].Part_start, mbr.Mbr_partitions[i].Part_size, nil
		}
	}
	return -1, -1, errors.New("no hay una particion extendida")
}

func (mbr *MBR) PrintMBR() {
	creationTime := time.Unix(int64(mbr.Mbr_creation_date), 0)

	diskFit := rune(mbr.Mbr_disk_fit[0])

	fmt.Printf("MBR Size: %d\n", mbr.Mbr_size)
	fmt.Printf("Creation Date: %s\n", creationTime.Format(time.RFC3339))
	fmt.Printf("Disk Signature: %d\n", mbr.Mbr_disk_signature)
	fmt.Printf("Disk Fit: %c\n", diskFit)
}

func (mbr *MBR) PrintPartitions() {
	for i, partition := range mbr.Mbr_partitions {
		partStatus := rune(partition.Part_status[0])
		partType := rune(partition.Part_type[0])
		partFit := rune(partition.Part_fit[0])

		partName := string(partition.Part_name[:])
		partID := string(partition.Part_id[:])

		fmt.Printf("Partition %d:\n", i+1)
		fmt.Printf("  Status: %c\n", partStatus)
		fmt.Printf("  Type: %c\n", partType)
		fmt.Printf("  Fit: %c\n", partFit)
		fmt.Printf("  Start: %d\n", partition.Part_start)
		fmt.Printf("  Size: %d\n", partition.Part_size)
		fmt.Printf("  Name: %s\n", partName)
		fmt.Printf("  Correlative: %d\n", partition.Part_correlative)
		fmt.Printf("  ID: %s\n", partID)
	}
}

func (mbr *MBR) CanFitAnotherDisk(sizeBytes int) bool {
	bytesUsed := binary.Size(mbr)
	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		if mbr.Mbr_partitions[i].Part_size != -1 {
			bytesUsed += int(mbr.Mbr_partitions[i].Part_size)
		}
	}
	return bytesUsed+sizeBytes <= int(mbr.Mbr_size)

}

func (mbr *MBR) GetPartitionByID(id string) (*PARTITION, error) {
	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		partitionID := strings.Trim(string(mbr.Mbr_partitions[i].Part_id[:]), "\x00 ")
		inputID := strings.Trim(id, "\x00 ")
		if strings.EqualFold(partitionID, inputID) {
			return &mbr.Mbr_partitions[i], nil
		}
	}
	return nil, errors.New("partición no encontrada")
}