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
	mu         sync.RWMutex
	users      map[string]*User
	tenants    map[string]*Tenant
	workspaces map[string]*Workspace
	documents  map[string]*Document
	shares     map[string][]DocumentShare // documentID -> shares
	projects   map[string]*Project
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:      make(map[string]*User),
		tenants:    make(map[string]*Tenant),
		workspaces: make(map[string]*Workspace),
		documents:  make(map[string]*Document),
		shares:     make(map[string][]DocumentShare),
		projects:   make(map[string]*Project),
	}
}

// User operations

func (s *MemoryStore) CreateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; exists {
		return ErrAlreadyExists
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	s.users[user.ID] = user
	return nil
}

func (s *MemoryStore) GetUser(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrNotFound
	}
	return user, nil
}

func (s *MemoryStore) GetUserByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) UpdateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; !exists {
		return ErrNotFound
	}

	user.UpdatedAt = time.Now()
	s.users[user.ID] = user
	return nil
}

func (s *MemoryStore) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return ErrNotFound
	}

	delete(s.users, id)
	return nil
}

func (s *MemoryStore) ListUsers() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

func (s *MemoryStore) SetPlatformAdmin(userID string, isAdmin bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return ErrNotFound
	}

	user.IsPlatformAdmin = isAdmin
	user.UpdatedAt = time.Now()
	return nil
}

// Tenant operations

func (s *MemoryStore) CreateTenant(tenant *Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[tenant.ID]; exists {
		return ErrAlreadyExists
	}

	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	s.tenants[tenant.ID] = tenant
	return nil
}

func (s *MemoryStore) GetTenant(id string) (*Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenant, exists := s.tenants[id]
	if !exists {
		return nil, ErrNotFound
	}
	return tenant, nil
}

func (s *MemoryStore) ListTenants() []*Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenants := make([]*Tenant, 0, len(s.tenants))
	for _, tenant := range s.tenants {
		tenants = append(tenants, tenant)
	}
	return tenants
}

func (s *MemoryStore) DeleteTenant(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[id]; !exists {
		return ErrNotFound
	}

	delete(s.tenants, id)
	return nil
}

// Workspace operations

func (s *MemoryStore) CreateWorkspace(workspace *Workspace) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workspaces[workspace.ID]; exists {
		return ErrAlreadyExists
	}

	workspace.CreatedAt = time.Now()
	workspace.UpdatedAt = time.Now()
	s.workspaces[workspace.ID] = workspace
	return nil
}

func (s *MemoryStore) GetWorkspace(id string) (*Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workspace, exists := s.workspaces[id]
	if !exists {
		return nil, ErrNotFound
	}
	return workspace, nil
}

func (s *MemoryStore) ListWorkspaces() []*Workspace {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workspaces := make([]*Workspace, 0, len(s.workspaces))
	for _, workspace := range s.workspaces {
		workspaces = append(workspaces, workspace)
	}
	return workspaces
}

func (s *MemoryStore) ListWorkspacesByTenant(tenantID string) []*Workspace {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var workspaces []*Workspace
	for _, workspace := range s.workspaces {
		if workspace.TenantID == tenantID {
			workspaces = append(workspaces, workspace)
		}
	}
	return workspaces
}

func (s *MemoryStore) DeleteWorkspace(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workspaces[id]; !exists {
		return ErrNotFound
	}

	delete(s.workspaces, id)
	return nil
}

// Platform stats

func (s *MemoryStore) GetPlatformStats() *PlatformStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	adminCount := 0
	for _, user := range s.users {
		if user.IsPlatformAdmin {
			adminCount++
		}
	}

	return &PlatformStats{
		TotalUsers:      len(s.users),
		TotalTenants:    len(s.tenants),
		TotalWorkspaces: len(s.workspaces),
		TotalDocuments:  len(s.documents),
		TotalProjects:   len(s.projects),
		AdminCount:      adminCount,
	}
}

// GetAllDocuments returns all documents (for admin)
func (s *MemoryStore) GetAllDocuments() []*Document {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docs := make([]*Document, 0, len(s.documents))
	for _, doc := range s.documents {
		docs = append(docs, doc)
	}
	return docs
}

// GetAllProjects returns all projects (for admin)
func (s *MemoryStore) GetAllProjects() []*Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	projects := make([]*Project, 0, len(s.projects))
	for _, proj := range s.projects {
		projects = append(projects, proj)
	}
	return projects
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
