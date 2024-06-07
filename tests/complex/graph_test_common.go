package complex

import (
	"fmt"
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
	logging "github.com/op/go-logging"
)

type GraphTestInt interface {
	createGraphAndCollections()
	createVertexDocs()
	createEdgeDocs()
	traverseGraph()
	dropGraphAndCollections()
}

type GraphTestConf struct {
	MaxVertices                 int64
	VertexSize                  int
	EdgeSize                    int
	BatchSize                   int
	TraversalOperationsPerCycle int
}

type GraphTest struct {
	ComplextTest
	GraphTestConf
	GraphTestImpl           GraphTestInt
	graphCreated            bool
	edgeColCreated          bool
	vertexColCreated        bool
	graphIsBroken           bool
	vertexColName           string
	edgeColName             string
	graphName               string
	numberOfCreatedVertices int64
	numberOfCreatedEdges    int64
	vertexCreationOffset    int64
	edgeCreationOffset      int64
	existingEdgeDocuments   []TestDocument
	existingVertexDocuments []TestDocument
}

func NewGraphTest(testName string, log *logging.Logger, reportDir string, complexTestCfg ComplextTestConfig, config GraphTestConf) GraphTest {
	return GraphTest{
		GraphTestConf: config,
		ComplextTest: ComplextTest{
			TestName: testName,
			ComplextTestContext: ComplextTestContext{
				ComplextTestConfig: complexTestCfg,
				ComplextTestHarness: ComplextTestHarness{
					reportDir: reportDir,
					log:       log,
				},
				documentIdSeq:     0,
				collectionNameSeq: 0,
			},
		},
		vertexCreationOffset:    0,
		edgeCreationOffset:      0,
		numberOfCreatedVertices: 0,
		numberOfCreatedEdges:    0,
		graphCreated:            false,
		edgeColCreated:          false,
		vertexColCreated:        false,
		graphIsBroken:           false,
		existingVertexDocuments: make([]TestDocument, 0, config.MaxVertices),
		existingEdgeDocuments:   make([]TestDocument, 0, int64(float64(config.MaxVertices)*1.5)),
	}
}

