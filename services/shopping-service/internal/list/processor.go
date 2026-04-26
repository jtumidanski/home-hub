package list

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/categoryclient"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/item"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/recipeclient"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound       = errors.New("shopping list not found")
	ErrAlreadyArchived = errors.New("shopping list is already archived")
	ErrNotArchived    = errors.New("shopping list is not archived")
	ErrArchived       = errors.New("shopping list is archived")
)

type Processor struct {
	l            logrus.FieldLogger
	ctx          context.Context
	db           *gorm.DB
	catClient    *categoryclient.Client
	recipeClient *recipeclient.Client
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) *Processor {
	return &Processor{l: l, ctx: ctx, db: db, catClient: catClient, recipeClient: recipeClient}
}

// tenantHeaders returns the (tenantID, householdID) pair from the request
// context for forwarding to downstream services. The auth-service issues
// JWTs with nil tenant/household claims and the shared auth middleware
// falls back to X-Tenant-ID / X-Household-ID headers — outbound calls to
// category-service and recipe-service must therefore pass these explicitly
// or those services will resolve the request as the nil tenant.
func (p *Processor) tenantHeaders() (uuid.UUID, uuid.UUID) {
	if t, ok := tenantctx.FromContext(p.ctx); ok {
		return t.Id(), t.HouseholdId()
	}
	return uuid.Nil, uuid.Nil
}

func (p *Processor) List(status string) ([]Model, error) {
	if status == "" {
		status = "active"
	}
	entities, err := GetByStatus(status)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}

	listIDs := make([]uuid.UUID, len(entities))
	for i, e := range entities {
		listIDs[i] = e.Id
	}
	counts, err := getItemCounts(p.db.WithContext(p.ctx), listIDs)
	if err != nil {
		return nil, err
	}

	models := make([]Model, len(entities))
	for i, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		if c, ok := counts[e.Id]; ok {
			m = m.WithCounts(c.ItemCount, c.CheckedCount)
		}
		models[i] = m
	}
	return models, nil
}

func (p *Processor) Get(id uuid.UUID) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	itemCount, checkedCount, _ := getItemCountsForList(p.db.WithContext(p.ctx), id)
	return m.WithCounts(itemCount, checkedCount), nil
}

