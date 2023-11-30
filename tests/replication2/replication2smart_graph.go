package replication2

import (
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

type SmartGraphTest struct {
	CommunityGraphTest
	existingVertexKeys []string
}

func NewSmartGraphTest(log *logging.Logger, reportDir string, rep2config Replication2Config, config CommunityGraphConf) test.TestScript {
	return &SmartGraphTest{
		CommunityGraphTest{
			CommunityGraphConf: config,
			Replication2Test: Replication2Test{
				TestName: "smartGraphTest",
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
		},
		make([]string, 0, config.MaxVertices),
	}
}

func minimum(x int, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

func boolToInt(val bool) int {
	if val {
		return 1
	} else {
		return 0
	}
}

func calculateShortestPathLength(from int, to int) int {
	steps := 0
	for {
		length := to - from
		inc := 0
		if length < 10 {
			inc += length
			steps += inc
			break
		} else {
			inc += (10 - from%10) * boolToInt(from%10 > 0)
			inc += to % 10
		}
		steps += inc
		to = to / 10
		if from%10 > 0 {
			from = from/10 + 1
		} else {
			from = from / 10
		}
		if to == 0 && from == 0 {
			break
		}
	}
	return steps
}

func (t *SmartGraphTest) generateVertexCollectionName(seed int64) string {
	return "smart_vertices_" + strconv.FormatInt(seed, 10)
}

func (t *SmartGraphTest) generateEdgeCollectionName(seed int64) string {
	return "smart_edges_" + strconv.FormatInt(seed, 10)
}

func (t *SmartGraphTest) generateGraphName(seed int64) string {
	return "smart_graph_" + strconv.FormatInt(seed, 10)
}

func (t *SmartGraphTest) testLoop() {
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
					nil, true, false, "name", nil, t.NumberOfShards, t.ReplicationFactor, t.ReplicationFactor-1); err != nil {
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
			if t.vertexColCreated {
				for {
					if t.numberOfCreatedVertices >= t.MaxVertices {
						break
					}
					for i := 0; i < t.MaxVertices-t.numberOfCreatedVertices; i++ {
						seed := t.documentIdSeq
						t.documentIdSeq++
						doc := NewBigDocument(seed, t.VertexSize)
						doc.Key = doc.Name + ":" + doc.Key
						t.insertDocument(t.vertexColName, doc)
						t.actions++
						t.numberOfCreatedVertices++
						t.existingDocSeeds = append(t.existingDocSeeds, seed)
						t.existingVertexKeys = append(t.existingVertexKeys, doc.Key)
					}

				}
			}
			planIndex++

		case 2:
			// create edges
			if t.edgeColCreated {
				i := 0
				for {
					idx := i
					if len(t.existingVertexKeys) <= i+2 {
						break
					}
					from := t.existingVertexKeys[idx]
					to := t.existingVertexKeys[idx+1]
					if err := t.createEdge(to, from, t.edgeColName, t.vertexColName, t.EdgeSize); err != nil {
						t.log.Errorf("Failed to create edge document: %v", err)
					} else {
						t.actions++
						t.numberOfCreatedEdges++
					}
					if i%10 == 0 && i+10 < len(t.existingVertexKeys) {
						idx = i
						from := t.existingVertexKeys[idx]
						to := t.existingVertexKeys[idx+10]
						if err := t.createEdge(to, from, t.edgeColName, t.vertexColName, t.EdgeSize); err != nil {
							t.log.Errorf("Failed to create edge document: %v", err)
						} else {
							t.actions++
							t.numberOfCreatedEdges++
						}
					}
					if i%100 == 0 && i+100 < len(t.existingVertexKeys) {
						idx = i
						from := t.existingVertexKeys[idx]
						to := t.existingVertexKeys[idx+100]
						if err := t.createEdge(to, from, t.edgeColName, t.vertexColName, t.EdgeSize); err != nil {
							t.log.Errorf("Failed to create edge document: %v", err)
						} else {
							t.actions++
							t.numberOfCreatedEdges++
						}
					}
					if i%1000 == 0 && i+1000 < len(t.existingVertexKeys) {
						idx = i
						from := t.existingVertexKeys[idx]
						to := t.existingVertexKeys[idx+1000]
						if err := t.createEdge(to, from, t.edgeColName, t.vertexColName, t.EdgeSize); err != nil {
							t.log.Errorf("Failed to create edge document: %v", err)
						} else {
							t.actions++
							t.numberOfCreatedEdges++
						}
					}
					i++
				}
			}
			planIndex++

		case 3:
			//traverse graph
			if t.vertexColCreated && t.edgeColCreated && t.graphCreated {
				i := 0
				for {
					if i >= 1000 {
						break
					}
					maxLength := minimum(t.numberOfCreatedVertices-1, 5000)
					length := randInt(2, maxLength)
					startIdx := randInt(0, t.numberOfCreatedVertices-1-length)
					endIdx := startIdx + length
					from := t.vertexColName + "/" + t.existingVertexKeys[startIdx]
					to := t.vertexColName + "/" + t.existingVertexKeys[endIdx]
					expectedLength := calculateShortestPathLength(startIdx, endIdx)
					if err := t.traverseGraph(to, from, t.graphName, expectedLength); err != nil {
						t.log.Errorf("Failed to traverse graph: %v", err)
					} else {
						t.actions++
					}
					i++
				}
			}
			planIndex++

		case 4:
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

// Start triggers the test script to start.
// It should spwan actions in a go routine.
func (t *SmartGraphTest) Start(cluster cluster.Cluster, listener test.TestListener) error {
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

func (t *SmartGraphTest) Status() test.TestStatus {
	return t.CommunityGraphTest.Status()
}
