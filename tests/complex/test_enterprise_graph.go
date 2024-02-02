package complex

import (
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

type EnterpriseGraphTest struct {
	SmartGraphTest
}

func NewEnterpriseGraphTest(log *logging.Logger, reportDir string, rep2config ComplextTestConfig, config GraphTestConf) test.TestScript {
	return &EnterpriseGraphTest{SmartGraphTest{
		NewGraphTest("enterpriseGraphTest", log, reportDir, rep2config, config),
	}}
}

func (t *EnterpriseGraphTest) generateVertexCollectionName(seed int64) string {
	return "enterprise_vertices_" + strconv.FormatInt(seed, 10)
}

func (t *EnterpriseGraphTest) generateEdgeCollectionName(seed int64) string {
	return "enterprise_edges_" + strconv.FormatInt(seed, 10)
}

func (t *EnterpriseGraphTest) generateGraphName(seed int64) string {
	return "ent_graph_" + strconv.FormatInt(seed, 10)
}

func (t *EnterpriseGraphTest) testLoop() {
	t.active = true
	t.actions = 0
	defer func() { t.active = false }()

	var plan []int
	planIndex := 0
	for {
		// Should we stop
		if t.shouldStop() {
			return
		}
		if t.pauseRequested {
			t.paused = true
			time.Sleep(time.Second * 2)
			continue
		}
		t.paused = false
		t.actions++
		if plan == nil || planIndex >= len(plan) {
			plan = []int{0, 1, 2, 3, 4} // Update when more tests are added
			planIndex = 0
		}

		switch plan[planIndex] {
		case 0:
			// create graph and collections
			if !t.graphCreated {
				t.vertexColName = t.generateVertexCollectionName(t.collectionNameSeq)
				t.edgeColName = t.generateEdgeCollectionName(t.collectionNameSeq)
				t.graphName = t.generateGraphName(t.collectionNameSeq)
				if err := t.createGraph(
					t.graphName, t.edgeColName, []string{t.vertexColName}, []string{t.vertexColName},
					nil, true, false, "", nil, t.NumberOfShards, t.ReplicationFactor, t.ReplicationFactor-1); err != nil {
					t.log.Errorf("Failed to create graph: %v", err)
				} else {
					t.graphCreated = true
					t.vertexColCreated = true
					t.edgeColCreated = true
					t.actions++
				}
			}
			planIndex++

		case 1:
			// create vertex documents
			t.createVertices()
			planIndex++

		case 2:
			// create edges
			t.createEdges()
			planIndex++

		case 3:
			//traverse graph
			t.performGraphTraversal()
			planIndex++

		case 4:
			// drop graph and collections
			if t.graphIsBroken || (t.graphCreated && t.MaxVertices >= t.numberOfCreatedVertices) {
				t.dropGraphAndCollections()
			}
			planIndex++
		}
		time.Sleep(time.Second * 2)
	}
}

// Start triggers the test script to start.
// It should spwan actions in a go routine.
func (t *EnterpriseGraphTest) Start(cluster cluster.Cluster, listener test.TestListener) error {
	t.activeMutex.Lock()
	defer t.activeMutex.Unlock()

	if t.active {
		// No restart unless needed
		return nil
	}
	if err := t.setupLogger(cluster); err != nil {
		return maskAny(err)
	}

	t.cluster = cluster
	t.listener = listener
	t.client = util.NewArangoClient(t.log, cluster)

	t.active = true
	go t.testLoop()
	return nil
}

func (t *EnterpriseGraphTest) Status() test.TestStatus {
	return t.SmartGraphTest.GraphTest.Status()
}
