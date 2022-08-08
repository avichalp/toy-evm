package main

type Memory struct {
	memory []uint8
}

func NewMemory() *Memory {
	m := make([]uint8, 0)
	return &Memory{memory: m}
}

func (m *Memory) Store(offset uint64, value uint8) {
	/* if offset < 0 || offset > uint64(math.Pow(2, 256))-1 {
		panic(fmt.Sprintf("Invalid memory access %d", offset))
	} */

	// expand if needed
	// TODO: the offset should be uint256 not uint64!
	if offset >= uint64(len(m.memory)) {
		for i := 0; i < int(offset-uint64(len(m.memory))+1); i++ {
			m.memory = append(m.memory, 0)
		}
	}
	m.memory[offset] = value
}

func (m *Memory) Load(offset uint64) uint8 {
	if offset >= uint64(len(m.memory)) {
		return 0
	}
	return m.memory[offset]
}

func (m *Memory) LoadRange(offset uint64, length uint64) []byte {
	loaded := make([]byte, 0)
	for o := offset; o < offset+length; o++ {
		loaded = append(loaded, m.Load(o))
	}
	return loaded
}
