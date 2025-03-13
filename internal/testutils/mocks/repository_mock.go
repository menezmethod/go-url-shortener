package mocks

import (
	"context"

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

// MockURLRepository mocks the URLRepository interface
type MockURLRepository struct {
	CreateFunc    func(ctx context.Context, url *domain.URL) error
	GetByIDFunc   func(ctx context.Context, id string) (*domain.URL, error)
	GetByHashFunc func(ctx context.Context, hash string) (*domain.URL, error)
}

// Create mocks the Create method
func (m *MockURLRepository) Create(ctx context.Context, url *domain.URL) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, url)
	}
	return nil
}

// GetByID mocks the GetByID method
func (m *MockURLRepository) GetByID(ctx context.Context, id string) (*domain.URL, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetByHash mocks the GetByHash method
func (m *MockURLRepository) GetByHash(ctx context.Context, hash string) (*domain.URL, error) {
	if m.GetByHashFunc != nil {
		return m.GetByHashFunc(ctx, hash)
	}
	return nil, nil
}

// MockShortLinkRepository mocks the ShortLinkRepository interface
type MockShortLinkRepository struct {
	CreateFunc           func(ctx context.Context, link *domain.ShortLink) error
	GetByIDFunc          func(ctx context.Context, id string) (*domain.ShortLink, error)
	GetByCodeFunc        func(ctx context.Context, code string) (*domain.ShortLink, error)
	GetByCustomAliasFunc func(ctx context.Context, alias string) (*domain.ShortLink, error)
	GetAllByURLIDFunc    func(ctx context.Context, urlID string) ([]*domain.ShortLink, error)
	UpdateFunc           func(ctx context.Context, link *domain.ShortLink) error
	DeleteFunc           func(ctx context.Context, id string) error
	ListFunc             func(ctx context.Context, offset, limit int) ([]*domain.ShortLink, error)
	CountFunc            func(ctx context.Context) (int, error)
}

// Create mocks the Create method
func (m *MockShortLinkRepository) Create(ctx context.Context, link *domain.ShortLink) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, link)
	}
	return nil
}

// GetByID mocks the GetByID method
func (m *MockShortLinkRepository) GetByID(ctx context.Context, id string) (*domain.ShortLink, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetByCode mocks the GetByCode method
func (m *MockShortLinkRepository) GetByCode(ctx context.Context, code string) (*domain.ShortLink, error) {
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, code)
	}
	return nil, nil
}

// GetByCustomAlias mocks the GetByCustomAlias method
func (m *MockShortLinkRepository) GetByCustomAlias(ctx context.Context, alias string) (*domain.ShortLink, error) {
	if m.GetByCustomAliasFunc != nil {
		return m.GetByCustomAliasFunc(ctx, alias)
	}
	return nil, nil
}

// GetAllByURLID mocks the GetAllByURLID method
func (m *MockShortLinkRepository) GetAllByURLID(ctx context.Context, urlID string) ([]*domain.ShortLink, error) {
	if m.GetAllByURLIDFunc != nil {
		return m.GetAllByURLIDFunc(ctx, urlID)
	}
	return nil, nil
}

// Update mocks the Update method
func (m *MockShortLinkRepository) Update(ctx context.Context, link *domain.ShortLink) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, link)
	}
	return nil
}

// Delete mocks the Delete method
func (m *MockShortLinkRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// List mocks the List method
func (m *MockShortLinkRepository) List(ctx context.Context, offset, limit int) ([]*domain.ShortLink, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, offset, limit)
	}
	return nil, nil
}

// Count mocks the Count method
func (m *MockShortLinkRepository) Count(ctx context.Context) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx)
	}
	return 0, nil
}

// MockLinkClickRepository mocks the LinkClickRepository interface
type MockLinkClickRepository struct {
	CreateFunc                func(ctx context.Context, click *domain.LinkClick) error
	GetByShortLinkIDFunc      func(ctx context.Context, shortLinkID string, offset, limit int) ([]*domain.LinkClick, error)
	GetStatsByShortLinkIDFunc func(ctx context.Context, shortLinkID string) (*domain.LinkStats, error)
}

// Create mocks the Create method
func (m *MockLinkClickRepository) Create(ctx context.Context, click *domain.LinkClick) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, click)
	}
	return nil
}

// GetByShortLinkID mocks the GetByShortLinkID method
func (m *MockLinkClickRepository) GetByShortLinkID(ctx context.Context, shortLinkID string, offset, limit int) ([]*domain.LinkClick, error) {
	if m.GetByShortLinkIDFunc != nil {
		return m.GetByShortLinkIDFunc(ctx, shortLinkID, offset, limit)
	}
	return nil, nil
}

// GetStatsByShortLinkID mocks the GetStatsByShortLinkID method
func (m *MockLinkClickRepository) GetStatsByShortLinkID(ctx context.Context, shortLinkID string) (*domain.LinkStats, error) {
	if m.GetStatsByShortLinkIDFunc != nil {
		return m.GetStatsByShortLinkIDFunc(ctx, shortLinkID)
	}
	return nil, nil
}
