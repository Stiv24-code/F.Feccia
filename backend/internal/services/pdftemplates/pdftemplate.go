// Package pdftemplates ports OrderMesh's per-client PDF template store to
// GORM: each template describes how to read one client's PDF order layout
// (normalized zones onto inbound-order fields) and which mail senders it is
// preselected for.
package pdftemplates

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

type PdfTemplateService struct {
	db *gorm.DB
}

func NewPdfTemplateService(db *gorm.DB) *PdfTemplateService {
	return &PdfTemplateService{db: db}
}

func (s *PdfTemplateService) List(ctx context.Context) ([]dto.PdfTemplateResponse, error) {
	var items []models.PdfTemplate
	if err := s.db.WithContext(ctx).Order("name ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]dto.PdfTemplateResponse, 0, len(items))
	for _, t := range items {
		resp, err := toResponse(t)
		if err != nil {
			return nil, err
		}
		out = append(out, resp)
	}
	return out, nil
}

func (s *PdfTemplateService) GetByID(ctx context.Context, id uuid.UUID) (*dto.PdfTemplateResponse, error) {
	var t models.PdfTemplate
	if err := s.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	resp, err := toResponse(t)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *PdfTemplateService) Create(ctx context.Context, req dto.PdfTemplateRequest) (*dto.PdfTemplateResponse, error) {
	if err := normalize(&req); err != nil {
		return nil, err
	}
	t := models.PdfTemplate{
		ID:     uuid.New(),
		Name:   req.Name,
		Client: req.Client,
	}
	if err := marshalInto(&t, req); err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Create(&t).Error; err != nil {
		return nil, err
	}
	resp, err := toResponse(t)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *PdfTemplateService) Update(ctx context.Context, id uuid.UUID, req dto.PdfTemplateRequest) (*dto.PdfTemplateResponse, error) {
	if err := normalize(&req); err != nil {
		return nil, err
	}
	var t models.PdfTemplate
	if err := s.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	t.Name = req.Name
	t.Client = req.Client
	if err := marshalInto(&t, req); err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Save(&t).Error; err != nil {
		return nil, err
	}
	resp, err := toResponse(t)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *PdfTemplateService) Delete(ctx context.Context, id uuid.UUID) error {
	res := s.db.WithContext(ctx).Delete(&models.PdfTemplate{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Match returns the best template for a sender address — exact address wins
// over "@domain" match — or nil when nothing matches. The scan happens in Go
// (templates are few) so Senders can stay a plain JSON column.
func (s *PdfTemplateService) Match(ctx context.Context, sender string) (*dto.PdfTemplateResponse, error) {
	items, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	var best *dto.PdfTemplateResponse
	bestScore := 0
	for i := range items {
		if score := matchesSender(items[i].Senders, sender); score > bestScore {
			best, bestScore = &items[i], score
		}
	}
	return best, nil
}

// matchesSender reports how well a template's senders match an address:
// 2 = exact address, 1 = domain match, 0 = no match.
func matchesSender(senders []string, addr string) int {
	addr = strings.ToLower(strings.TrimSpace(addr))
	if addr == "" {
		return 0
	}
	domain := ""
	if i := strings.LastIndex(addr, "@"); i >= 0 {
		domain = addr[i:]
	}
	best := 0
	for _, s := range senders {
		switch {
		case s == addr:
			return 2
		case strings.HasPrefix(s, "@") && s == domain:
			best = 1
		}
	}
	return best
}

// normalize trims/lowercases senders, assigns missing field IDs and rejects
// unknown field targets — same rules as OrderMesh's Template.normalize.
func normalize(req *dto.PdfTemplateRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return utils.NewAPIError(400, "il template deve avere un nome")
	}
	clean := make([]string, 0, len(req.Senders))
	for _, s := range req.Senders {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" {
			clean = append(clean, s)
		}
	}
	req.Senders = clean
	if req.Fields == nil {
		req.Fields = []dto.PdfTemplateFieldDTO{}
	}
	valid := map[string]bool{}
	for _, k := range models.InboundOrderFieldTargets {
		valid[k] = true
	}
	for i := range req.Fields {
		f := &req.Fields[i]
		if f.ID == "" {
			f.ID = newFieldID()
		}
		if !valid[f.Target] {
			return utils.NewAPIError(400, "campo con destinazione sconosciuta: "+f.Target)
		}
	}
	return nil
}

func newFieldID() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	return "fld-" + hex.EncodeToString(b)
}

func marshalInto(t *models.PdfTemplate, req dto.PdfTemplateRequest) error {
	senders, err := json.Marshal(req.Senders)
	if err != nil {
		return err
	}
	fields, err := json.Marshal(req.Fields)
	if err != nil {
		return err
	}
	t.Senders = senders
	t.Fields = fields
	return nil
}

func toResponse(t models.PdfTemplate) (dto.PdfTemplateResponse, error) {
	resp := dto.PdfTemplateResponse{
		ID:        t.ID,
		Name:      t.Name,
		Client:    t.Client,
		Senders:   []string{},
		Fields:    []dto.PdfTemplateFieldDTO{},
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	if len(t.Senders) > 0 {
		if err := json.Unmarshal(t.Senders, &resp.Senders); err != nil {
			return resp, errors.New("template " + t.ID.String() + ": senders non validi: " + err.Error())
		}
	}
	if len(t.Fields) > 0 {
		if err := json.Unmarshal(t.Fields, &resp.Fields); err != nil {
			return resp, errors.New("template " + t.ID.String() + ": fields non validi: " + err.Error())
		}
	}
	return resp, nil
}
