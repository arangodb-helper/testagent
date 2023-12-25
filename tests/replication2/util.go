package replication2

import "math/rand"

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randInt64(min int64, max int64) int64 {
	return min + rand.Int63n(max-min)
}
