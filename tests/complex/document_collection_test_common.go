package complex

import (
	"time"
)

type DocumentCollectionTest interface {
	createTestDatabase()
	createTestCollection()
	createDocuments()
	readDocuments()
	updateDocuments()
	dropTestDatabase()
	dropTestCollection()
}

type DocColConfig struct {
	MaxDocuments int
	MaxUpdates   int
	BatchSize    int
	DocumentSize int
}

type DocColTest struct {
	ComplextTest
	DocColConfig
	DocColTestImpl           DocumentCollectionTest
	numberOfExistingDocs     int
	numberOfCreatedDocsTotal int64
	docCollectionCreated     bool
	docCollectionName        string
	readOffset               int
	updateOffset             int
}

func (t *DocColTest) runTest() {
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
			// create a database
			t.DocColTestImpl.createTestDatabase()
			planIndex++

		case 1:
			// create a document collection
			t.DocColTestImpl.createTestCollection()
			planIndex++

		case 2:
			// create documents
			t.DocColTestImpl.createDocuments()
			planIndex++

		case 3:
			// read documents
			t.DocColTestImpl.readDocuments()
			planIndex++

		case 4:
			// update documents
			t.DocColTestImpl.updateDocuments()
			planIndex++

		case 5:
			// drop collections
			t.DocColTestImpl.dropTestCollection()
			planIndex++

		case 6:
			// drop database
			t.DocColTestImpl.dropTestDatabase()
			planIndex++
		}
		time.Sleep(time.Second * 2)
	}
}

func (t *DocColTest) createDocuments() {
	if t.docCollectionCreated && t.numberOfExistingDocs < t.MaxDocuments {
		var thisBatchSize int
		if t.BatchSize <= t.MaxDocuments-t.numberOfExistingDocs {
			thisBatchSize = t.BatchSize
		} else {
			thisBatchSize = t.MaxDocuments - t.numberOfExistingDocs
		}

		for i := 0; i < thisBatchSize; i++ {
			seed := t.documentIdSeq
			t.documentIdSeq++
			document := NewBigDocument(seed, t.DocumentSize)
			if err := t.insertDocument(t.docCollectionName, document); err != nil {
				t.log.Errorf("Failed to create document with key '%s' in collection '%s': %v",
					document.Key, t.docCollectionName, err)
			} else {
				t.actions++
				t.existingDocuments = append(t.existingDocuments, document.TestDocument)
				t.numberOfExistingDocs++
				t.numberOfCreatedDocsTotal++
			}
		}
	}
}

func (t *DocColTest) readDocuments() {
	if t.docCollectionCreated && t.numberOfExistingDocs >= t.BatchSize {
		var upperBound int
		var lowerBound int = t.readOffset
		if t.numberOfExistingDocs-t.readOffset < t.BatchSize {
			upperBound = t.numberOfExistingDocs
		} else {
			upperBound = t.readOffset + t.BatchSize
		}

		for _, testDoc := range t.existingDocuments[lowerBound:upperBound] {
			expectedDocument := NewBigDocumentFromTestDocument(testDoc, t.DocumentSize)
			if err := t.readExistingDocument(t.docCollectionName, expectedDocument, false); err != nil {
				t.log.Errorf("Failed to read document: %v", err)
			} else {
				t.actions++
			}
		}
		if upperBound == t.numberOfExistingDocs {
			t.readOffset = 0
		} else {
			t.readOffset = upperBound
		}
	}
}

func (t *DocColTest) updateDocuments() {
	if t.docCollectionCreated && t.numberOfExistingDocs >= t.BatchSize {
		if t.updateOffset == t.numberOfExistingDocs {
			t.updateOffset = 0
		}
		var upperBound int
		var lowerBound int = t.updateOffset
		if t.numberOfExistingDocs-t.updateOffset < t.BatchSize {
			upperBound = t.numberOfExistingDocs
		} else {
			upperBound = t.updateOffset + t.BatchSize
		}
		for i := lowerBound; i < upperBound; i++ {
			oldDoc := t.existingDocuments[i]
			if newDoc, err := t.updateExistingDocument(t.docCollectionName, oldDoc); err != nil {
				t.log.Errorf("Failed to update document: %v", err)
			} else {
				t.existingDocuments[i] = *newDoc
				t.actions++
			}
		}
		t.updateOffset = upperBound
	}
}
