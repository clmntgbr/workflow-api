package quota

const MaxExecutorPriority = 10

func NormalizeExecutorPriority(priority int) uint8 {
	if priority <= 0 {
		return 0
	}
	if priority > MaxExecutorPriority {
		return MaxExecutorPriority
	}
	return uint8(priority)
}
