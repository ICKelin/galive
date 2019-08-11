package galive

const (
	cmd_syn = iota
	cmd_heartbeat
	cmd_data
)

type hdr [4]byte

func (h hdr) cmd() int {
	return int(h[0])
}

func (h hdr) version() int {
	return int(h[1])
}

func (h hdr) bodylen() int {
	return (int(h[2]) << 8) + int(h[3])
}
