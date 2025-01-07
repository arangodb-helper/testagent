package complex

import "math/rand"

func randInt64(min int64, max int64) int64 {
	return min + rand.Int63n(max-min)
}
