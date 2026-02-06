package supplier

import (
	"back/pkg/middleware"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ========== Supplier Handlers ==========

// GetSupplier 获取供应商
func (h *Handler) GetSupplier(c *gin.Context) {
	id := c.Param("id")

	supplier, err := h.service.GetSupplier(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "supplier not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, SupplierResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SupplierResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    supplier,
	})
}

// ListSuppliers 获取供应商列表
func (h *Handler) ListSuppliers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	suppliers, total, err := h.service.ListSuppliers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, SupplierListResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SupplierListResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    suppliers,
		Total:   total,
	})
}

// ========== Application Handlers - 用户操作 ==========

// ApplySupplier 提交供应商认证申请（图片已在OCR识别阶段上传）
func (h *Handler) ApplySupplier(c *gin.Context) {
	// 从RequestContext获取当前用户
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		c.JSON(http.StatusUnauthorized, ApplicationResponse{
			Code:    http.StatusUnauthorized,
			Message: "未登录",
		})
		return
	}
	username := rc.Username

	// 获取JSON请求体
	var body struct {
		Name              string `json:"name"`
		BusinessLicenseNo string `json:"businessLicenseNo"`
		ImagePath         string `json:"imagePath"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ApplicationResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数格式错误",
		})
		return
	}

	if body.Name == "" || body.BusinessLicenseNo == "" {
		c.JSON(http.StatusBadRequest, ApplicationResponse{
			Code:    http.StatusBadRequest,
			Message: "供应商名称和营业执照号不能为空",
		})
		return
	}

	if body.ImagePath == "" {
		c.JSON(http.StatusBadRequest, ApplicationResponse{
			Code:    http.StatusBadRequest,
			Message: "请先上传营业执照图片进行识别",
		})
		return
	}

	// 验证图片文件确实存在
	if _, err := os.Stat(body.ImagePath); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, ApplicationResponse{
			Code:    http.StatusBadRequest,
			Message: "营业执照图片不存在，请重新上传",
		})
		return
	}

	// 创建申请
	req := &ApplySupplierRequest{
		Name:              body.Name,
		BusinessLicenseNo: body.BusinessLicenseNo,
	}

	app, err := h.service.ApplySupplier(c.Request.Context(), username, req, body.ImagePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApplicationResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// 设置审计信息
	rc.Action = "supplier.apply"
	rc.Resource = "supplier"
	rc.Detail = map[string]any{
		"supplier_name": body.Name,
	}

	c.JSON(http.StatusCreated, ApplicationResponse{
		Code:    http.StatusCreated,
		Message: "申请提交成功，请等待审核",
		Data:    app,
	})
}

// RecognizeLicense 上传营业执照图片进行OCR识别
func (h *Handler) RecognizeLicense(c *gin.Context) {
	// 从RequestContext获取当前用户
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		c.JSON(http.StatusUnauthorized, OCRResponse{
			Code:    http.StatusUnauthorized,
			Message: "未登录",
		})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("license")
	if err != nil {
		c.JSON(http.StatusBadRequest, OCRResponse{
			Code:    http.StatusBadRequest,
			Message: "请上传营业执照图片",
		})
		return
	}

	// 验证文件类型（忽略大小写）
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, OCRResponse{
			Code:    http.StatusBadRequest,
			Message: "只支持 jpg、jpeg、png、pdf 格式的文件",
		})
		return
	}

	// 创建上传目录
	uploadDir := "uploads/licenses"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, OCRResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建上传目录失败",
		})
		return
	}

	// 生成唯一文件名并保存
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, OCRResponse{
			Code:    http.StatusInternalServerError,
			Message: "保存文件失败",
		})
		return
	}

	// 调用OCR识别
	result, err := h.service.RecognizeLicense(c.Request.Context(), filePath)
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, OCRResponse{
			Code:    http.StatusInternalServerError,
			Message: "识别失败: " + err.Error(),
		})
		return
	}

	// 设置审计信息
	rc.Action = "supplier.recognize_license"
	rc.Resource = "supplier"

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "识别成功",
		"data":    result,
		"image":   filePath,
	})
}

// GetMySupplierStatus 获取我的供应商认证状态
func (h *Handler) GetMySupplierStatus(c *gin.Context) {
	// 从RequestContext获取当前用户
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		c.JSON(http.StatusUnauthorized, MySupplierStatusResponse{
			Code:    http.StatusUnauthorized,
			Message: "未登录",
		})
		return
	}

	status, err := h.service.GetMySupplierStatus(c.Request.Context(), rc.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MySupplierStatusResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MySupplierStatusResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    status,
	})
}
