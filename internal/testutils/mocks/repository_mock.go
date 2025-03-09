package mocks

import (
	"github.com/menezmethod/ref_go/internal/domain"
)

// MockLinkRepository mocks the LinkRepository interface
type MockLinkRepository struct {
	CreateFunc          func(link *domain.Link) error
	GetByIDFunc         func(id string) (*domain.Link, error)
	GetByShortURLFunc   func(shortURL string) (*domain.Link, error)
	UpdateFunc          func(link *domain.Link) error
	DeleteFunc          func(id string) error
	ListFunc            func(userID string, limit, offset int) ([]*domain.Link, error)
	CountFunc           func(userID string) (int, error)
	IncrementVisitsFunc func(id string) error
	CreateClickFunc     func(click *domain.Click) error
	GetClicksFunc       func(linkID string, limit, offset int) ([]*domain.Click, error)
	CountClicksFunc     func(linkID string) (int, error)
}

// Create mocks the Create method
func (m *MockLinkRepository) Create(link *domain.Link) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(link)
	}
	return nil
}

// GetByID mocks the GetByID method
func (m *MockLinkRepository) GetByID(id string) (*domain.Link, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

// GetByShortURL mocks the GetByShortURL method
func (m *MockLinkRepository) GetByShortURL(shortURL string) (*domain.Link, error) {
	if m.GetByShortURLFunc != nil {
		return m.GetByShortURLFunc(shortURL)
	}
	return nil, nil
}

// Update mocks the Update method
func (m *MockLinkRepository) Update(link *domain.Link) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(link)
	}
	return nil
}

// Delete mocks the Delete method
func (m *MockLinkRepository) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

// List mocks the List method
func (m *MockLinkRepository) List(userID string, limit, offset int) ([]*domain.Link, error) {
	if m.ListFunc != nil {
		return m.ListFunc(userID, limit, offset)
	}
	return nil, nil
}

// Count mocks the Count method
func (m *MockLinkRepository) Count(userID string) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(userID)
	}
	return 0, nil
}

// IncrementVisits mocks the IncrementVisits method
func (m *MockLinkRepository) IncrementVisits(id string) error {
	if m.IncrementVisitsFunc != nil {
		return m.IncrementVisitsFunc(id)
	}
	return nil
}

// CreateClick mocks the CreateClick method
func (m *MockLinkRepository) CreateClick(click *domain.Click) error {
	if m.CreateClickFunc != nil {
		return m.CreateClickFunc(click)
	}
	return nil
}

// GetClicks mocks the GetClicks method
func (m *MockLinkRepository) GetClicks(linkID string, limit, offset int) ([]*domain.Click, error) {
	if m.GetClicksFunc != nil {
		return m.GetClicksFunc(linkID, limit, offset)
	}
	return nil, nil
}

// CountClicks mocks the CountClicks method
func (m *MockLinkRepository) CountClicks(linkID string) (int, error) {
	if m.CountClicksFunc != nil {
		return m.CountClicksFunc(linkID)
	}
	return 0, nil
}

// MockUserRepository mocks the UserRepository interface
type MockUserRepository struct {
	CreateFunc     func(user *domain.User) error
	GetByIDFunc    func(id string) (*domain.User, error)
	GetByEmailFunc func(email string) (*domain.User, error)
	UpdateFunc     func(user *domain.User) error
	DeleteFunc     func(id string) error
}

// Create mocks the Create method
func (m *MockUserRepository) Create(user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}

// GetByID mocks the GetByID method
func (m *MockUserRepository) GetByID(id string) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

// GetByEmail mocks the GetByEmail method
func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, nil
}

// Update mocks the Update method
func (m *MockUserRepository) Update(user *domain.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	return nil
}

// Delete mocks the Delete method
func (m *MockUserRepository) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}
