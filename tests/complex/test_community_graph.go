package complex

import (
	"strconv"

	"github.com/arangodb-helper/testagent/service/test"
	logging "github.com/op/go-logging"
)

type CommunityGraphTest struct {
	GraphTest
}

func NewComGraphTest(log *logging.Logger, reportDir string, rep2config ComplextTestConfig, config GraphTestConf) test.TestScript {
	comGraphTest := &CommunityGraphTest{
		NewGraphTest("communityGraphTest", log, reportDir, rep2config, config),
	}
	comGraphTest.GraphTestImpl = comGraphTest
	comGraphTest.ComplexTestImpl = comGraphTest
	return comGraphTest
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

func (t *CommunityGraphTest) createVertexCollections() {
	if !t.vertexColCreated {
		t.vertexColName = t.generateVertexCollectionName(t.collectionNameSeq)
		if err := t.createCollection(t.vertexColName, false); err != nil {
			t.log.Errorf("Failed to create collection: %v", err)
		} else {
			t.vertexColCreated = true
			t.actions++
		}
	}
}

func (t *CommunityGraphTest) createEdgeCollections() {
	if !t.edgeColCreated {
		t.edgeColName = t.generateEdgeCollectionName(t.collectionNameSeq)
		if err := t.createCollection(t.edgeColName, true); err != nil {
			t.log.Errorf("Failed to create collection: %v", err)
		} else {
			t.edgeColCreated = true
			t.actions++
		}
	}
}

func (t *CommunityGraphTest) createNamedGraph() {
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
}

func (t *CommunityGraphTest) createGraphAndCollections() {
	t.createVertexCollections()
	t.createEdgeCollections()
	t.createNamedGraph()
}
