package commands



import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"server/structures"
	"server/utils"
	"strconv"
	"strings"
)

type FDISK struct {
	size int
	unit string
	fit  string
	path string
	typ  string
	name string
}

func ParseFdisk(tokens []string) (string, error) {
	cmd := &FDISK{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmMbB]|-fit=[bBfF]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parametro invalid: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", err
			}
			cmd.size = size
		case "-unit":
			value = strings.ToUpper(value)
			if value != "K" && value != "M" && value != "B" {
				return "", errors.New("la unidad debe ser K o M o B")
			}
			cmd.unit = strings.ToUpper(value)
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return "", errors.New("el tipo debe ser P, E o L")
			}
			cmd.typ = value
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.size == 0 {
		return "", errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	if cmd.unit == "" {
		cmd.unit = "M"
	}

	if cmd.fit == "" {
		cmd.fit = "WF"
	}

	if cmd.typ == "" {
		cmd.typ = "P"
	}

	err := commandFdisk(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("FDISK: %s creado exitosamente", cmd.name), nil
}

func commandFdisk(fdisk *FDISK) error {
	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		return err
	}

	if fdisk.typ == "P" {
		err = createPrimaryPartition(fdisk, sizeBytes)
		if err != nil {
			return err
		}

	} else if fdisk.typ == "E" {
		err = createExtendedPartittion(fdisk, sizeBytes)
		if err != nil {
			return err
		}
	} else if fdisk.typ == "L" {
		err = createLogicPartition(fdisk, sizeBytes)
		if err != nil {
			return err
		}
	}

	return nil

}

func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil {
		return err
	}

	if !mbr.CanFitAnotherDisk(sizeBytes) {
		return errors.New("no se puede crear una particion por falta de espacio")
	}

	// fmt.Println("\nMBR original: ")
	// mbr.PrintMBR()

	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return errors.New("no hay partitciones disponibles")
	}

	// fmt.Println("\nParticion disponible:")
	// availablePartition.PrintPartition()

	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	// fmt.Println("\nParticion creada (modificada):")
	// availablePartition.PrintPartition()

	mbr.Mbr_partitions[indexPartition] = *availablePartition

	// fmt.Println("\nParticiones del MBR:")
	// mbr.PrintPartitions()

	err = mbr.SerializeMBR(fdisk.path)
	if err != nil {
		return err
	}
	return nil

}

func createExtendedPartittion(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil {
		return err
	}

	if !mbr.CanFitAnotherDisk(sizeBytes) {
		return errors.New("no se puede crear una particion por falta de espacio")
	}

	// fmt.Println("\nMBR original: ")
	// mbr.PrintMBR()

	if mbr.IsThereExtendedPartition() {
		return errors.New("no se puede crear mas de 1 particion extendida por disco")
	}

	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return errors.New("no hay partitciones disponibles")
	}

	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	// fmt.Println("\nParticion creada (modificada):")
	// availablePartition.PrintPartition()

	mbr.Mbr_partitions[indexPartition] = *availablePartition

	// fmt.Println("\n Particiones del MBR(actualizado): ")
	// mbr.PrintPartitions()

	err = mbr.SerializeMBR(fdisk.path)
	if err != nil {
		return err
	}

	// Creamos el EBR
	err = createEBR(fdisk, startPartition)
	if err != nil {
		return err
	}

	return nil

}

func createEBR(fdisk *FDISK, offset int) error {

	ebr := &structures.EBR{
		Part_mount: [1]byte{'N'},
		Part_fit:   [1]byte{'N'},
		Part_start: -1,
		Part_size:  -1,
		Part_next:  -1,
		Part_name:  [16]byte{'N'},
	}

	err := ebr.SerializeEBR(fdisk.path, int32(offset))
	if err != nil {
		return err
	}
	return nil
}

func createLogicPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil {
		return err
	}

	if !mbr.CanFitAnotherDisk(sizeBytes) {
		return errors.New("no se puede crear una particion por falta de espacio")
	}

	if !mbr.IsThereExtendedPartition() {
		return errors.New("primero es necesario crear una particion extendida antes que una logica")
	}
	startPosition, extendedPartitionSize, err := mbr.GetOffsetFirstEBR()
	if err != nil {
		return err
	}
	var ebr structures.EBR

	offset, err := ebr.DeserializeEBRAvailable(fdisk.path, int32(startPosition))
	if err != nil {
		return err
	}

	if extendedPartitionSize+startPosition < offset+int32(binary.Size(ebr))+int32(sizeBytes) {
		return errors.New("no se puede crear la particion logica porque excede el tam de la particion extendida")
	}

	// fmt.Println("Antes de modificar el EBR: ")
	// ebr.PrintEBR()

	ebr.SetNeoEBR(fdisk.fit, int(offset), sizeBytes, fdisk.name)

	// fmt.Println("Se ha seteado un nuevo EBR: ")
	// ebr.PrintEBR()

	err = ebr.SerializeEBR(fdisk.path, int32(ebr.Part_start-int32(binary.Size(ebr))))
	if err != nil {
		return err
	}

	// fmt.Println("El nuevo deberia estar en: ")
	// fmt.Println(ebr.Part_next)

	err = createEBR(fdisk, int(ebr.Part_next))
	if err != nil {
		return err
	}

	return nil

}