func (p *Processor) Create(tenantID, householdID, userID uuid.UUID, name string) (Model, error) {
	name = strings.TrimSpace(name)
	if _, err := NewBuilder().SetName(name).Build(); err != nil {
		return Model{}, err
	}

	e := Entity{
		TenantId:    tenantID,
		HouseholdId: householdID,
		Name:        name,
		Status:      "active",
		CreatedBy:   userID,
	}
	if err := createList(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Update(id uuid.UUID, name string) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.Status == "archived" {
		return Model{}, ErrArchived
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Model{}, ErrNameRequired
	}
	if len(name) > 255 {
		return Model{}, ErrNameTooLong
	}
	e.Name = name

	if err := updateList(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	itemCount, checkedCount, _ := getItemCountsForList(p.db.WithContext(p.ctx), id)
	return m.WithCounts(itemCount, checkedCount), nil
}

func (p *Processor) Delete(id uuid.UUID) error {
	if _, err := GetByID(id)(p.db.WithContext(p.ctx))(); err != nil {
		return ErrNotFound
	}
	return deleteList(p.db.WithContext(p.ctx), id)
}

func (p *Processor) Archive(id uuid.UUID) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.Status == "archived" {
		return Model{}, ErrAlreadyArchived
	}

	now := time.Now().UTC()
	e.Status = "archived"
	e.ArchivedAt = &now

	if err := updateList(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	itemCount, checkedCount, _ := getItemCountsForList(p.db.WithContext(p.ctx), id)
	return m.WithCounts(itemCount, checkedCount), nil
}

func (p *Processor) validateListModifiable(listID uuid.UUID) error {
	m, err := p.Get(listID)
	if err != nil {
		return ErrNotFound
	}
	if m.IsArchived() {
		return ErrArchived
	}
	return nil
}

func (p *Processor) GetWithItems(id uuid.UUID) (Model, []item.Model, error) {
	m, err := p.Get(id)
	if err != nil {
		return Model{}, nil, err
	}
	itemProc := item.NewProcessor(p.l, p.ctx, p.db)
	items, err := itemProc.GetByListID(id)
	if err != nil {
		return Model{}, nil, err
	}
	return m, items, nil
}

func (p *Processor) AddItem(listID uuid.UUID, input item.AddInput, accessToken string) (item.Model, error) {
	if err := p.validateListModifiable(listID); err != nil {
		return item.Model{}, err
	}
	input.ListID = listID
	if input.CategoryID == nil {
		p.autoCategorize(&input, accessToken)
	}
	p.enrichCategory(&input, accessToken)
	itemProc := item.NewProcessor(p.l, p.ctx, p.db)
	return itemProc.Add(input)
}

func (p *Processor) UpdateItem(listID uuid.UUID, itemID uuid.UUID, input item.UpdateInput, accessToken string) (item.Model, error) {
	if err := p.validateListModifiable(listID); err != nil {
		return item.Model{}, err
	}
	if input.CategoryID == nil && !input.ClearCategory && input.Name != nil {
		p.autoCategorizeUpdate(&input, accessToken)
	}
	p.enrichUpdateCategory(&input, accessToken)
	itemProc := item.NewProcessor(p.l, p.ctx, p.db)
	return itemProc.Update(itemID, input)
}

func (p *Processor) RemoveItem(listID uuid.UUID, itemID uuid.UUID) error {
	if err := p.validateListModifiable(listID); err != nil {
		return err
	}
	itemProc := item.NewProcessor(p.l, p.ctx, p.db)
	return itemProc.Delete(itemID)
}

func (p *Processor) CheckItem(listID uuid.UUID, itemID uuid.UUID, checked bool) (item.Model, error) {
	if err := p.validateListModifiable(listID); err != nil {
		return item.Model{}, err
	}
	itemProc := item.NewProcessor(p.l, p.ctx, p.db)
	return itemProc.Check(itemID, checked)
}

func (p *Processor) UncheckAllItems(listID uuid.UUID) (Model, []item.Model, error) {
	if err := p.validateListModifiable(listID); err != nil {
		return Model{}, nil, err
	}
	itemProc := item.NewProcessor(p.l, p.ctx, p.db)
	if err := itemProc.UncheckAll(listID); err != nil {
		return Model{}, nil, err
	}
	return p.GetWithItems(listID)
}

func (p *Processor) ImportFromMealPlan(listID uuid.UUID, planID uuid.UUID, accessToken string) (Model, []item.Model, int, error) {
	if err := p.validateListModifiable(listID); err != nil {
		return Model{}, nil, 0, err
	}

	tenantID, householdID := p.tenantHeaders()

	ingredients, err := p.recipeClient.GetPlanIngredients(planID, accessToken, tenantID, householdID)
	if err != nil {
		return Model{}, nil, 0, err
	}

	categoryMap := make(map[string]categoryclient.Category)
	if p.catClient != nil {
		cats, catErr := p.catClient.ListCategories(accessToken, tenantID, householdID)
		if catErr != nil {
			p.l.WithError(catErr).WithFields(logrus.Fields{
				"plan_id": planID,
				"list_id": listID,
			}).Error("Failed to fetch categories for meal plan import; imported items will be uncategorized")
		} else {
			for _, cat := range cats {
				categoryMap[cat.Name] = cat
			}
		}
	}

	itemProc := item.NewProcessor(p.l, p.ctx, p.db)
	addedCount := 0
	for _, ing := range ingredients {
		qtyStr := recipeclient.FormatQuantityString(ing.Quantity, ing.Unit)

		addInput := item.AddInput{
			ListID: listID,
			Name:   ing.Name,
		}
		if qtyStr != "" {
			addInput.Quantity = &qtyStr
		}

		if ing.CategoryName != "" {
			if cat, ok := categoryMap[ing.CategoryName]; ok {
				addInput.CategoryID = &cat.ID
				addInput.CategoryName = &cat.Name
				addInput.CategorySortOrder = &cat.SortOrder
			}
		}

		if _, err := itemProc.Add(addInput); err != nil {
			p.l.WithError(err).Warn("Failed to add imported item")
		} else {
			addedCount++
		}

		for _, eq := range ing.ExtraQuantities {
			eqStr := recipeclient.FormatQuantityString(eq.Quantity, eq.Unit)
			eqInput := item.AddInput{
				ListID:            listID,
				Name:              ing.Name,
				CategoryID:        addInput.CategoryID,
				CategoryName:      addInput.CategoryName,
				CategorySortOrder: addInput.CategorySortOrder,
			}
			if eqStr != "" {
				eqInput.Quantity = &eqStr
			}
			if _, err := itemProc.Add(eqInput); err != nil {
				p.l.WithError(err).Warn("Failed to add imported item")
			} else {
				addedCount++
			}
		}
	}
	m, items, err := p.GetWithItems(listID)
	if err != nil {
		return Model{}, nil, 0, err
	}
	return m, items, addedCount, nil
}

func (p *Processor) Unarchive(id uuid.UUID) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.Status != "archived" {
		return Model{}, ErrNotArchived
	}

	e.Status = "active"
	e.ArchivedAt = nil

	if err := updateList(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	itemCount, checkedCount, _ := getItemCountsForList(p.db.WithContext(p.ctx), id)
	return m.WithCounts(itemCount, checkedCount), nil
}

func (p *Processor) enrichCategory(input *item.AddInput, accessToken string) {
	if input.CategoryID == nil || p.catClient == nil {
		return
	}
	tenantID, householdID := p.tenantHeaders()
	cat, err := p.catClient.GetCategory(*input.CategoryID, accessToken, tenantID, householdID)
	if err != nil {
		p.l.WithError(err).WithFields(logrus.Fields{
			"category_id": *input.CategoryID,
			"item_name":   input.Name,
		}).Warn("Failed to enrich shopping item category; item will store category_id without name/sort order")
		return
	}
	input.CategoryName = &cat.Name
	input.CategorySortOrder = &cat.SortOrder
}

func (p *Processor) enrichUpdateCategory(input *item.UpdateInput, accessToken string) {
	if input.CategoryID == nil || p.catClient == nil {
		return
	}
	tenantID, householdID := p.tenantHeaders()
	cat, err := p.catClient.GetCategory(*input.CategoryID, accessToken, tenantID, householdID)
	if err != nil {
		fields := logrus.Fields{"category_id": *input.CategoryID}
		if input.Name != nil {
			fields["item_name"] = *input.Name
		}
		p.l.WithError(err).WithFields(fields).
			Warn("Failed to enrich shopping item category on update; item will store category_id without name/sort order")
		return
	}
	input.CategoryName = &cat.Name
	input.CategorySortOrder = &cat.SortOrder
}

// autoCategorize asks recipe-service to resolve the item name to a canonical
// ingredient and copies its category id onto the input. Failures and misses
// are logged at debug level and leave the input untouched (the caller falls
// back to "uncategorized"). enrichCategory then populates the name/sort order
// from category-service.
func (p *Processor) autoCategorize(input *item.AddInput, accessToken string) {
	if p.recipeClient == nil || input.Name == "" {
		return
	}
	tenantID, householdID := p.tenantHeaders()
	lookup, matched, err := p.recipeClient.LookupIngredient(input.Name, accessToken, tenantID, householdID)
	if err != nil {
		p.l.WithError(err).WithField("item_name", input.Name).
			Warn("Recipe ingredient lookup failed; shopping item will be uncategorized")
		return
	}
	if !matched || lookup.CategoryID == nil {
		return
	}
	input.CategoryID = lookup.CategoryID
}

func (p *Processor) autoCategorizeUpdate(input *item.UpdateInput, accessToken string) {
	if p.recipeClient == nil || input.Name == nil || *input.Name == "" {
		return
	}
	tenantID, householdID := p.tenantHeaders()
	lookup, matched, err := p.recipeClient.LookupIngredient(*input.Name, accessToken, tenantID, householdID)
	if err != nil {
		p.l.WithError(err).WithField("item_name", *input.Name).
			Warn("Recipe ingredient lookup failed on update; shopping item will be uncategorized")
		return
	}
	if !matched || lookup.CategoryID == nil {
		return
	}
	input.CategoryID = lookup.CategoryID
}
