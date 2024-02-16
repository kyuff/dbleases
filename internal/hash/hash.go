package hash

import "hash/fnv"

const (
	// apogee advisory lock. Consider making this a hash of the stream type
	advisoryLockPID = 49053
)

func Hash(s string) int {
	h := fnv.New32()
	_, _ = h.Write([]byte(s))
	return int(h.Sum32())
}

func Mod(s string, mod int) int {
	return Hash(s) % mod
}
