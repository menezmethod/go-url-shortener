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

	// Add tests for other service methods like GetLinkByShortURL, UpdateLink, DeleteLink, etc.
})
