package data

type Record struct {
	len  uint32
	data []byte
	crc  uint32
}
