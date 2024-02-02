package complex

import (
	"fmt"

	"github.com/arangodb-helper/testagent/service/test"
	logging "github.com/op/go-logging"
)

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

func NewGraphTest(testName string, log *logging.Logger, reportDir string, rep2config ComplextTestConfig, config GraphTestConf) GraphTest {
	return GraphTest{
		GraphTestConf: config,
		ComplextTest: ComplextTest{
			TestName: testName,
			ComplextTestContext: ComplextTestContext{
				ComplextTestConfig: rep2config,
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

func minimum(x int, y int) int {
	if x < y {
		return x
	} else {
		return y
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

func (t *GraphTest) createVertices() {
	if t.vertexColCreated && t.numberOfCreatedVertices < t.MaxVertices && !t.graphIsBroken {
		var thisBatchSize int
		if int64(t.BatchSize) <= t.MaxVertices-t.numberOfCreatedVertices {
			thisBatchSize = t.BatchSize
		} else {
			thisBatchSize = int(t.MaxVertices - t.numberOfCreatedVertices)
		}
		for i := 0; i < thisBatchSize; i++ {
			seed := t.documentIdSeq
			t.documentIdSeq++
			doc := NewBigDocument(seed, t.VertexSize)
			doc.Key = doc.Name + ":" + doc.Key
			e := t.insertDocument(t.vertexColName, doc)
			t.actions++
			if e == nil {
				t.numberOfCreatedVertices++
				t.existingVertexDocuments = append(t.existingVertexDocuments, doc.TestDocument)
			} else {
				t.log.Errorf("Failed to create vertex document: %v", e)
				t.graphIsBroken = true
			}
		}
		t.vertexCreationOffset += int64(thisBatchSize)
	}
}

func (t *GraphTest) performGraphTraversal() {
	if t.vertexColCreated && t.edgeColCreated && t.graphCreated && !t.graphIsBroken {
		for i := 0; i < t.TraversalOperationsPerCycle; i++ {
			maxLength := minimum64(t.numberOfCreatedVertices-1, 100000)
			length := randInt64(2, maxLength)
			startIdx := randInt64(0, t.numberOfCreatedVertices-1-length)
			endIdx := startIdx + length
			from := t.vertexColName + "/" + t.existingVertexDocuments[startIdx].Key
			to := t.vertexColName + "/" + t.existingVertexDocuments[endIdx].Key
			expectedLength := calculateShortestOutboundPathLength(startIdx, endIdx)
			if err := t.traverseGraph(to, from, t.graphName, expectedLength); err != nil {
				t.log.Errorf("Failed to traverse graph: %v", err)
			} else {
				t.actions++
			}
		}
	}
}

func (t *GraphTest) dropGraphAndCollections() {
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

func (t *GraphTest) createEdges() {
	if t.edgeColCreated && t.edgeCreationOffset < t.vertexCreationOffset && !t.graphIsBroken {
		for i := t.edgeCreationOffset; i < t.vertexCreationOffset-1; i++ {
			from := t.existingVertexDocuments[i].Key
			to := t.existingVertexDocuments[i+1].Key
			if err := t.createEdge(to, from, t.edgeColName, t.vertexColName, t.EdgeSize); err != nil {
				t.log.Errorf("Failed to create edge document: %v", err)
				t.graphIsBroken = true
			} else {
				t.actions++
				t.numberOfCreatedEdges++
			}
		}
		for i := int64(1); i < t.vertexCreationOffset-1; i++ {
			for div := int64(10); div <= i; div = div * 10 {
				if i%div == 0 && i+div < int64(len(t.existingVertexDocuments)) && i+div > t.edgeCreationOffset {
					from := t.existingVertexDocuments[i].Key
					to := t.existingVertexDocuments[i+div].Key
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
		t.edgeCreationOffset = t.vertexCreationOffset - 2
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
