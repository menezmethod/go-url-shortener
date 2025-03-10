package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/menezmethod/ref_go/internal/cache"
	"github.com/menezmethod/ref_go/internal/domain"
	"github.com/menezmethod/ref_go/internal/service"
	"github.com/menezmethod/ref_go/internal/testutils/mocks"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Suite")
}

var _ = Describe("Service Suite", func() {
	// LinkService tests
	Describe("LinkService", func() {
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
					Expect(link.OriginalURL).To(Equal("https://updated-example.com"))
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
						OriginalURL: "https://example.com",
					}

					link, err := srv.UpdateLink("non-existent-id", updateReq)

					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(domain.ErrNotFound))
					Expect(link).To(BeNil())
				})
			})
		})

		Describe("DeleteLink", func() {
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
					mockRepo.DeleteFunc = func(id string) error {
						return nil
					}
				})

				It("deletes the link successfully", func() {
					err := srv.DeleteLink("link-123")

					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the link doesn't exist", func() {
				BeforeEach(func() {
					mockRepo.GetByIDFunc = func(id string) (*domain.Link, error) {
						return nil, domain.ErrNotFound
					}
					mockRepo.DeleteFunc = func(id string) error {
						return domain.ErrNotFound
					}
				})

				It("returns a not found error", func() {
					err := srv.DeleteLink("non-existent-id")

					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(domain.ErrNotFound))
				})
			})
		})

		Describe("ListLinks", func() {
			Context("when listing links successfully", func() {
				BeforeEach(func() {
					mockRepo.ListFunc = func(userID string, limit, offset int) ([]*domain.Link, error) {
						links := []*domain.Link{
							{
								ID:          "link-1",
								UserID:      "user-123",
								OriginalURL: "https://example1.com",
								ShortURL:    "abc123",
								Visits:      5,
							},
							{
								ID:          "link-2",
								UserID:      "user-123",
								OriginalURL: "https://example2.com",
								ShortURL:    "def456",
								Visits:      10,
							},
						}
						return links, nil
					}

					mockRepo.CountFunc = func(userID string) (int, error) {
						return 2, nil
					}
				})

				It("returns the list of links and total count", func() {
					links, total, err := srv.ListLinks("user-123", 1, 10)

					Expect(err).NotTo(HaveOccurred())
					Expect(links).To(HaveLen(2))
					Expect(total).To(Equal(2))
					Expect(links[0].ID).To(Equal("link-1"))
					Expect(links[1].ID).To(Equal("link-2"))
				})
			})

			Context("when there's an error getting links", func() {
				BeforeEach(func() {
					mockRepo.ListFunc = func(userID string, limit, offset int) ([]*domain.Link, error) {
						return nil, errors.New("database error")
					}
				})

				It("returns the error", func() {
					links, total, err := srv.ListLinks("user-123", 1, 10)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("database error"))
					Expect(links).To(BeNil())
					Expect(total).To(Equal(0))
				})
			})

			Context("when there's an error getting total count", func() {
				BeforeEach(func() {
					mockRepo.ListFunc = func(userID string, limit, offset int) ([]*domain.Link, error) {
						return []*domain.Link{}, nil
					}

					mockRepo.CountFunc = func(userID string) (int, error) {
						return 0, errors.New("database error")
					}
				})

				It("returns the error", func() {
					links, total, err := srv.ListLinks("user-123", 1, 10)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("database error"))
					Expect(links).To(BeNil())
					Expect(total).To(Equal(0))
				})
			})
		})

		Describe("RecordClick", func() {
			Context("when recording a click successfully", func() {
				BeforeEach(func() {
					mockRepo.IncrementVisitsFunc = func(linkID string) error {
						return nil
					}

					mockRepo.CreateClickFunc = func(click *domain.Click) error {
						return nil
					}
				})

				It("records the click and increments visits", func() {
					err := srv.RecordClick("link-123", "Mozilla/5.0", "https://referrer.com", "127.0.0.1")

					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when there's an error incrementing visits", func() {
				BeforeEach(func() {
					mockRepo.IncrementVisitsFunc = func(linkID string) error {
						return errors.New("database error")
					}
				})

				It("returns the error", func() {
					err := srv.RecordClick("link-123", "Mozilla/5.0", "https://referrer.com", "127.0.0.1")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("database error"))
				})
			})

			Context("when there's an error creating click record", func() {
				BeforeEach(func() {
					mockRepo.IncrementVisitsFunc = func(linkID string) error {
						return nil
					}

					mockRepo.CreateClickFunc = func(click *domain.Click) error {
						return errors.New("database error")
					}
				})

				It("returns the error", func() {
					err := srv.RecordClick("link-123", "Mozilla/5.0", "https://referrer.com", "127.0.0.1")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("database error"))
				})
			})
		})

		Describe("GetClicks", func() {
			Context("when getting clicks successfully", func() {
				BeforeEach(func() {
					mockRepo.GetClicksFunc = func(linkID string, limit, offset int) ([]*domain.Click, error) {
						clicks := []*domain.Click{
							{
								ID:        "click-1",
								LinkID:    "link-123",
								UserAgent: "Mozilla/5.0",
								Referer:   "https://referrer1.com",
								IPAddress: "127.0.0.1",
							},
							{
								ID:        "click-2",
								LinkID:    "link-123",
								UserAgent: "Chrome/90.0",
								Referer:   "https://referrer2.com",
								IPAddress: "127.0.0.2",
							},
						}
						return clicks, nil
					}

					mockRepo.CountClicksFunc = func(linkID string) (int, error) {
						return 2, nil
					}
				})

				It("returns the list of clicks and total count", func() {
					clicks, total, err := srv.GetClicks("link-123", 1, 10)

					Expect(err).NotTo(HaveOccurred())
					Expect(clicks).To(HaveLen(2))
					Expect(total).To(Equal(2))
					Expect(clicks[0].ID).To(Equal("click-1"))
					Expect(clicks[1].ID).To(Equal("click-2"))
				})
			})

			Context("when there's an error getting clicks", func() {
				BeforeEach(func() {
					mockRepo.GetClicksFunc = func(linkID string, limit, offset int) ([]*domain.Click, error) {
						return nil, errors.New("database error")
					}
				})

				It("returns the error", func() {
					clicks, total, err := srv.GetClicks("link-123", 1, 10)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("database error"))
					Expect(clicks).To(BeNil())
					Expect(total).To(Equal(0))
				})
			})

			Context("when there's an error getting total count", func() {
				BeforeEach(func() {
					mockRepo.GetClicksFunc = func(linkID string, limit, offset int) ([]*domain.Click, error) {
						return []*domain.Click{}, nil
					}

					mockRepo.CountClicksFunc = func(linkID string) (int, error) {
						return 0, errors.New("database error")
					}
				})

				It("returns the error", func() {
					clicks, total, err := srv.GetClicks("link-123", 1, 10)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("database error"))
					Expect(clicks).To(BeNil())
					Expect(total).To(Equal(0))
				})
			})
		})
	})

	// URLShortenerService tests
	Describe("URLShortenerService", func() {
		var (
			mockURLRepo       *mocks.MockURLRepository
			mockShortLinkRepo *mocks.MockShortLinkRepository
			mockClickRepo     *mocks.MockLinkClickRepository
			logger            *zap.Logger
			svc               *service.URLShortenerService
			ctx               context.Context
		)

		BeforeEach(func() {
			mockURLRepo = &mocks.MockURLRepository{}
			mockShortLinkRepo = &mocks.MockShortLinkRepository{}
			mockClickRepo = &mocks.MockLinkClickRepository{}
			logger = zaptest.NewLogger(GinkgoT())
			ctx = context.Background()

			svc = service.NewURLShortenerService(
				mockURLRepo,
				mockShortLinkRepo,
				mockClickRepo,
				logger,
				"https://short.example.com",
				30*24*time.Hour,
			)
		})

		Describe("CreateShortLink", func() {
			var (
				req *domain.CreateShortLinkRequest
			)

			BeforeEach(func() {
				req = &domain.CreateShortLinkRequest{
					URL: "https://example.com/some-long-url",
				}

				mockURLRepo.GetByHashFunc = func(ctx context.Context, hash string) (*domain.URL, error) {
					return nil, errors.New("not found")
				}

				mockURLRepo.CreateFunc = func(ctx context.Context, url *domain.URL) error {
					return nil
				}

				mockShortLinkRepo.GetByCodeFunc = func(ctx context.Context, code string) (*domain.ShortLink, error) {
					return nil, errors.New("not found")
				}

				mockShortLinkRepo.CreateFunc = func(ctx context.Context, link *domain.ShortLink) error {
					return nil
				}

				mockURLRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.URL, error) {
					return &domain.URL{
						ID:          id,
						OriginalURL: req.URL,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil
				}
			})

			Context("when creating a short link with auto-generated code", func() {
				It("should create a short link successfully", func() {
					link, err := svc.CreateShortLink(ctx, req)

					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
					Expect(link.Code).NotTo(BeEmpty())
					Expect(link.IsActive).To(BeTrue())
					Expect(link.ExpirationDate).NotTo(BeNil())
					expectedExpiry := time.Now().Add(30 * 24 * time.Hour).Truncate(time.Second)
					Expect(link.ExpirationDate.Truncate(time.Second)).To(BeTemporally("~", expectedExpiry, time.Second))
				})
			})

			Context("when creating a short link with custom alias", func() {
				BeforeEach(func() {
					customAlias := "my-custom-alias"
					req.CustomAlias = &customAlias

					mockShortLinkRepo.GetByCustomAliasFunc = func(ctx context.Context, alias string) (*domain.ShortLink, error) {
						return nil, errors.New("not found")
					}
				})

				It("should create a short link with the custom alias", func() {
					link, err := svc.CreateShortLink(ctx, req)

					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
					Expect(link.CustomAlias).NotTo(BeNil())
					Expect(*link.CustomAlias).To(Equal("my-custom-alias"))
					Expect(link.Code).To(Equal("my-custom-alias"))
				})
			})

			Context("when the custom alias is already taken", func() {
				BeforeEach(func() {
					customAlias := "taken-alias"
					req.CustomAlias = &customAlias

					mockShortLinkRepo.GetByCustomAliasFunc = func(ctx context.Context, alias string) (*domain.ShortLink, error) {
						return &domain.ShortLink{
							ID:   "existing-id",
							Code: "taken-alias",
						}, nil
					}
				})

				It("should return an error", func() {
					link, err := svc.CreateShortLink(ctx, req)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("custom alias already in use"))
					Expect(link).To(BeNil())
				})
			})

			Context("when the URL is invalid", func() {
				BeforeEach(func() {
					req.URL = "invalid-url"
				})

				It("should return an error", func() {
					link, err := svc.CreateShortLink(ctx, req)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("invalid URL"))
					Expect(link).To(BeNil())
				})
			})

			Context("when the URL already exists", func() {
				BeforeEach(func() {
					mockURLRepo.GetByHashFunc = func(ctx context.Context, hash string) (*domain.URL, error) {
						return &domain.URL{
							ID:          "existing-url-id",
							OriginalURL: req.URL,
							Hash:        hash,
						}, nil
					}
				})

				It("should reuse the existing URL ID", func() {
					link, err := svc.CreateShortLink(ctx, req)

					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
					Expect(link.URLID).To(Equal("existing-url-id"))
				})
			})

			Context("when there's a code collision", func() {
				BeforeEach(func() {
					callCount := 0
					mockShortLinkRepo.GetByCodeFunc = func(ctx context.Context, code string) (*domain.ShortLink, error) {
						callCount++
						if callCount == 1 {
							return &domain.ShortLink{
								ID:   "existing-id",
								Code: code,
							}, nil
						}
						return nil, errors.New("not found")
					}
				})

				It("should retry with a different code", func() {
					link, err := svc.CreateShortLink(ctx, req)

					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
					Expect(link.Code).NotTo(BeEmpty())
				})
			})
		})

		Describe("GetShortLink", func() {
			Context("when the short link exists", func() {
				BeforeEach(func() {
					mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
						return &domain.ShortLink{
							ID:        id,
							Code:      "abc123",
							URLID:     "url-123",
							IsActive:  true,
							CreatedAt: time.Now(),
						}, nil
					}

					mockURLRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.URL, error) {
						return &domain.URL{
							ID:          id,
							OriginalURL: "https://example.com",
							CreatedAt:   time.Now(),
						}, nil
					}
				})

				It("should return the short link with URL details", func() {
					link, err := svc.GetShortLink(ctx, "link-123")

					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
					Expect(link.Code).To(Equal("abc123"))
					Expect(link.URL).NotTo(BeNil())
					Expect(link.URL.OriginalURL).To(Equal("https://example.com"))
				})
			})

			Context("when the short link doesn't exist", func() {
				BeforeEach(func() {
					mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
						return nil, domain.ErrNotFound
					}
				})

				It("should return not found error", func() {
					link, err := svc.GetShortLink(ctx, "non-existent")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("resource not found"))
					Expect(link).To(BeNil())
				})
			})
		})

		Describe("GetShortLinkByCode", func() {
			Context("when the short link exists", func() {
				BeforeEach(func() {
					mockShortLinkRepo.GetByCodeFunc = func(ctx context.Context, code string) (*domain.ShortLink, error) {
						return &domain.ShortLink{
							ID:        "link-123",
							Code:      code,
							URLID:     "url-123",
							IsActive:  true,
							CreatedAt: time.Now(),
						}, nil
					}

					mockURLRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.URL, error) {
						return &domain.URL{
							ID:          id,
							OriginalURL: "https://example.com",
							CreatedAt:   time.Now(),
						}, nil
					}
				})

				It("should return the short link with URL details", func() {
					link, err := svc.GetShortLinkByCode(ctx, "abc123")

					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
					Expect(link.Code).To(Equal("abc123"))
					Expect(link.URL).NotTo(BeNil())
					Expect(link.URL.OriginalURL).To(Equal("https://example.com"))
				})
			})

			Context("when the short link doesn't exist", func() {
				BeforeEach(func() {
					mockShortLinkRepo.GetByCodeFunc = func(ctx context.Context, code string) (*domain.ShortLink, error) {
						return nil, domain.ErrNotFound
					}
				})

				It("should return not found error", func() {
					link, err := svc.GetShortLinkByCode(ctx, "non-existent")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("resource not found"))
					Expect(link).To(BeNil())
				})
			})
		})

		Describe("UpdateShortLink", func() {
			var (
				updateReq *domain.UpdateShortLinkRequest
			)

			BeforeEach(func() {
				updateReq = &domain.UpdateShortLinkRequest{
					CustomAlias: stringPtr("new-alias"),
					IsActive:    boolPtr(true),
				}

				mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
					return &domain.ShortLink{
						ID:          id,
						Code:        "old-code",
						CustomAlias: stringPtr("old-alias"),
						URLID:       "url-123",
						IsActive:    true,
						CreatedAt:   time.Now(),
					}, nil
				}

				mockShortLinkRepo.UpdateFunc = func(ctx context.Context, link *domain.ShortLink) error {
					return nil
				}

				mockURLRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.URL, error) {
					return &domain.URL{
						ID:          "url-123",
						OriginalURL: "https://example.com",
						CreatedAt:   time.Now(),
					}, nil
				}
			})

			Context("when updating a short link successfully", func() {
				It("should update the short link", func() {
					link, err := svc.UpdateShortLink(ctx, "link-123", updateReq)

					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
					Expect(link.CustomAlias).NotTo(BeNil())
					Expect(*link.CustomAlias).To(Equal("new-alias"))
					Expect(link.IsActive).To(BeTrue())
				})
			})

			Context("when the short link doesn't exist", func() {
				BeforeEach(func() {
					mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
						return nil, domain.ErrNotFound
					}
				})

				It("should return not found error", func() {
					link, err := svc.UpdateShortLink(ctx, "non-existent", updateReq)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("resource not found"))
					Expect(link).To(BeNil())
				})
			})
		})

		Describe("DeleteShortLink", func() {
			Context("when deleting a short link successfully", func() {
				BeforeEach(func() {
					mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
						return &domain.ShortLink{
							ID:        id,
							Code:      "abc123",
							URLID:     "url-123",
							IsActive:  true,
							CreatedAt: time.Now(),
						}, nil
					}

					mockShortLinkRepo.DeleteFunc = func(ctx context.Context, id string) error {
						return nil
					}
				})

				It("should delete the short link", func() {
					err := svc.DeleteShortLink(ctx, "link-123")

					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the short link doesn't exist", func() {
				BeforeEach(func() {
					mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
						return nil, domain.ErrNotFound
					}

					mockShortLinkRepo.DeleteFunc = func(ctx context.Context, id string) error {
						return domain.ErrNotFound
					}
				})

				It("should return not found error", func() {
					err := svc.DeleteShortLink(ctx, "non-existent")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("resource not found"))
				})
			})
		})

		Describe("ListShortLinks", func() {
			Context("when listing short links successfully", func() {
				BeforeEach(func() {
					mockShortLinkRepo.CountFunc = func(ctx context.Context) (int, error) {
						return 2, nil
					}

					mockShortLinkRepo.ListFunc = func(ctx context.Context, offset, limit int) ([]*domain.ShortLink, error) {
						links := []*domain.ShortLink{
							{
								ID:        "link-1",
								Code:      "abc123",
								URLID:     "url-1",
								IsActive:  true,
								CreatedAt: time.Now(),
							},
							{
								ID:        "link-2",
								Code:      "def456",
								URLID:     "url-2",
								IsActive:  true,
								CreatedAt: time.Now(),
							},
						}
						return links, nil
					}
				})

				It("should return the list of short links", func() {
					links, total, err := svc.ListShortLinks(ctx, 1, 10)

					Expect(err).NotTo(HaveOccurred())
					Expect(links).To(HaveLen(2))
					Expect(total).To(Equal(2))
					Expect(links[0].Code).To(Equal("abc123"))
					Expect(links[1].Code).To(Equal("def456"))
				})
			})

			Context("when there's an error listing short links", func() {
				BeforeEach(func() {
					mockShortLinkRepo.CountFunc = func(ctx context.Context) (int, error) {
						return 0, errors.New("database error")
					}
				})

				It("should return the error", func() {
					links, total, err := svc.ListShortLinks(ctx, 1, 10)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("counting short links"))
					Expect(links).To(BeNil())
					Expect(total).To(Equal(0))
				})
			})

			Context("when there's an error getting the links", func() {
				BeforeEach(func() {
					mockShortLinkRepo.CountFunc = func(ctx context.Context) (int, error) {
						return 2, nil
					}

					mockShortLinkRepo.ListFunc = func(ctx context.Context, offset, limit int) ([]*domain.ShortLink, error) {
						return nil, errors.New("database error")
					}
				})

				It("should return the error", func() {
					links, total, err := svc.ListShortLinks(ctx, 1, 10)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("listing short links"))
					Expect(links).To(BeNil())
					Expect(total).To(Equal(0))
				})
			})
		})

		Describe("RecordClick", func() {
			Context("when recording a click successfully", func() {
				BeforeEach(func() {
					mockClickRepo.CreateFunc = func(ctx context.Context, click *domain.LinkClick) error {
						return nil
					}
				})

				It("should record the click with all fields", func() {
					err := svc.RecordClick(ctx, "link-123", "https://referrer.com", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "127.0.0.1")

					Expect(err).NotTo(HaveOccurred())
					// Since click recording is asynchronous, we need to wait a bit
					time.Sleep(100 * time.Millisecond)
				})

				It("should handle empty optional fields", func() {
					err := svc.RecordClick(ctx, "link-123", "", "", "")

					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
				})
			})
		})

		Describe("GetLinkStats", func() {
			Context("when getting stats successfully", func() {
				BeforeEach(func() {
					now := time.Now()
					mockClickRepo.GetStatsByShortLinkIDFunc = func(ctx context.Context, shortLinkID string) (*domain.LinkStats, error) {
						return &domain.LinkStats{
							TotalClicks: 100,
							LastClicked: &now,
							TopReferrers: map[string]int{
								"https://google.com":  30,
								"https://twitter.com": 20,
							},
							TopBrowsers: map[string]int{
								"Chrome":  40,
								"Firefox": 30,
							},
							TopOS: map[string]int{
								"Windows": 45,
								"macOS":   35,
							},
							TopDevices: map[string]int{
								"Desktop": 60,
								"Mobile":  40,
							},
							ClicksByDay: map[string]int{
								"2024-03-09": 50,
								"2024-03-08": 50,
							},
							RecentClicks: []domain.LinkClick{
								{
									ID:          "click-1",
									ShortLinkID: shortLinkID,
									CreatedAt:   now,
								},
								{
									ID:          "click-2",
									ShortLinkID: shortLinkID,
									CreatedAt:   now.Add(-time.Hour),
								},
							},
						}, nil
					}
				})

				It("should return link statistics", func() {
					stats, err := svc.GetLinkStats(ctx, "link-123")

					Expect(err).NotTo(HaveOccurred())
					Expect(stats).NotTo(BeNil())
					Expect(stats.TotalClicks).To(Equal(100))
					Expect(stats.LastClicked).NotTo(BeNil())
					Expect(stats.TopReferrers).To(HaveLen(2))
					Expect(stats.TopBrowsers).To(HaveLen(2))
					Expect(stats.TopOS).To(HaveLen(2))
					Expect(stats.TopDevices).To(HaveLen(2))
					Expect(stats.ClicksByDay).To(HaveLen(2))
					Expect(stats.RecentClicks).To(HaveLen(2))
				})
			})

			Context("when there's an error getting stats", func() {
				BeforeEach(func() {
					mockClickRepo.GetStatsByShortLinkIDFunc = func(ctx context.Context, shortLinkID string) (*domain.LinkStats, error) {
						return nil, errors.New("database error")
					}
				})

				It("should return the error", func() {
					stats, err := svc.GetLinkStats(ctx, "link-123")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("database error"))
					Expect(stats).To(BeNil())
				})
			})
		})

		Describe("URL validation through CreateShortLink", func() {
			Context("when validating URLs", func() {
				It("should accept valid HTTP URLs", func() {
					req := &domain.CreateShortLinkRequest{URL: "http://example.com"}
					link, err := svc.CreateShortLink(ctx, req)
					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
				})

				It("should accept valid HTTPS URLs", func() {
					req := &domain.CreateShortLinkRequest{URL: "https://example.com"}
					link, err := svc.CreateShortLink(ctx, req)
					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
				})

				It("should reject empty URLs", func() {
					req := &domain.CreateShortLinkRequest{URL: ""}
					link, err := svc.CreateShortLink(ctx, req)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("URL cannot be empty"))
					Expect(link).To(BeNil())
				})

				It("should reject invalid URL format", func() {
					req := &domain.CreateShortLinkRequest{URL: "not-a-url"}
					link, err := svc.CreateShortLink(ctx, req)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("must use HTTP or HTTPS protocol"))
					Expect(link).To(BeNil())
				})

				It("should reject non-HTTP/HTTPS protocols", func() {
					req := &domain.CreateShortLinkRequest{URL: "ftp://example.com"}
					link, err := svc.CreateShortLink(ctx, req)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("must use HTTP or HTTPS"))
					Expect(link).To(BeNil())
				})

				It("should reject URLs without host", func() {
					req := &domain.CreateShortLinkRequest{URL: "https:///path"}
					link, err := svc.CreateShortLink(ctx, req)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("must have a host"))
					Expect(link).To(BeNil())
				})
			})
		})

		Describe("User agent parsing through RecordClick", func() {
			var capturedClick *domain.LinkClick

			BeforeEach(func() {
				mockClickRepo.CreateFunc = func(ctx context.Context, click *domain.LinkClick) error {
					capturedClick = click
					return nil
				}
			})

			Context("when parsing browser information", func() {
				It("should detect Chrome", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond) // Wait for async processing
					Expect(capturedClick.Browser).NotTo(BeNil())
					Expect(*capturedClick.Browser).To(Equal("Chrome"))
				})

				It("should detect Firefox", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.Browser).NotTo(BeNil())
					Expect(*capturedClick.Browser).To(Equal("Firefox"))
				})

				It("should detect Safari", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.Browser).NotTo(BeNil())
					Expect(*capturedClick.Browser).To(Equal("Safari"))
				})

				It("should detect Edge", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.Browser).NotTo(BeNil())
					Expect(*capturedClick.Browser).To(Equal("Edge"))
				})

				It("should detect Opera", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 OPR/77.0.4054.277", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.Browser).NotTo(BeNil())
					Expect(*capturedClick.Browser).To(Equal("Opera"))
				})

				It("should mark unknown browsers as Other", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Unknown Browser", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.Browser).NotTo(BeNil())
					Expect(*capturedClick.Browser).To(Equal("Other"))
				})
			})

			Context("when parsing OS information", func() {
				It("should detect Windows", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.OS).NotTo(BeNil())
					Expect(*capturedClick.OS).To(Equal("Windows"))
				})

				It("should detect macOS", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.OS).NotTo(BeNil())
					Expect(*capturedClick.OS).To(Equal("macOS"))
				})

				It("should detect Linux", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.OS).NotTo(BeNil())
					Expect(*capturedClick.OS).To(Equal("Linux"))
				})

				It("should detect Android", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Mobile Safari/537.36", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.OS).NotTo(BeNil())
					Expect(*capturedClick.OS).To(Equal("Android"))
				})

				It("should detect iOS", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.OS).NotTo(BeNil())
					Expect(*capturedClick.OS).To(Equal("iOS"))
				})

				It("should mark unknown OS as Other", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Unknown OS", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.OS).NotTo(BeNil())
					Expect(*capturedClick.OS).To(Equal("Other"))
				})
			})

			Context("when parsing device information", func() {
				It("should detect mobile devices", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.Device).NotTo(BeNil())
					Expect(*capturedClick.Device).To(Equal("Mobile"))
				})

				It("should detect tablets", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.Device).NotTo(BeNil())
					Expect(*capturedClick.Device).To(Equal("Tablet"))
				})

				It("should mark other devices as Desktop", func() {
					err := svc.RecordClick(ctx, "link-123", "", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "")
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(100 * time.Millisecond)
					Expect(capturedClick.Device).NotTo(BeNil())
					Expect(*capturedClick.Device).To(Equal("Desktop"))
				})
			})
		})
	})

	// CachedURLShortenerService tests
	Describe("CachedURLShortenerService", func() {
		var (
			mockURLRepo       *mocks.MockURLRepository
			mockShortLinkRepo *mocks.MockShortLinkRepository
			mockClickRepo     *mocks.MockLinkClickRepository
			mockCache         *mocks.MockCache
			logger            *zap.Logger
			baseService       *service.URLShortenerService
			svc               *service.CachedURLShortenerService
			ctx               context.Context
		)

		BeforeEach(func() {
			mockURLRepo = &mocks.MockURLRepository{}
			mockShortLinkRepo = &mocks.MockShortLinkRepository{}
			mockClickRepo = &mocks.MockLinkClickRepository{}
			mockCache = &mocks.MockCache{}
			logger = zaptest.NewLogger(GinkgoT())
			ctx = context.Background()

			baseService = service.NewURLShortenerService(
				mockURLRepo,
				mockShortLinkRepo,
				mockClickRepo,
				logger,
				"https://short.example.com",
				30*24*time.Hour,
			)

			svc = service.NewCachedURLShortenerService(baseService, mockCache, logger)
		})

		Describe("CreateShortLink", func() {
			var (
				req *domain.CreateShortLinkRequest
			)

			BeforeEach(func() {
				req = &domain.CreateShortLinkRequest{
					URL: "https://example.com/some-long-url",
				}

				mockURLRepo.GetByHashFunc = func(ctx context.Context, hash string) (*domain.URL, error) {
					return nil, errors.New("not found")
				}

				mockURLRepo.CreateFunc = func(ctx context.Context, url *domain.URL) error {
					return nil
				}

				mockShortLinkRepo.GetByCodeFunc = func(ctx context.Context, code string) (*domain.ShortLink, error) {
					return nil, errors.New("not found")
				}

				mockShortLinkRepo.CreateFunc = func(ctx context.Context, link *domain.ShortLink) error {
					return nil
				}

				mockURLRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.URL, error) {
					return &domain.URL{
						ID:          id,
						OriginalURL: req.URL,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil
				}
			})

			Context("when creating a short link successfully", func() {
				var capturedCacheKey string
				var capturedCacheValue interface{}

				BeforeEach(func() {
					mockCache.SetFunc = func(key string, value interface{}, ttl int) {
						capturedCacheKey = key
						capturedCacheValue = value
					}
				})

				It("should create the link and cache it", func() {
					link, err := svc.CreateShortLink(ctx, req)

					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
					Expect(capturedCacheKey).To(Equal(link.Code))
					Expect(capturedCacheValue).To(Equal(link))
				})
			})

			Context("when there's an error creating the link", func() {
				BeforeEach(func() {
					mockShortLinkRepo.CreateFunc = func(ctx context.Context, link *domain.ShortLink) error {
						return errors.New("database error")
					}
				})

				It("should not cache anything and return the error", func() {
					var cacheWasSet bool
					mockCache.SetFunc = func(key string, value interface{}, ttl int) {
						cacheWasSet = true
					}

					link, err := svc.CreateShortLink(ctx, req)

					Expect(err).To(HaveOccurred())
					Expect(link).To(BeNil())
					Expect(cacheWasSet).To(BeFalse())
				})
			})
		})

		Describe("GetShortLink", func() {
			Context("when the link is in cache", func() {
				var cachedLink *domain.ShortLink

				BeforeEach(func() {
					cachedLink = &domain.ShortLink{
						ID:        "cached-id",
						Code:      "cached-code",
						URLID:     "url-123",
						IsActive:  true,
						CreatedAt: time.Now(),
					}

					mockCache.GetFunc = func(key string) (interface{}, bool) {
						if key == "id:cached-id" {
							return cachedLink, true
						}
						return nil, false
					}
				})

				It("should return the cached link without hitting the database", func() {
					var dbWasHit bool
					mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
						dbWasHit = true
						return nil, errors.New("should not be called")
					}

					link, err := svc.GetShortLink(ctx, "cached-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(link).To(Equal(cachedLink))
					Expect(dbWasHit).To(BeFalse())
				})
			})

			Context("when the link is not in cache", func() {
				var dbLink *domain.ShortLink
				var capturedCacheKeys []string
				var capturedCacheValues []interface{}

				BeforeEach(func() {
					dbLink = &domain.ShortLink{
						ID:        "db-id",
						Code:      "db-code",
						URLID:     "url-123",
						IsActive:  true,
						CreatedAt: time.Now(),
					}

					mockCache.GetFunc = func(key string) (interface{}, bool) {
						return nil, false
					}

					mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
						return dbLink, nil
					}

					capturedCacheKeys = nil
					capturedCacheValues = nil
					mockCache.SetFunc = func(key string, value interface{}, ttl int) {
						capturedCacheKeys = append(capturedCacheKeys, key)
						capturedCacheValues = append(capturedCacheValues, value)
					}
				})

				It("should fetch from database and cache the result", func() {
					link, err := svc.GetShortLink(ctx, "db-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(link).To(Equal(dbLink))
					Expect(capturedCacheKeys).To(ContainElements("id:db-id", "db-code"))
					Expect(capturedCacheValues).To(ContainElements(dbLink, dbLink))
				})
			})
		})

		Describe("GetShortLinkByCode", func() {
			Context("when the link is in cache", func() {
				var cachedLink *domain.ShortLink

				BeforeEach(func() {
					cachedLink = &domain.ShortLink{
						ID:        "cached-id",
						Code:      "cached-code",
						URLID:     "url-123",
						IsActive:  true,
						CreatedAt: time.Now(),
					}

					mockCache.GetFunc = func(key string) (interface{}, bool) {
						if key == "cached-code" {
							return cachedLink, true
						}
						return nil, false
					}
				})

				It("should return the cached link without hitting the database", func() {
					var dbWasHit bool
					mockShortLinkRepo.GetByCodeFunc = func(ctx context.Context, code string) (*domain.ShortLink, error) {
						dbWasHit = true
						return nil, errors.New("should not be called")
					}

					link, err := svc.GetShortLinkByCode(ctx, "cached-code")

					Expect(err).NotTo(HaveOccurred())
					Expect(link).To(Equal(cachedLink))
					Expect(dbWasHit).To(BeFalse())
				})
			})

			Context("when the link is not in cache", func() {
				var dbLink *domain.ShortLink
				var capturedCacheKeys []string
				var capturedCacheValues []interface{}

				BeforeEach(func() {
					dbLink = &domain.ShortLink{
						ID:        "db-id",
						Code:      "db-code",
						URLID:     "url-123",
						IsActive:  true,
						CreatedAt: time.Now(),
					}

					mockCache.GetFunc = func(key string) (interface{}, bool) {
						return nil, false
					}

					mockShortLinkRepo.GetByCodeFunc = func(ctx context.Context, code string) (*domain.ShortLink, error) {
						return dbLink, nil
					}

					capturedCacheKeys = nil
					capturedCacheValues = nil
					mockCache.SetFunc = func(key string, value interface{}, ttl int) {
						capturedCacheKeys = append(capturedCacheKeys, key)
						capturedCacheValues = append(capturedCacheValues, value)
					}
				})

				It("should fetch from database and cache the result", func() {
					link, err := svc.GetShortLinkByCode(ctx, "db-code")

					Expect(err).NotTo(HaveOccurred())
					Expect(link).To(Equal(dbLink))
					Expect(capturedCacheKeys).To(ContainElements("db-code", "id:db-id"))
					Expect(capturedCacheValues).To(ContainElements(dbLink, dbLink))
				})
			})
		})

		Describe("UpdateShortLink", func() {
			var (
				updateReq *domain.UpdateShortLinkRequest
				oldLink   *domain.ShortLink
				newLink   *domain.ShortLink
			)

			BeforeEach(func() {
				customAlias := "new-alias"
				updateReq = &domain.UpdateShortLinkRequest{
					CustomAlias: &customAlias,
				}

				oldLink = &domain.ShortLink{
					ID:        "link-123",
					Code:      "old-code",
					URLID:     "url-123",
					IsActive:  true,
					CreatedAt: time.Now(),
				}

				newLink = &domain.ShortLink{
					ID:          "link-123",
					Code:        "new-code",
					CustomAlias: &customAlias,
					URLID:       "url-123",
					IsActive:    true,
					CreatedAt:   time.Now(),
				}

				mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
					return oldLink, nil
				}

				mockShortLinkRepo.UpdateFunc = func(ctx context.Context, link *domain.ShortLink) error {
					return nil
				}
			})

			Context("when updating successfully", func() {
				var deletedKeys []string
				var setKeys []string
				var setValues []interface{}

				BeforeEach(func() {
					deletedKeys = nil
					setKeys = nil
					setValues = nil

					mockCache.DeleteFunc = func(key string) {
						deletedKeys = append(deletedKeys, key)
					}

					mockCache.SetFunc = func(key string, value interface{}, ttl int) {
						setKeys = append(setKeys, key)
						setValues = append(setValues, value)
					}

					mockShortLinkRepo.UpdateFunc = func(ctx context.Context, link *domain.ShortLink) error {
						return nil
					}

					// After update, return the new link
					mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
						if id == "link-123" {
							return newLink, nil
						}
						return nil, errors.New("not found")
					}
				})

				It("should invalidate old cache entries and set new ones", func() {
					link, err := svc.UpdateShortLink(ctx, "link-123", updateReq)

					Expect(err).NotTo(HaveOccurred())
					Expect(link).NotTo(BeNil())
					Expect(deletedKeys).To(ContainElement("id:link-123"))
					Expect(setKeys).To(ContainElements("id:link-123", "new-code"))
					Expect(setValues).To(ContainElements(newLink, newLink))
				})
			})
		})

		Describe("DeleteShortLink", func() {
			var oldLink *domain.ShortLink

			BeforeEach(func() {
				oldLink = &domain.ShortLink{
					ID:        "link-123",
					Code:      "old-code",
					URLID:     "url-123",
					IsActive:  true,
					CreatedAt: time.Now(),
				}

				mockShortLinkRepo.GetByIDFunc = func(ctx context.Context, id string) (*domain.ShortLink, error) {
					return oldLink, nil
				}

				mockShortLinkRepo.DeleteFunc = func(ctx context.Context, id string) error {
					return nil
				}
			})

			Context("when deleting successfully", func() {
				var deletedKeys []string

				BeforeEach(func() {
					deletedKeys = nil
					mockCache.DeleteFunc = func(key string) {
						deletedKeys = append(deletedKeys, key)
					}
				})

				It("should invalidate cache entries", func() {
					err := svc.DeleteShortLink(ctx, "link-123")

					Expect(err).NotTo(HaveOccurred())
					Expect(deletedKeys).To(ContainElements("old-code", "id:link-123"))
				})
			})
		})

		Describe("ListShortLinks", func() {
			Context("when listing links", func() {
				var dbLinks []*domain.ShortLink

				BeforeEach(func() {
					dbLinks = []*domain.ShortLink{
						{
							ID:        "link-1",
							Code:      "code-1",
							URLID:     "url-1",
							IsActive:  true,
							CreatedAt: time.Now(),
						},
						{
							ID:        "link-2",
							Code:      "code-2",
							URLID:     "url-2",
							IsActive:  true,
							CreatedAt: time.Now(),
						},
					}

					mockShortLinkRepo.ListFunc = func(ctx context.Context, offset, limit int) ([]*domain.ShortLink, error) {
						return dbLinks, nil
					}

					mockShortLinkRepo.CountFunc = func(ctx context.Context) (int, error) {
						return len(dbLinks), nil
					}
				})

				It("should bypass cache and return results directly", func() {
					var cacheWasUsed bool
					mockCache.GetFunc = func(key string) (interface{}, bool) {
						cacheWasUsed = true
						return nil, false
					}

					links, total, err := svc.ListShortLinks(ctx, 1, 10)

					Expect(err).NotTo(HaveOccurred())
					Expect(links).To(Equal(dbLinks))
					Expect(total).To(Equal(2))
					Expect(cacheWasUsed).To(BeFalse())
				})
			})
		})

		Describe("RecordClick", func() {
			Context("when recording a click", func() {
				BeforeEach(func() {
					mockClickRepo.CreateFunc = func(ctx context.Context, click *domain.LinkClick) error {
						return nil
					}
				})

				It("should bypass cache and delegate to base service", func() {
					var cacheWasUsed bool
					mockCache.GetFunc = func(key string) (interface{}, bool) {
						cacheWasUsed = true
						return nil, false
					}

					err := svc.RecordClick(ctx, "link-123", "referrer", "user-agent", "127.0.0.1")

					Expect(err).NotTo(HaveOccurred())
					Expect(cacheWasUsed).To(BeFalse())
				})
			})
		})

		Describe("GetLinkStats", func() {
			Context("when getting stats", func() {
				var stats *domain.LinkStats

				BeforeEach(func() {
					stats = &domain.LinkStats{
						TotalClicks: 100,
						TopReferrers: map[string]int{
							"https://google.com": 30,
						},
					}

					mockClickRepo.GetStatsByShortLinkIDFunc = func(ctx context.Context, shortLinkID string) (*domain.LinkStats, error) {
						return stats, nil
					}
				})

				It("should bypass cache and delegate to base service", func() {
					var cacheWasUsed bool
					mockCache.GetFunc = func(key string) (interface{}, bool) {
						cacheWasUsed = true
						return nil, false
					}

					result, err := svc.GetLinkStats(ctx, "link-123")

					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(stats))
					Expect(cacheWasUsed).To(BeFalse())
				})
			})
		})

		Describe("GetCacheStats", func() {
			Context("when getting cache stats", func() {
				BeforeEach(func() {
					mockCache.GetStatsFunc = func() cache.Stats {
						return cache.Stats{
							Size:    100,
							Hits:    50,
							Misses:  25,
							Evicted: 10,
						}
					}
				})

				It("should return cache statistics", func() {
					stats := svc.GetCacheStats()

					Expect(stats.Size).To(Equal(100))
					Expect(stats.Hits).To(Equal(50))
					Expect(stats.Misses).To(Equal(25))
					Expect(stats.Evicted).To(Equal(10))
				})
			})
		})
	})
})

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
