package replication2

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

type CommunityGraphConf struct {
	MaxEdges    int
	MaxVertices int
	VertexSize  int
	EdgeSize    int
}

type CommunityGraphTest struct {
	Replication2Test
	CommunityGraphConf
	graphCreated            bool
	edgeColCreated          bool
	vertexColCreated        bool
	vertexColName           string
	edgeColName             string
	graphName               string
	numberOfCreatedVertices int
	numberOfCreatedEdges    int
}

func NewComGraphTest(log *logging.Logger, reportDir string, rep2config Replication2Config, config CommunityGraphConf) test.TestScript {
	return &CommunityGraphTest{
		CommunityGraphConf: config,
		Replication2Test: Replication2Test{
			Replication2TestContext: Replication2TestContext{
				Replication2Config: rep2config,
				Replication2TestHarness: Replication2TestHarness{
					reportDir: reportDir,
					log:       log,
				},
				documentIdSeq:     0,
				collectionNameSeq: 0,
			},
		},
	}
}

func (t *CommunityGraphTest) generateVertexCollectionName(seed int64) string {
	return "replication2_vertices_" + strconv.FormatInt(seed, 10)
}

func (t *CommunityGraphTest) generateEdgeCollectionName(seed int64) string {
	return "replication2_edges_" + strconv.FormatInt(seed, 10)
}

func (t *CommunityGraphTest) generateGraphName(seed int64) string {
	return "simple_named_graph_" + strconv.FormatInt(seed, 10)
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
			plan = []int{0, 1, 2, 3, 4, 5, 6} // Update when more tests are added
			planIndex = 0
		}

		switch plan[planIndex] {
		case 0:
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
			planIndex++

		case 1:
			// create vertex documents
			if t.vertexColCreated {
				for {
					if t.numberOfCreatedVertices >= t.MaxVertices {
						break
					}
					for i := 0; i < t.MaxVertices-t.numberOfCreatedVertices; i++ {
						seed := t.documentIdSeq
						t.documentIdSeq++
						t.insertDocument(t.vertexColName, NewBigDocument(seed, t.VertexSize))
						t.actions++
						t.numberOfCreatedVertices++
						t.existingDocSeeds = append(t.existingDocSeeds, seed)
					}

				}
			}
			planIndex++

		case 2:
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
			planIndex++

		case 3:
			// create edges
			if t.edgeColCreated {
				for {
					if t.numberOfCreatedEdges >= t.MaxEdges {
						break
					}
					for i := 0; i < t.MaxEdges-t.numberOfCreatedEdges; i++ {
						from := strconv.FormatInt(t.existingDocSeeds[rand.Intn(len(t.existingDocSeeds))], 10)
						to := strconv.FormatInt(t.existingDocSeeds[rand.Intn(len(t.existingDocSeeds))], 10)
						t.createEdge(from, to, t.edgeColName, t.vertexColName, t.EdgeSize)
						t.actions++
						t.numberOfCreatedEdges++
					}
				}
			}
			planIndex++

		case 4:
			//create a named graph
			if t.vertexColCreated && t.edgeColCreated {
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

		case 5:
			//traverse graph
			planIndex++

		case 6:
			// drop graph and collections
			if t.vertexColCreated && t.edgeColCreated && t.graphCreated {
				if err := t.dropGraph(t.graphName, true); err != nil {
					t.log.Errorf("Failed to drop graph: %v", err)
				} else {
					t.graphCreated = false
					t.vertexColCreated = false
					t.edgeColCreated = false
					t.collectionNameSeq++
					t.numberOfCreatedEdges = 0
					t.numberOfCreatedVertices = 0
					t.existingDocSeeds = t.existingDocSeeds[:0]
					t.actions++
				}
			}
			planIndex++
		}
		time.Sleep(time.Second * 2)
	}
}

// Status returns the current status of the test
func (t *CommunityGraphTest) Status() test.TestStatus {
	cc := func(name string, c counter) test.Counter {
		return test.Counter{
			Name:      name,
			Succeeded: c.succeeded,
			Failed:    c.failed,
		}
	}

	status := test.TestStatus{
		Active:   t.active && !t.paused,
		Pausing:  t.pauseRequested,
		Failures: t.failures,
		Actions:  t.actions,
		Counters: []test.Counter{
			cc("#collections created", t.createCollectionCounter),
			cc("#collections dropped", t.dropCollectionCounter),
			cc("#graphs created", t.createGraphCounter),
			cc("#graphs dropped", t.dropGraphCounter),
			cc("#edges created", t.edgeDocumentCreateCounter),
			cc("#vertex documents created", t.singleDocCreateCounter),
		},
	}

	status.Messages = append(status.Messages,
		fmt.Sprintf("Number of vertex documents in the database: %d", t.numberOfCreatedVertices),
		fmt.Sprintf("Number of edge documents in the database: %d", t.numberOfCreatedEdges),
	)

	return status
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

// Name returns the name of the script
func (t *CommunityGraphTest) Name() string {
	return "communityGraphTest"
}
