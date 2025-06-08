package structures

import "fmt"

type PARTITION struct {
	Part_status      [1]byte
	Part_type        [1]byte
	Part_fit         [1]byte
	Part_start       int32
	Part_size        int32
	Part_name        [16]byte
	Part_correlative int32
	Part_id          [4]byte
}

/*
Part Status:

	N: Disponible
	0: Creado
	1: Montado
*/
func (p *PARTITION) CreatePartition(partStart, partSize int, partType, partFit, partName string) {
	p.Part_status[0] = '0'

	p.Part_start = int32(partStart)

	p.Part_size = int32(partSize)

	if len(partType) > 0 {
		p.Part_type[0] = partType[0]
	}

	if len(partFit) > 0 {
		p.Part_fit[0] = partFit[0]
	}

	copy(p.Part_name[:], partName)
}

func (p *PARTITION) MountPartition(correlative int, id string) error {
	p.Part_status[0] = '1'

	p.Part_correlative = int32(correlative)

	copy(p.Part_id[:], id)

	return nil
}

func (p *PARTITION) PrintPartition() {
	fmt.Printf("Part_status: %c\n", p.Part_status[0])
	fmt.Printf("Part_type: %c\n", p.Part_type[0])
	fmt.Printf("Part_fit: %c\n", p.Part_fit[0])
	fmt.Printf("Part_start: %d\n", p.Part_start)
	fmt.Printf("Part_size: %d\n", p.Part_size)
	fmt.Printf("Part_name: %s\n", string(p.Part_name[:]))
	fmt.Printf("Part_correlative: %d\n", p.Part_correlative)
	fmt.Printf("Part_id: %s\n", string(p.Part_id[:]))
}
