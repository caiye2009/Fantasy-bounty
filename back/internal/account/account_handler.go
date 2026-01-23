package account

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateAccount 创建账号
func (h *Handler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AccountResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	account, err := h.service.CreateAccount(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AccountResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, AccountResponse{
		Code:    http.StatusCreated,
		Message: "Account created successfully",
		Data:    account,
	})
}

// GetAccount 获取账号
func (h *Handler) GetAccount(c *gin.Context) {
	id := c.Param("id")

	account, err := h.service.GetAccount(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "account not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, AccountResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AccountResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    account,
	})
}

// UpdateAccount 更新账号
func (h *Handler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AccountResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	account, err := h.service.UpdateAccount(c.Request.Context(), id, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "account not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, AccountResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AccountResponse{
		Code:    http.StatusOK,
		Message: "Account updated successfully",
		Data:    account,
	})
}

// DeleteAccount 删除账号
func (h *Handler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteAccount(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "account not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, AccountResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AccountResponse{
		Code:    http.StatusOK,
		Message: "Account deleted successfully",
	})
}

// ListAccounts 获取账号列表
func (h *Handler) ListAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	accounts, total, err := h.service.ListAccounts(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AccountListResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AccountListResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    accounts,
		Total:   total,
	})
}
