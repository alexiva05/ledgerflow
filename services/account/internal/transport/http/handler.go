package http

import (
	"context"
	"ledgerflow/services/account/internal/domain"
	pkgerrors "ledgerflow/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type accountService interface {
	CreateAccount(ctx context.Context, owner uuid.UUID, currency string) (*domain.Account, error)
	GetBalance(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error)
}

type Handler struct {
	app accountService
}

func NewHandler(app accountService) *Handler {
	return &Handler{
		app: app,
	}
}

type createAccountRequest struct {
	Owner uuid.UUID `json:"owner"`
	Currency string `json:"currency"`
}

func (h *Handler) CreateAccount(c *gin.Context) {

	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	account, err := h.app.CreateAccount(c.Request.Context(), req.Owner, req.Currency)
	if err != nil {
		c.JSON(pkgerrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, account)
}

type getBalanceResponse struct {
	Balance decimal.Decimal `json:"balance"`
}

func (h *Handler) GetBalance(c *gin.Context) {

	idStr := c.Param("id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	b, err := h.app.GetBalance(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(pkgerrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	balance := getBalanceResponse{
		Balance: b,
	}

	c.JSON(200, balance)
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/accounts", h.CreateAccount)
	r.GET("/accounts/:id/balance", h.GetBalance)
}