package main

import (
	"fmt"
	"sync"
)

type DocumentStore interface {
	CreateDocument(document Document) (Document, error)
	GetDocumentById(documentId string) (Document, bool)
	ListDocumentsByPatientId(patientId string) []Document
	UpdateDocument(document Document) error
}

type InMemoryDocumentStore struct {
	mutex           sync.Mutex
	nextDocumentNum int
	documentsById   map[string]Document
}

func NewInMemoryDocumentStore() *InMemoryDocumentStore {
	return &InMemoryDocumentStore{
		nextDocumentNum: 1,
		documentsById:   make(map[string]Document),
	}
}

func (store *InMemoryDocumentStore) CreateDocument(document Document) (Document, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	documentId := fmt.Sprintf("doc-%d", store.nextDocumentNum)
	store.nextDocumentNum += 1

	document.DocumentId = documentId
	store.documentsById[documentId] = document

	return document, nil
}

func (store *InMemoryDocumentStore) GetDocumentById(documentId string) (Document, bool) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	document, found := store.documentsById[documentId]
	return document, found
}

func (store *InMemoryDocumentStore) ListDocumentsByPatientId(patientId string) []Document {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	var documents []Document
	for _, document := range store.documentsById {
		if document.PatientId == patientId {
			documents = append(documents, document)
		}
	}
	return documents
}

func (store *InMemoryDocumentStore) UpdateDocument(document Document) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.documentsById[document.DocumentId] = document
	return nil
}
