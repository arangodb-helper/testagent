package server

import "time"

func cssReady(ready bool) string {
	if ready {
		return "ready"
	} else {
		return "notready"
	}
}

func cssTestOK(failures int) string {
	if failures == 0 {
		return "testok"
	} else {
		return "testfailures"
	}
}

func formatTime(value time.Time) string {
	return value.Format("2006-01-02 15:04:05")
}
