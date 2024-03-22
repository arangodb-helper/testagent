package complex

import (
	"strconv"

	"github.com/arangodb-helper/testagent/service/test"
	logging "github.com/op/go-logging"
)

type SmartGraphTest struct {
	GraphTest
}

func NewSmartGraphTest(log *logging.Logger, reportDir string, complexTestCfg ComplextTestConfig, config GraphTestConf) test.TestScript {
	smartGraph := &SmartGraphTest{
		NewGraphTest("smartGraphTest", log, reportDir, complexTestCfg, config),
	}
	smartGraph.GraphTestImpl = smartGraph
	smartGraph.ComplexTestImpl = smartGraph
	return smartGraph
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

func (t *SmartGraphTest) createGraphAndCollections() {
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
}
