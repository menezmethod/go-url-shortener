package service_test

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/menezmethod/ref_go/internal/domain"
	"github.com/menezmethod/ref_go/internal/service"
	"github.com/menezmethod/ref_go/internal/testutils/mocks"
)

func TestLinkService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Link Service Suite")
}

var _ = Describe("LinkService", func() {
	var (
		mockRepo *mocks.MockLinkRepository
		srv      *service.LinkService
	)

	BeforeEach(func() {
		mockRepo = &mocks.MockLinkRepository{}
		srv = service.NewLinkService(mockRepo)
	})

	Describe("CreateLink", func() {
		Context("when the link creation is successful", func() {
			BeforeEach(func() {
				mockRepo.CreateFunc = func(link *domain.Link) error {
					// Simulate successful creation
					return nil
				}
			})

			It("creates a link successfully", func() {
				createReq := service.CreateLinkRequest{
					UserID:      "user-123",
					OriginalURL: "https://example.com",
					CustomAlias: "mylink",
				}

				link, err := srv.CreateLink(createReq)

				Expect(err).NotTo(HaveOccurred())
				Expect(link).NotTo(BeNil())
				Expect(link.UserID).To(Equal("user-123"))
				Expect(link.OriginalURL).To(Equal("https://example.com"))
				Expect(link.ShortURL).To(Equal("mylink"))
			})

			It("generates a short URL when custom alias is not provided", func() {
				createReq := service.CreateLinkRequest{
					UserID:      "user-123",
					OriginalURL: "https://example.com",
					CustomAlias: "", // No custom alias
				}

				link, err := srv.CreateLink(createReq)

				Expect(err).NotTo(HaveOccurred())
				Expect(link).NotTo(BeNil())
				Expect(link.UserID).To(Equal("user-123"))
				Expect(link.OriginalURL).To(Equal("https://example.com"))
				// The short URL should be generated
				Expect(link.ShortURL).NotTo(BeEmpty())
			})
		})

		Context("when the repository returns an error", func() {
			BeforeEach(func() {
				mockRepo.CreateFunc = func(link *domain.Link) error {
					return errors.New("database error")
				}
			})

			It("returns the error", func() {
				createReq := service.CreateLinkRequest{
					UserID:      "user-123",
					OriginalURL: "https://example.com",
				}

				link, err := srv.CreateLink(createReq)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(link).To(BeNil())
			})
		})

		Context("when the URL is invalid", func() {
			It("returns a validation error", func() {
				createReq := service.CreateLinkRequest{
					UserID:      "user-123",
					OriginalURL: "invalid-url",
				}

				link, err := srv.CreateLink(createReq)

				Expect(err).To(HaveOccurred())
				// In a real implementation, we'd expect a domain.ErrValidation error here
				Expect(link).To(BeNil())
			})
		})

		Context("when the custom alias is already taken", func() {
			BeforeEach(func() {
				mockRepo.GetByShortURLFunc = func(shortURL string) (*domain.Link, error) {
					// Simulate that the alias is already taken
					return &domain.Link{
						ID:          "existing-id",
						ShortURL:    "mylink",
						OriginalURL: "https://another-example.com",
					}, nil
				}
			})

			It("returns a conflict error", func() {
				createReq := service.CreateLinkRequest{
					UserID:      "user-123",
					OriginalURL: "https://example.com",
					CustomAlias: "mylink", // This alias is already taken
				}

				link, err := srv.CreateLink(createReq)

				Expect(err).To(HaveOccurred())
				// In a real implementation, we'd expect a domain.ErrConflict error here
				Expect(link).To(BeNil())
			})
		})
	})

	Describe("GetLink", func() {
		Context("when the link exists", func() {
			BeforeEach(func() {
				mockRepo.GetByIDFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "mylink",
						Visits:      10,
					}, nil
				}
			})

			It("returns the link", func() {
				link, err := srv.GetLink("link-123")

				Expect(err).NotTo(HaveOccurred())
				Expect(link).NotTo(BeNil())
				Expect(link.ID).To(Equal("link-123"))
				Expect(link.UserID).To(Equal("user-123"))
				Expect(link.OriginalURL).To(Equal("https://example.com"))
				Expect(link.ShortURL).To(Equal("mylink"))
				Expect(link.Visits).To(Equal(10))
			})
		})

		Context("when the link doesn't exist", func() {
			BeforeEach(func() {
				mockRepo.GetByIDFunc = func(id string) (*domain.Link, error) {
					return nil, domain.ErrNotFound
				}
			})

			It("returns a not found error", func() {
				link, err := srv.GetLink("non-existent-id")

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(domain.ErrNotFound))
				Expect(link).To(BeNil())
			})
		})
	})

	Describe("GetLinkByShortURL", func() {
		Context("when the link exists", func() {
			BeforeEach(func() {
				mockRepo.GetByShortURLFunc = func(shortURL string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "mylink",
						Visits:      10,
					}, nil
				}
			})

			It("returns the link", func() {
				link, err := srv.GetLinkByShortURL("mylink")

				Expect(err).NotTo(HaveOccurred())
				Expect(link).NotTo(BeNil())
				Expect(link.ID).To(Equal("link-123"))
				Expect(link.UserID).To(Equal("user-123"))
				Expect(link.OriginalURL).To(Equal("https://example.com"))
				Expect(link.ShortURL).To(Equal("mylink"))
				Expect(link.Visits).To(Equal(10))
			})
		})

		Context("when the link doesn't exist", func() {
			BeforeEach(func() {
				mockRepo.GetByShortURLFunc = func(shortURL string) (*domain.Link, error) {
					return nil, domain.ErrNotFound
				}
			})

			It("returns a not found error", func() {
				link, err := srv.GetLinkByShortURL("non-existent-link")

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(domain.ErrNotFound))
				Expect(link).To(BeNil())
			})
		})
	})

	Describe("UpdateLink", func() {
		Context("when the link exists", func() {
			BeforeEach(func() {
				mockRepo.GetByIDFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "mylink",
					}, nil
				}
				mockRepo.UpdateFunc = func(link *domain.Link) error {
					return nil
				}
			})

			It("updates the link successfully", func() {
				updateReq := service.UpdateLinkRequest{
					OriginalURL: "https://updated-example.com",
					CustomAlias: "newlink",
				}

				link, err := srv.UpdateLink("link-123", updateReq)

				Expect(err).NotTo(HaveOccurred())
				Expect(link).NotTo(BeNil())
				Expect(link.ID).To(Equal("link-123"))
				Expect(link.OriginalURL).To(Equal("https://updated-example.com"))
				Expect(link.ShortURL).To(Equal("newlink"))
			})

			It("updates only the original URL if custom alias is not provided", func() {
				updateReq := service.UpdateLinkRequest{
					OriginalURL: "https://updated-example.com",
				}

				link, err := srv.UpdateLink("link-123", updateReq)

				Expect(err).NotTo(HaveOccurred())
				Expect(link).NotTo(BeNil())
				Expect(link.OriginalURL).To(Equal("https://updated-example.com"))
				Expect(link.ShortURL).To(Equal("mylink")) // Unchanged
			})

			It("updates only the custom alias if original URL is not provided", func() {
				updateReq := service.UpdateLinkRequest{
					CustomAlias: "newlink",
				}

				link, err := srv.UpdateLink("link-123", updateReq)

				Expect(err).NotTo(HaveOccurred())
				Expect(link).NotTo(BeNil())
				Expect(link.OriginalURL).To(Equal("https://example.com")) // Unchanged
				Expect(link.ShortURL).To(Equal("newlink"))
			})
		})

		Context("when the link doesn't exist", func() {
			BeforeEach(func() {
				mockRepo.GetByIDFunc = func(id string) (*domain.Link, error) {
					return nil, domain.ErrNotFound
				}
			})

			It("returns a not found error", func() {
				updateReq := service.UpdateLinkRequest{
					OriginalURL: "https://updated-example.com",
				}

				link, err := srv.UpdateLink("non-existent-id", updateReq)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(domain.ErrNotFound))
				Expect(link).To(BeNil())
			})
		})

		Context("when the repository update fails", func() {
			BeforeEach(func() {
				mockRepo.GetByIDFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "mylink",
					}, nil
				}
				mockRepo.UpdateFunc = func(link *domain.Link) error {
					return errors.New("database error")
				}
			})

			It("returns the error", func() {
				updateReq := service.UpdateLinkRequest{
					OriginalURL: "https://updated-example.com",
				}

				link, err := srv.UpdateLink("link-123", updateReq)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(link).To(BeNil())
			})
		})
	})

	Describe("DeleteLink", func() {
		Context("when the link exists", func() {
			BeforeEach(func() {
				mockRepo.DeleteFunc = func(id string) error {
					return nil
				}
			})

			It("deletes the link successfully", func() {
				err := srv.DeleteLink("link-123")

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the link doesn't exist or there's a database error", func() {
			BeforeEach(func() {
				mockRepo.DeleteFunc = func(id string) error {
					return domain.ErrNotFound
				}
			})

			It("returns the error", func() {
				err := srv.DeleteLink("non-existent-id")

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(domain.ErrNotFound))
			})
		})
	})

	Describe("ListLinks", func() {
		Context("when links exist", func() {
			BeforeEach(func() {
				mockRepo.ListFunc = func(userID string, limit, offset int) ([]*domain.Link, error) {
					return []*domain.Link{
						{
							ID:          "link-1",
							UserID:      userID,
							OriginalURL: "https://example1.com",
							ShortURL:    "link1",
						},
						{
							ID:          "link-2",
							UserID:      userID,
							OriginalURL: "https://example2.com",
							ShortURL:    "link2",
						},
					}, nil
				}
				mockRepo.CountFunc = func(userID string) (int, error) {
					return 2, nil
				}
			})

			It("returns the links and total count", func() {
				links, count, err := srv.ListLinks("user-123", 1, 10)

				Expect(err).NotTo(HaveOccurred())
				Expect(links).To(HaveLen(2))
				Expect(count).To(Equal(2))
				Expect(links[0].ID).To(Equal("link-1"))
				Expect(links[1].ID).To(Equal("link-2"))
			})
		})

		Context("when no links exist", func() {
			BeforeEach(func() {
				mockRepo.ListFunc = func(userID string, limit, offset int) ([]*domain.Link, error) {
					return []*domain.Link{}, nil
				}
				mockRepo.CountFunc = func(userID string) (int, error) {
					return 0, nil
				}
			})

			It("returns an empty slice and zero count", func() {
				links, count, err := srv.ListLinks("user-123", 1, 10)

				Expect(err).NotTo(HaveOccurred())
				Expect(links).To(HaveLen(0))
				Expect(count).To(Equal(0))
			})
		})

		Context("when there's a database error in List", func() {
			BeforeEach(func() {
				mockRepo.ListFunc = func(userID string, limit, offset int) ([]*domain.Link, error) {
					return nil, errors.New("database error")
				}
			})

			It("returns the error", func() {
				links, count, err := srv.ListLinks("user-123", 1, 10)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(links).To(BeNil())
				Expect(count).To(Equal(0))
			})
		})

		Context("when there's a database error in Count", func() {
			BeforeEach(func() {
				mockRepo.ListFunc = func(userID string, limit, offset int) ([]*domain.Link, error) {
					return []*domain.Link{
						{
							ID:          "link-1",
							UserID:      "user-123",
							OriginalURL: "https://example1.com",
							ShortURL:    "link1",
						},
					}, nil
				}
				mockRepo.CountFunc = func(userID string) (int, error) {
					return 0, errors.New("database error")
				}
			})

			It("returns the error", func() {
				links, count, err := srv.ListLinks("user-123", 1, 10)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(links).To(BeNil())
				Expect(count).To(Equal(0))
			})
		})
	})

	Describe("RecordClick", func() {
		Context("when successful", func() {
			BeforeEach(func() {
				mockRepo.IncrementVisitsFunc = func(id string) error {
					return nil
				}
				mockRepo.CreateClickFunc = func(click *domain.Click) error {
					return nil
				}
			})

			It("records the click successfully", func() {
				err := srv.RecordClick("link-123", "Mozilla/5.0", "https://referrer.com", "192.168.1.1")

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when IncrementVisits fails", func() {
			BeforeEach(func() {
				mockRepo.IncrementVisitsFunc = func(id string) error {
					return errors.New("database error")
				}
			})

			It("returns the error", func() {
				err := srv.RecordClick("link-123", "Mozilla/5.0", "https://referrer.com", "192.168.1.1")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
			})
		})

		Context("when CreateClick fails", func() {
			BeforeEach(func() {
				mockRepo.IncrementVisitsFunc = func(id string) error {
					return nil
				}
				mockRepo.CreateClickFunc = func(click *domain.Click) error {
					return errors.New("database error")
				}
			})

			It("returns the error", func() {
				err := srv.RecordClick("link-123", "Mozilla/5.0", "https://referrer.com", "192.168.1.1")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
			})
		})
	})

	Describe("GetClicks", func() {
		Context("when clicks exist", func() {
			BeforeEach(func() {
				mockRepo.GetClicksFunc = func(linkID string, limit, offset int) ([]*domain.Click, error) {
					return []*domain.Click{
						{
							ID:        "click-1",
							LinkID:    linkID,
							UserAgent: "Mozilla/5.0",
							Referer:   "https://referrer1.com",
							IPAddress: "192.168.1.1",
						},
						{
							ID:        "click-2",
							LinkID:    linkID,
							UserAgent: "Chrome/90.0",
							Referer:   "https://referrer2.com",
							IPAddress: "192.168.1.2",
						},
					}, nil
				}
				mockRepo.CountClicksFunc = func(linkID string) (int, error) {
					return 2, nil
				}
			})

			It("returns the clicks and total count", func() {
				clicks, count, err := srv.GetClicks("link-123", 1, 10)

				Expect(err).NotTo(HaveOccurred())
				Expect(clicks).To(HaveLen(2))
				Expect(count).To(Equal(2))
				Expect(clicks[0].ID).To(Equal("click-1"))
				Expect(clicks[1].ID).To(Equal("click-2"))
			})
		})

		Context("when no clicks exist", func() {
			BeforeEach(func() {
				mockRepo.GetClicksFunc = func(linkID string, limit, offset int) ([]*domain.Click, error) {
					return []*domain.Click{}, nil
				}
				mockRepo.CountClicksFunc = func(linkID string) (int, error) {
					return 0, nil
				}
			})

			It("returns an empty slice and zero count", func() {
				clicks, count, err := srv.GetClicks("link-123", 1, 10)

				Expect(err).NotTo(HaveOccurred())
				Expect(clicks).To(HaveLen(0))
				Expect(count).To(Equal(0))
			})
		})

		Context("when there's a database error in GetClicks", func() {
			BeforeEach(func() {
				mockRepo.GetClicksFunc = func(linkID string, limit, offset int) ([]*domain.Click, error) {
					return nil, errors.New("database error")
				}
			})

			It("returns the error", func() {
				clicks, count, err := srv.GetClicks("link-123", 1, 10)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(clicks).To(BeNil())
				Expect(count).To(Equal(0))
			})
		})

		Context("when there's a database error in CountClicks", func() {
			BeforeEach(func() {
				mockRepo.GetClicksFunc = func(linkID string, limit, offset int) ([]*domain.Click, error) {
					return []*domain.Click{
						{
							ID:        "click-1",
							LinkID:    "link-123",
							UserAgent: "Mozilla/5.0",
							Referer:   "https://referrer1.com",
							IPAddress: "192.168.1.1",
						},
					}, nil
				}
				mockRepo.CountClicksFunc = func(linkID string) (int, error) {
					return 0, errors.New("database error")
				}
			})

			It("returns the error", func() {
				clicks, count, err := srv.GetClicks("link-123", 1, 10)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(clicks).To(BeNil())
				Expect(count).To(Equal(0))
			})
		})
	})

	// Add tests for other service methods like GetLinkByShortURL, UpdateLink, DeleteLink, etc.
})