func minimum64(x int64, y int64) int64 {
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

func calculateShortestOutboundPathLength(from int64, to int64) int {
	if from == 0 {
		return 10 + calculateShortestOutboundPathLength(10, to)
	}
	steps := 0
	for {
		length := int(to - from)
		inc := 0
		if length < 10 {
			inc += length
			steps += inc
			break
		} else {
			inc += int(10-from%10) * boolToInt(from%10 > 0)
			inc += int(to % 10)
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

func (t *GraphTest) createVertexDocs() {
	if t.vertexColCreated && t.numberOfCreatedVertices < t.MaxVertices && !t.graphIsBroken {
		var thisBatchSize int
		if int64(t.BatchSize) <= t.MaxVertices-t.numberOfCreatedVertices {
			thisBatchSize = t.BatchSize
		} else {
			thisBatchSize = int(t.MaxVertices - t.numberOfCreatedVertices)
		}
		oldVertexCreationOffsetValue := t.vertexCreationOffset
		maxVertexCreationOffsetValue := oldVertexCreationOffsetValue + int64(thisBatchSize)
		// for i := 0; i < thisBatchSize; i++ {
		for {
			if t.vertexCreationOffset >= maxVertexCreationOffsetValue || t.pauseRequested {
				break
			}
			seed := t.documentIdSeq
			t.documentIdSeq++
			doc := NewBigDocumentWithName(seed, t.VertexSize, strconv.FormatInt(t.vertexCreationOffset, 10))
			doc.Key = doc.Name + ":" + doc.Key
			e := t.insertDocument(t.vertexColName, doc)
			t.actions++
			if e == nil {
				t.numberOfCreatedVertices++
				t.existingVertexDocuments = append(t.existingVertexDocuments, doc.TestDocument)
				t.vertexCreationOffset++
			} else {
				t.log.Errorf("Failed to create vertex document: %v", e)
				t.graphIsBroken = true
			}
		}
	}
}

func (t *GraphTest) traverseGraph() {
	if t.vertexColCreated && t.edgeColCreated && t.graphCreated && !t.graphIsBroken {
		for i := 0; i < t.TraversalOperationsPerCycle; i++ {
			if t.pauseRequested {
				break
			}
			maxLength := minimum64(t.edgeCreationOffset, 100000)
			length := randInt64(2, maxLength)
			startIdx := randInt64(0, t.edgeCreationOffset-length)
			endIdx := startIdx + length
			from := t.vertexColName + "/" + t.existingVertexDocuments[startIdx].Key
			to := t.vertexColName + "/" + t.existingVertexDocuments[endIdx].Key
			expectedLength := calculateShortestOutboundPathLength(startIdx, endIdx)
			if err := t.traverse(to, from, t.graphName, expectedLength); err != nil {
				t.log.Errorf("Failed to traverse graph: %v", err)
			} else {
				t.actions++
			}
		}
	}
}

func (t *GraphTest) dropGraphAndCollections() {
	if t.graphIsBroken || (t.graphCreated && t.MaxVertices <= t.numberOfCreatedVertices) {
		if err := t.dropGraph(t.graphName, true); err != nil {
			t.log.Errorf("Failed to drop graph: %v", err)
		} else {
			t.graphCreated = false
			t.graphIsBroken = false
			t.vertexColCreated = false
			t.edgeColCreated = false
			t.collectionNameSeq++
			t.numberOfCreatedEdges = 0
			t.numberOfCreatedVertices = 0
			t.existingVertexDocuments = t.existingVertexDocuments[:0]
			t.existingEdgeDocuments = t.existingEdgeDocuments[:0]
			t.edgeCreationOffset = 0
			t.vertexCreationOffset = 0
			t.actions++
		}
	}
}

func (t *GraphTest) createEdgeDocs() {
	if t.edgeColCreated && t.edgeCreationOffset < t.vertexCreationOffset && !t.graphIsBroken {
		edgeCreationOffsetOldValue := t.edgeCreationOffset
		for {
			if t.edgeCreationOffset >= t.vertexCreationOffset-1 || t.pauseRequested {
				break
			}
			idx := t.edgeCreationOffset
			from := t.existingVertexDocuments[idx].Key
			to := t.existingVertexDocuments[idx+1].Key
			if err := t.createEdge(to, from, t.edgeColName, t.vertexColName, t.EdgeSize); err != nil {
				t.log.Errorf("Failed to create edge document: %v", err)
				t.graphIsBroken = true
			} else {
				t.actions++
				t.numberOfCreatedEdges++
				t.edgeCreationOffset++
			}
		}
		for idx := int64(1); idx < t.vertexCreationOffset-1; idx++ {
			for div := int64(10); div <= idx; div = div * 10 {
				if idx%div == 0 && idx+div < t.edgeCreationOffset && idx+div >= edgeCreationOffsetOldValue {
					from := t.existingVertexDocuments[idx].Key
					to := t.existingVertexDocuments[idx+div].Key
					if err := t.createEdge(to, from, t.edgeColName, t.vertexColName, t.EdgeSize); err != nil {
						t.log.Errorf("Failed to create edge document: %v", err)
						t.graphIsBroken = true
					} else {
						t.actions++
						t.numberOfCreatedEdges++
					}
				}
			}
		}
	}
}

// Status returns the current status of the test
func (t *GraphTest) Status() test.TestStatus {
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
			cc("#graph traversals performed", t.traverseGraphCounter),
		},
	}

	status.Messages = append(status.Messages,
		fmt.Sprintf("Number of vertex documents in the database: %d", t.numberOfCreatedVertices),
		fmt.Sprintf("Number of edge documents in the database: %d", t.numberOfCreatedEdges),
	)

	return status
}

func (t *GraphTest) runTest() {
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
			time.Sleep(t.StepTimeout)
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
			t.GraphTestImpl.createGraphAndCollections()
			planIndex++

		case 1:
			// create vertex documents
			t.GraphTestImpl.createVertexDocs()
			planIndex++

		case 2:
			// create edges
			t.GraphTestImpl.createEdgeDocs()
			planIndex++

		case 3:
			//traverse graph
			t.GraphTestImpl.traverseGraph()
			planIndex++

		case 4:
			// drop graph and collections
			t.GraphTestImpl.dropGraphAndCollections()
			planIndex++
		}
		time.Sleep(t.StepTimeout)
	}
}
