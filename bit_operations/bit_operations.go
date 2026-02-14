package bit_operations

// IsEven 使用最低位判断奇偶。
func IsEven(n int) bool {
	return n&1 == 0
}

func SetBit(n uint64, pos uint) uint64 {
	return n | (1 << pos)
}

func ClearBit(n uint64, pos uint) uint64 {
	return n &^ (1 << pos)
}

func ToggleBit(n uint64, pos uint) uint64 {
	return n ^ (1 << pos)
}

func HasBit(n uint64, pos uint) bool {
	return n&(1<<pos) != 0
}

// CountBits 使用 Brian Kernighan 算法统计 1 的个数。
func CountBits(n uint64) int {
	count := 0
	for n != 0 {
		n &= n - 1
		count++
	}
	return count
}

// IsPowerOfTwo 判断是否为 2 的幂（n>0 且仅一个 1）。
func IsPowerOfTwo(n uint64) bool {
	return n != 0 && (n&(n-1)) == 0
}

// NextPowerOfTwo 返回 >=n 的最小 2 的幂。
func NextPowerOfTwo(n uint64) uint64 {
	if n <= 1 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

// 权限位案例。
const (
	PermRead uint8 = 1 << iota
	PermWrite
	PermExecute
)

func GrantPerm(mask uint8, perm uint8) uint8 {
	return mask | perm
}

func RevokePerm(mask uint8, perm uint8) uint8 {
	return mask &^ perm
}

func HasPerm(mask uint8, perm uint8) bool {
	return mask&perm != 0
}
