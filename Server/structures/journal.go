package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

type Journal struct {
	J_next    int32
	J_content Information
}

type Information struct {
	I_operation [10]byte
	I_path      [74]byte
	I_content   [64]byte
	I_date      float32
}

func (journal *Journal) Serialize(path string, offset int64) error {
	// offset := journaling_start + (int64(binary.Size(Journal{})))*int64(journal.J_next)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.LittleEndian, journal)
	if err != nil {
		return err
	}
	return nil
}

func (journal *Journal) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}
	journalSize := binary.Size(Journal{})

	buffer := make([]byte, journalSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, journal)
	if err != nil {
		return err
	}

	return nil
}

func (journal *Journal) Print() {
	// Convertir el tiempo de montaje a una fecha
	date := time.Unix(int64(journal.J_content.I_date), 0)

	fmt.Println("Journal:")
	fmt.Printf("J_count: %d", journal.J_next)
	fmt.Println("Information:")
	fmt.Printf("I_operation: %s", strings.TrimRight(string(journal.J_content.I_operation[:]), "\x00"))
	fmt.Printf("I_path: %s", string(journal.J_content.I_path[:]))
	fmt.Printf("I_content: %s", string(journal.J_content.I_content[:]))
	fmt.Printf("I_date: %s", date.Format(time.RFC3339))
}
