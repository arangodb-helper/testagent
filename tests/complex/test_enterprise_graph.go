package complex

import (
	"strconv"

	"github.com/arangodb-helper/testagent/service/test"
	logging "github.com/op/go-logging"
)

type EnterpriseGraphTest struct {
	SmartGraphTest
}

func NewEnterpriseGraphTest(log *logging.Logger, reportDir string, complexTestCfg ComplextTestConfig, config GraphTestConf) test.TestScript {
	entGraphTest := &EnterpriseGraphTest{SmartGraphTest{
		NewGraphTest("enterpriseGraphTest", log, reportDir, complexTestCfg, config),
	}}
	entGraphTest.GraphTestImpl = entGraphTest
	entGraphTest.ComplexTestImpl = entGraphTest
	return entGraphTest
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

func (t *EnterpriseGraphTest) createGraphAndCollections() {
	if !t.graphCreated {
		t.vertexColName = t.generateVertexCollectionName(t.collectionNameSeq)
		t.edgeColName = t.generateEdgeCollectionName(t.collectionNameSeq)
		t.graphName = t.generateGraphName(t.collectionNameSeq)
		if err := t.createGraph(
			t.graphName, t.edgeColName, []string{t.vertexColName}, []string{t.vertexColName},
			nil, true, false, "", nil, t.NumberOfShards, t.ReplicationFactor, t.ReplicationFactor-1); err != nil {
			t.log.Errorf("Failed to create graph: %v", err)
			t.collectionNameSeq++
		} else {
			t.graphCreated = true
			t.vertexColCreated = true
			t.edgeColCreated = true
			t.actions++
		}
	}
}
