package replication2

import (
	"fmt"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

func (t *Replication2Test) createOneShardDatabase(databaseName string) error {
	return t.createDatabase(databaseName, "single", t.ReplicationFactor, t.ReplicationFactor-1)
}

// createDatabase creates a new database.
// The operation is expected to succeed.
func (t *Replication2Test) createDatabase(databaseName string, sharding string, replicationFactor int, writeConcern int) error {
	body := struct {
		Name    string `json:"name"`
		Options struct {
			Sharding          string `json:"sharding"`
			ReplicationFactor int    `json:"replicationFactor"`
			WriteConcern      int    `json:"writeConcern"`
		} `json:"options"`
	}{Name: databaseName,
		Options: struct {
			Sharding          string `json:"sharding"`
			ReplicationFactor int    `json:"replicationFactor"`
			WriteConcern      int    `json:"writeConcern"`
		}{
			Sharding:          sharding,
			ReplicationFactor: replicationFactor,
			WriteConcern:      writeConcern,
		},
	}
	//operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	// For now, we increase the timeout to 5 minutes, since the cluster-internal
	// timeout is 4 minutes:
	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 5)

	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		checkRetry := false
		success := false
		shouldNotExist := false
		shouldExist := false

		t.log.Infof("Creating database '%s'...", databaseName)
		resp, err := t.client.Post(
			"/_api/database", nil, nil, body, "", nil, []int{0, 1, 201, 409, 500, 503},
			[]int{400, 404, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		// 0, 503: recheck without erxpectations
		//     there: good
		//     not there: retry
		// 200   : good
		// 1, 500: collection couldn't be finished.
		//     there: failure
		//     not there: retry
		// 409   :
		//     first attempt: failure
		//     later attempts:
		//     recheck
		//         there: done
		//         else : failure

		if err[0] == nil {
			if resp[0].StatusCode == 201 {
				success = true
			} else {
				if resp[0].StatusCode == 1 || resp[0].StatusCode == 500 { // connection refused or not created
					checkRetry = true
					shouldNotExist = true
					t.log.Debugf("Error code: %d\nError num: %d\nError message: %s", resp[0].Error_.Code, resp[0].Error_.ErrorNum, resp[0].Error_.ErrorMessage)
				} else if resp[0].StatusCode == 409 {
					if i == 1 {
						// This is a failure
						t.createDatabaseCounter.failed++
						t.reportFailure(test.NewFailure(t.Name(), "Failed to create database '%s': got 409 on first attempt", databaseName))
						return maskAny(fmt.Errorf("Failed to create database '%s': got 409 on first attempt", databaseName))
					} else {
						shouldExist = true
					}
				}
				checkRetry = true
			}
		} else {
			// This is a failure
			t.createDatabaseCounter.failed++
			t.reportFailure(test.NewFailure(t.Name(), "Failed to create database '%s': %v", databaseName, err[0]))
			return maskAny(err[0])
		}

		if checkRetry {

			t.log.Infof("Checking existence of database '%s' ...", databaseName)
			exists, checkErr := t.databaseExists(databaseName)
			t.log.Infof("... got result %v and error %v", exists, checkErr)

			if checkErr == nil {
				if exists {
					if shouldNotExist {
						// This is a failure
						t.createDatabaseCounter.failed++
						t.reportFailure(test.NewFailure(t.Name(),
							"Failed to create database '%s' rechecked and failed existence", databaseName))
						return maskAny(fmt.Errorf("Failed to create database '%s' rechecked and failed existence", databaseName))
					}
					success = true
				} else {
					if shouldExist {
						// This is a failure
						t.createDatabaseCounter.failed++
						t.reportFailure(test.NewFailure(t.Name(),
							"Failed to create database '%s' rechecked and failed existence", databaseName))
						return maskAny(fmt.Errorf("Failed to create database '%s' rechecked and failed existence", databaseName))
					}
				}
			} else {
				return maskAny(checkErr)
			}
		}

		if success {
			t.createDatabaseCounter.succeeded++
			t.log.Infof(
				"Creating database '%s' succeeded", databaseName)
			return nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.reportFailure(
		test.NewFailure(t.Name(), "Timed out while trying to create (%d) database %s.", i, databaseName))
	return maskAny(fmt.Errorf("Timed out while trying to create (%d) database %s.", i, databaseName))

}

// dropDatabase removes an existing database.
// The operation is expected to succeed.
func (t *Replication2Test) dropDatabase(databaseName string) error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(t.OperationTimeout * 5)

	url := fmt.Sprintf("/_api/database/%s", databaseName)
	backoff := time.Millisecond * 250
	i := 0

	success := false
	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		t.log.Infof("Removing (%d) database '%s'...", i, databaseName)
		resp, err := t.client.Delete(
			url, nil, nil, []int{0, 1, 200, 404, 500, 503}, []int{400, 409, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d", resp[0].StatusCode, resp[0].Error_.ErrorNum)

		if err[0] != nil {
			// This is a failure
			t.dropDatabaseCounter.failed++
			t.reportFailure(test.NewFailure(t.Name(), "Failed to drop database '%s': %v", databaseName, err[0]))
			return maskAny(err[0])
		} else if resp[0].StatusCode == 404 {
			// Database not found.
			// This can happen if the first attempt timed out, but did actually succeed.
			// So we accept this if there are multiple attempts.
			if i == 1 { // this is a failure in first run
				// Not enough attempts, this is a failure
				t.dropDatabaseCounter.failed++
				t.reportFailure(
					test.NewFailure(t.Name(), "Failed to drop database '%s': got 404 after only 1 attempt", databaseName))
				return maskAny(fmt.Errorf("Failed to drop database '%s': got 404 after only 1 attempt", databaseName))
			} else {
				success = true
			}
		} else if resp[0].StatusCode == 200 {
			success = true
		}

		if success {
			t.dropDatabaseCounter.succeeded++
			t.log.Infof("Droping database '%s' succeeded", databaseName)
			return nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	t.dropDatabaseCounter.failed++
	t.reportFailure(test.NewFailure(t.Name(), "Timed out (%d) while droping database '%s'", i, databaseName))
	return maskAny(fmt.Errorf("Timed out (%d) while droping database '%s'", i, databaseName))

}

func (t *Replication2Test) databaseExists(databaseName string) (bool, error) {

	type DatabasesResponse struct {
		Error  bool
		Code   string
		Result []string
	}

	operationTimeout := time.Duration(ReadTimeout) * time.Second
	timeout := time.Now().Add(operationTimeout)

	i := 0
	backoff := time.Millisecond * 250
	url := fmt.Sprintf("/_api/database")

	for {

		i++
		if time.Now().After(timeout) {
			break
		}

		t.log.Infof("Checking (%d) database '%s'...", i, databaseName)
		var response DatabasesResponse
		resp, err := t.client.Get(
			url, nil, nil, response, []int{0, 1, 200, 404, 503}, []int{400, 409, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d", resp[0].StatusCode, resp[0].Error_.ErrorNum)

		if err[0] != nil {
			// This is a failure
			t.log.Infof("Failed checking for database '%s': %v", databaseName, err[0])
			return false, maskAny(err[0])
		} else if resp[0].StatusCode == 404 {
			return false, nil
		} else if resp[0].StatusCode == 200 {
			for k := 0; k < len(response.Result); k++ {
				if response.Result[k] == databaseName {
					return true, nil
				}
			}
			return false, nil
		}

		// 0, 1, 503 retry
		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// This is a failure
	out := fmt.Errorf("Timed out checking for database '%s'", databaseName)
	t.log.Error(out)
	return false, maskAny(out)

}
