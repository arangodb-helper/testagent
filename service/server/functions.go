package server

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
