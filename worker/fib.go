package worker

func fib(index int64) int64 {
	if index < 2 {
		return 1
	}
	return fib(index-1) + fib(index-2)
}
