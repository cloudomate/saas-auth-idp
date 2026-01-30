package store

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// MemoryStore is an in-memory store for demo purposes
type MemoryStore struct {
	mu        sync.RWMutex
	documents map[string]*Document
	shares    map[string][]DocumentShare // documentID -> shares
	projects  map[string]*Project
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		documents: make(map[string]*Document),
		shares:    make(map[string][]DocumentShare),
		projects:  make(map[string]*Project),
	}
}

// Document operations

func (s *MemoryStore) CreateDocument(doc *Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.documents[doc.ID]; exists {
		return ErrAlreadyExists
	}

	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	s.documents[doc.ID] = doc

	// Create owner share
	s.shares[doc.ID] = []DocumentShare{
		{DocumentID: doc.ID, UserID: doc.OwnerID, Role: "owner"},
	}

	return nil
}

func (s *MemoryStore) GetDocument(id string) (*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, exists := s.documents[id]
	if !exists {
		return nil, ErrNotFound
	}
	return doc, nil
}

func (s *MemoryStore) UpdateDocument(doc *Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.documents[doc.ID]; !exists {
		return ErrNotFound
	}

	doc.UpdatedAt = time.Now()
	s.documents[doc.ID] = doc
	return nil
}

func (s *MemoryStore) DeleteDocument(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.documents[id]; !exists {
		return ErrNotFound
	}

	delete(s.documents, id)
	delete(s.shares, id)
	return nil
}

func (s *MemoryStore) ListDocuments(workspaceID string) []*Document {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var docs []*Document
	for _, doc := range s.documents {
		if doc.WorkspaceID == workspaceID {
			docs = append(docs, doc)
		}
	}
	return docs
}

func (s *MemoryStore) ListDocumentsForUser(workspaceID, userID string) []*Document {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var docs []*Document
	for _, doc := range s.documents {
		if doc.WorkspaceID != workspaceID {
			continue
		}

		// Check visibility
		switch doc.Visibility {
		case "public":
			docs = append(docs, doc)
		case "workspace":
			docs = append(docs, doc)
		case "private":
			// Only include if user has explicit access
			if doc.OwnerID == userID || s.hasDocumentAccess(doc.ID, userID) {
				docs = append(docs, doc)
			}
		}
	}
	return docs
}

func (s *MemoryStore) hasDocumentAccess(docID, userID string) bool {
	shares, exists := s.shares[docID]
	if !exists {
		return false
	}
	for _, share := range shares {
		if share.UserID == userID {
			return true
		}
	}
	return false
}

func (s *MemoryStore) AddDocumentShare(share DocumentShare) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.documents[share.DocumentID]; !exists {
		return ErrNotFound
	}

	// Check if already shared
	shares := s.shares[share.DocumentID]
	for _, existing := range shares {
		if existing.UserID == share.UserID {
			return ErrAlreadyExists
		}
	}

	s.shares[share.DocumentID] = append(shares, share)
	return nil
}

func (s *MemoryStore) GetDocumentShares(docID string) []DocumentShare {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.shares[docID]
}

func (s *MemoryStore) GetUserDocumentRole(docID, userID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shares := s.shares[docID]
	for _, share := range shares {
		if share.UserID == userID {
			return share.Role
		}
	}
	return ""
}

// Project operations

func (s *MemoryStore) CreateProject(proj *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[proj.ID]; exists {
		return ErrAlreadyExists
	}

	proj.CreatedAt = time.Now()
	proj.UpdatedAt = time.Now()
	s.projects[proj.ID] = proj
	return nil
}

func (s *MemoryStore) GetProject(id string) (*Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	proj, exists := s.projects[id]
	if !exists {
		return nil, ErrNotFound
	}
	return proj, nil
}

func (s *MemoryStore) UpdateProject(proj *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[proj.ID]; !exists {
		return ErrNotFound
	}

	proj.UpdatedAt = time.Now()
	s.projects[proj.ID] = proj
	return nil
}

func (s *MemoryStore) DeleteProject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[id]; !exists {
		return ErrNotFound
	}

	delete(s.projects, id)
	return nil
}

func (s *MemoryStore) ListProjects(workspaceID string) []*Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var projects []*Project
	for _, proj := range s.projects {
		if proj.WorkspaceID == workspaceID {
			projects = append(projects, proj)
		}
	}
	return projects
}

func (s *MemoryStore) ListProjectsByEnvironment(workspaceID, env string) []*Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var projects []*Project
	for _, proj := range s.projects {
		if proj.WorkspaceID == workspaceID && proj.Environment == env {
			projects = append(projects, proj)
		}
	}
	return projects
}
