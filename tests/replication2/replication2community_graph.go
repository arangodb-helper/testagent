package replication2

import (
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

type CommunityGraphTest struct {
	GraphTest
}

func NewComGraphTest(log *logging.Logger, reportDir string, rep2config Replication2Config, config GraphTestConf) test.TestScript {
	return &CommunityGraphTest{
		NewGraphTest("communityGraphTest", log, reportDir, rep2config, config),
	}
}

func (t *CommunityGraphTest) generateVertexCollectionName(seed int64) string {
	return "community_vertices_" + strconv.FormatInt(seed, 10)
}

func (t *CommunityGraphTest) generateEdgeCollectionName(seed int64) string {
	return "community_edges_" + strconv.FormatInt(seed, 10)
}

func (t *CommunityGraphTest) generateGraphName(seed int64) string {
	return "community_graph_" + strconv.FormatInt(seed, 10)
}

func (t *CommunityGraphTest) testLoop() {
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
			//create graph and underlying collections

			// create a vertex collection
			if !t.vertexColCreated {
				t.vertexColName = t.generateVertexCollectionName(t.collectionNameSeq)
				if err := t.createCollection(t.vertexColName, false); err != nil {
					t.log.Errorf("Failed to create collection: %v", err)
				} else {
					t.vertexColCreated = true
					t.actions++
				}
			}
			//create an edge collection
			if !t.edgeColCreated {
				t.edgeColName = t.generateEdgeCollectionName(t.collectionNameSeq)
				if err := t.createCollection(t.edgeColName, true); err != nil {
					t.log.Errorf("Failed to create collection: %v", err)
				} else {
					t.edgeColCreated = true
					t.actions++
				}
			}
			//create named graph
			if t.vertexColCreated && t.edgeColCreated && !t.graphCreated {
				t.graphName = t.generateGraphName(t.collectionNameSeq)
				if err := t.createGraph(t.graphName, t.edgeColName, []string{t.vertexColName}, []string{t.vertexColName},
					nil, false, false, "", nil, 0, 0, 0); err != nil {
					t.log.Errorf("Failed to create graph: %v", err)
				} else {
					t.graphCreated = true
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
			if t.graphIsBroken || (t.vertexColCreated && t.edgeColCreated && t.graphCreated &&
				t.MaxVertices == t.numberOfCreatedVertices) {
				t.dropGraphAndCollections()
			}
			planIndex++
		}
		time.Sleep(time.Second * 2)
	}
}

// Start triggers the test script to start.
// It should spwan actions in a go routine.
func (t *CommunityGraphTest) Start(cluster cluster.Cluster, listener test.TestListener) error {
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
