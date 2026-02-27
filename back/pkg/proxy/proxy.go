package proxy

import (
	"back/pkg/internal_token"
	"back/pkg/jwt"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// forwardToInternal 通用转发方法：获取内部 token → 组装请求 → POST → 透传响应
func (p *InternalProxy) forwardToInternal(c *gin.Context, code string, pars map[string]interface{}, outPars map[string]interface{}) {
	start := time.Now()
	log.Printf("[PROXY] ========== 转发请求: %s ==========", code)

	// 获取内部 token
	internalToken, err := p.tokenManager.GetToken()
	if err != nil {
		log.Printf("[PROXY] 获取内部token失败: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    http.StatusServiceUnavailable,
			"message": "获取内部系统token失败",
		})
		return
	}

	// 组装请求体
	requestBody := map[string]interface{}{
		"code":    code,
		"pars":    pars,
		"outPars": outPars,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	apiURL := p.targetURL.String() + "/api/Public/GetProcedureDataSet"

	log.Printf("[PROXY]   URL: %s", apiURL)
	log.Printf("[PROXY]   请求体: %s", string(bodyBytes))
	log.Printf("[PROXY]   Token: %s", internalToken)

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[PROXY] 创建请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建内部请求失败",
		})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", internalToken)

	// 打印实际发出的请求头（含 Go 自动添加的）
	log.Printf("[PROXY]   发出请求头:")
	for k, v := range req.Header {
		log.Printf("[PROXY]     %s: %s", k, strings.Join(v, ", "))
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[PROXY] 请求内部系统失败: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    502,
			"message": "内部系统不可达",
		})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[PROXY] 读取响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "读取内部响应失败",
		})
		return
	}

	respStr := string(respBody)
	if len(respStr) > 500 {
		log.Printf("[PROXY]   响应体(前500字符): %s...", respStr[:500])
	} else {
		log.Printf("[PROXY]   响应体: %s", respStr)
	}
	log.Printf("[PROXY] ========== 请求完成, 耗时: %v ==========", time.Since(start))

	c.Data(resp.StatusCode, "application/json", respBody)
}

// BindWeChatHandler godoc
// @Summary      微信端绑定供应商
// @Description  验证供应商身份并绑定 OpenId，成功后 isSucceed=true
// @Tags         proxy
// @Accept       json
// @Produce      json
// @Param        request  body      BindWeChatRequest  true  "绑定参数"
// @Success      200      {object}  BindWeChatResponse
// @Failure      400      {object}  BaseResponse
// @Router       /proxy/bind-wechat [post]
func (p *InternalProxy) BindWeChatHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
			return
		}
		p.forwardToInternal(c, "BC_Customer_BindWeChat", body, map[string]interface{}{
			"intRetVal":  "0",
			"strMessage": "",
		})
	}
}

// GetByWeChatHandler godoc
// @Summary      根据 Openid 获取供应商信息
// @Description  data 有数据=已绑定供应商；data 为空数组=未绑定；isSucceed=false 看 message
// @Tags         proxy
// @Accept       json
// @Produce      json
// @Param        request  body      GetByWeChatRequest  true  "查询参数"
// @Success      200      {object}  GetByWeChatResponse
// @Failure      400      {object}  BaseResponse
// @Router       /proxy/get-by-wechat [post]
func (p *InternalProxy) GetByWeChatHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
			return
		}
		p.forwardToInternal(c, "BC_Customer_GetByWeChat", body, map[string]interface{}{})
	}
}

// InquiryQueryHandler godoc
// @Summary      查询供应商可报价的招标信息
// @Tags         proxy
// @Accept       json
// @Produce      json
// @Param        request  body      InquiryQueryRequest  true  "查询参数"
// @Success      200      {object}  InquiryQueryResponse
// @Failure      400      {object}  BaseResponse
// @Router       /proxy/inquiry-query [post]
func (p *InternalProxy) InquiryQueryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
			return
		}
		p.forwardToInternal(c, "Pur_InquiryQueryForSupplier", body, map[string]interface{}{})
	}
}

// InquiryDetailHandler godoc
// @Summary      供应商查看采购需求详情及报价情况
// @Tags         proxy
// @Accept       json
// @Produce      json
// @Param        request  body      InquiryDetailRequest  true  "查询参数"
// @Success      200      {object}  InquiryDetailResponse
// @Failure      400      {object}  BaseResponse
// @Router       /proxy/inquiry-detail [post]
func (p *InternalProxy) InquiryDetailHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
			return
		}
		p.forwardToInternal(c, "Pur_Inquiry_DetailForSupplier", body, map[string]interface{}{
			"strMessage": "",
		})
	}
}

// QuoteDeleteHandler godoc
// @Summary      撤回或删除供应商报价
// @Tags         proxy
// @Accept       json
// @Produce      json
// @Param        request  body      QuoteDeleteRequest  true  "操作参数"
// @Success      200      {object}  QuoteDeleteResponse
// @Failure      400      {object}  BaseResponse
// @Router       /proxy/quote-delete [post]
func (p *InternalProxy) QuoteDeleteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
			return
		}
		p.forwardToInternal(c, "Pur_Inquiry_QuoteDelete", body, map[string]interface{}{
			"strMessage": "",
		})
	}
}

// QuoteSaveHandler godoc
// @Summary      供应商提交或保存报价
// @Tags         proxy
// @Accept       json
// @Produce      json
// @Param        request  body      QuoteSaveRequest  true  "报价参数"
// @Success      200      {object}  QuoteSaveResponse
// @Failure      400      {object}  BaseResponse
// @Router       /proxy/quote-save [post]
func (p *InternalProxy) QuoteSaveHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
			return
		}
		p.forwardToInternal(c, "Pur_Inquiry_QuoteSave", body, map[string]interface{}{
			"ReturnId":   "",
			"strMessage": "",
		})
	}
}

// InternalProxy 内部系统反向代理
// 接收通过外部 JWT 认证的请求，替换为内部 token 后转发到内部系统
type InternalProxy struct {
	tokenManager *internal_token.Manager
	jwtService   *jwt.JWTService
	targetURL    *url.URL
	proxy        *httputil.ReverseProxy
}

// NewInternalProxy 创建内部系统反向代理
func NewInternalProxy(tokenManager *internal_token.Manager, targetURL string, jwtService *jwt.JWTService) *InternalProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		panic(fmt.Sprintf("无效的内部系统 URL: %s", targetURL))
	}

	log.Printf("[PROXY-INIT] 内部系统目标 URL: %s", targetURL)

	p := &InternalProxy{
		tokenManager: tokenManager,
		jwtService:   jwtService,
		targetURL:    target,
	}

	p.proxy = &httputil.ReverseProxy{
		Director: p.director,
		ModifyResponse: func(resp *http.Response) error {
			resp.Header.Del("Server")
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[PROXY-ERROR] 反向代理错误: %v, URL: %s", err, r.URL.String())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, `{"code":502,"message":"internal system unavailable"}`)
		},
	}

	return p
}

// director 重写请求：替换目标 URL 和 Authorization header
func (p *InternalProxy) director(req *http.Request) {
	req.URL.Scheme = p.targetURL.Scheme
	req.URL.Host = p.targetURL.Host
	req.Host = p.targetURL.Host
}

// BountiesHandler 专门处理 /bounties 请求的 Gin handler
// 流程: 前端外部JWT -> JWTAuth中间件验证 -> 交换内部token -> 调内部系统API -> 返回
func (p *InternalProxy) BountiesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestPath := c.Request.URL.Path
		log.Printf("[BOUNTIES] ========== 开始处理请求 ==========")
		log.Printf("[BOUNTIES] 请求路径: %s, 方法: %s", requestPath, c.Request.Method)
		log.Printf("[BOUNTIES] 前端 Authorization: %s", maskToken(c.GetHeader("Authorization")))

		// 从 JWT 中间件设置的 RequestContext 获取用户名（验证外部JWT已通过）
		username := ""
		if rc, exists := c.Get("request_context"); exists {
			log.Printf("[BOUNTIES] RequestContext 存在")
			_ = rc
		}
		// 尝试获取用户信息
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if claims, err := p.jwtService.ValidateToken(tokenStr); err == nil {
				username = claims.Username
				log.Printf("[BOUNTIES] 外部JWT验证成功, 用户: %s", username)
			} else {
				log.Printf("[BOUNTIES] 外部JWT验证失败（中间件应已拦截）: %v", err)
			}
		}

		// 步骤1: 用内部账号密码去交换内部系统的token
		log.Printf("[BOUNTIES] 步骤1: 获取内部系统token（交换）...")
		internalToken, err := p.tokenManager.GetToken()
		if err != nil {
			log.Printf("[BOUNTIES] 获取内部token失败: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    http.StatusServiceUnavailable,
				"message": "获取内部系统token失败",
				"debug":   err.Error(),
			})
			return
		}
		log.Printf("[BOUNTIES] 获取内部token成功: %s", maskToken("Bearer "+internalToken))

		// 步骤2: 用内部token调用内部系统业务接口
		// 检查是否有详情ID（/bounties/:id）
		bountyID := c.Param("id")
		if bountyID != "" {
			log.Printf("[BOUNTIES] 请求详情, InquiryId: %s", bountyID)
			p.handleBountyDetail(c, internalToken, bountyID, username, start)
			return
		}

		log.Printf("[BOUNTIES] 请求列表")
		p.handleBountyList(c, internalToken, username, start)
	}
}

// handleBountyList 处理悬赏列表请求
func (p *InternalProxy) handleBountyList(c *gin.Context, internalToken, supplier string, start time.Time) {
	// 构造请求体
	if supplier == "" {
		supplier = "WBDY" // 默认供应商编码
	}
	pChnName := c.Query("keyword")
	includeEnd := c.DefaultQuery("include_end", "0")
	beginDate := c.Query("begin_date")
	endDate := c.Query("end_date")

	pars := map[string]interface{}{
		"Supplier":   supplier,
		"P_chnName":  pChnName,
		"IncludeEnd": includeEnd,
	}
	if beginDate != "" {
		pars["BeginDate"] = beginDate
	}
	if endDate != "" {
		pars["EndDate"] = endDate
	}

	requestBody := map[string]interface{}{
		"code":    "Pur_InquiryQueryForSupplier",
		"pars":    pars,
		"outPars": map[string]interface{}{},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	apiURL := p.targetURL.String() + "/api/Public/GetProcedureDataSet"

	log.Printf("[BOUNTIES] 步骤2: 调用内部系统业务接口")
	log.Printf("[BOUNTIES]   URL: %s", apiURL)
	log.Printf("[BOUNTIES]   请求体: %s", string(bodyBytes))
	log.Printf("[BOUNTIES]   Authorization: %s", maskToken("Bearer "+internalToken))

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[BOUNTIES] 创建请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建内部请求失败",
			"debug":   err.Error(),
		})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+internalToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[BOUNTIES] 请求内部系统失败: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    502,
			"message": "内部系统不可达",
			"debug":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[BOUNTIES] 读取响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "读取内部响应失败",
		})
		return
	}

	log.Printf("[BOUNTIES] 步骤3: 内部系统响应")
	log.Printf("[BOUNTIES]   HTTP状态码: %d", resp.StatusCode)
	log.Printf("[BOUNTIES]   响应体长度: %d bytes", len(respBody))
	// 打印响应体（截断避免太长）
	respStr := string(respBody)
	if len(respStr) > 500 {
		log.Printf("[BOUNTIES]   响应体(前500字符): %s...", respStr[:500])
	} else {
		log.Printf("[BOUNTIES]   响应体: %s", respStr)
	}

	elapsed := time.Since(start)
	log.Printf("[BOUNTIES] ========== 请求完成, 耗时: %v ==========", elapsed)

	// 直接透传内部系统响应
	c.Data(resp.StatusCode, "application/json", respBody)
}

// handleBountyDetail 处理悬赏详情请求
func (p *InternalProxy) handleBountyDetail(c *gin.Context, internalToken, inquiryID, supplier string, start time.Time) {
	if supplier == "" {
		supplier = "WBDY"
	}

	requestBody := map[string]interface{}{
		"code": "Pur_Inquiry_DetailForSupplier",
		"pars": map[string]interface{}{
			"InquiryId": inquiryID,
			"Supplier":  supplier,
		},
		"outPars": map[string]interface{}{
			"strMessage": "",
		},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	apiURL := p.targetURL.String() + "/api/Public/GetProcedureDataSet"

	log.Printf("[BOUNTIES-DETAIL] 步骤2: 调用内部系统详情接口")
	log.Printf("[BOUNTIES-DETAIL]   URL: %s", apiURL)
	log.Printf("[BOUNTIES-DETAIL]   请求体: %s", string(bodyBytes))

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[BOUNTIES-DETAIL] 创建请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建内部请求失败",
		})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+internalToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[BOUNTIES-DETAIL] 请求内部系统失败: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    502,
			"message": "内部系统不可达",
		})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[BOUNTIES-DETAIL] 读取响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "读取内部响应失败",
		})
		return
	}

	log.Printf("[BOUNTIES-DETAIL] 内部系统响应: HTTP %d, 长度: %d", resp.StatusCode, len(respBody))
	respStr := string(respBody)
	if len(respStr) > 500 {
		log.Printf("[BOUNTIES-DETAIL] 响应体(前500字符): %s...", respStr[:500])
	} else {
		log.Printf("[BOUNTIES-DETAIL] 响应体: %s", respStr)
	}

	elapsed := time.Since(start)
	log.Printf("[BOUNTIES-DETAIL] 请求完成, 耗时: %v", elapsed)

	c.Data(resp.StatusCode, "application/json", respBody)
}

// Handler 返回通用 Gin 处理函数（用于非 bounties 的通配路径）
func (p *InternalProxy) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		targetPath := c.Param("path")
		if targetPath == "" {
			targetPath = "/"
		}

		log.Printf("[PROXY] 通用代理请求: %s -> %s", c.Request.URL.Path, targetPath)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			token, err := p.tokenManager.GetToken()
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"code":    http.StatusServiceUnavailable,
					"message": "failed to authenticate with internal system",
				})
				c.Abort()
				return
			}
			c.Request.URL.Path = targetPath
			c.Request.Header.Set("Authorization", "Bearer "+token)
			c.Request.Header.Del("Cookie")
			p.proxy.ServeHTTP(c.Writer, c.Request)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		_, err := p.jwtService.ValidateToken(tokenString)
		if err == nil {
			internalToken, err := p.tokenManager.GetToken()
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"code":    http.StatusServiceUnavailable,
					"message": "failed to get internal token",
				})
				c.Abort()
				return
			}
			c.Request.URL.Path = targetPath
			c.Request.Header.Set("Authorization", "Bearer "+internalToken)
			c.Request.Header.Del("Cookie")
			p.proxy.ServeHTTP(c.Writer, c.Request)
			return
		}

		c.Request.URL.Path = targetPath
		c.Request.Header.Del("Cookie")
		p.proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// ForceRefreshTokenHandler 手动强制刷新内部 token 的接口
func (p *InternalProxy) ForceRefreshTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[PROXY] 手动触发token刷新，来源IP: %s", c.ClientIP())
		if err := p.tokenManager.ForceRefresh(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "token刷新失败: " + err.Error(),
			})
			return
		}
		expiresAt := p.tokenManager.ExpiresAt()
		log.Printf("[PROXY] 手动刷新token成功，新过期时间: %v", expiresAt.Format(time.RFC3339))
		c.JSON(http.StatusOK, gin.H{
			"code":       200,
			"message":    "token刷新成功",
			"expires_at": expiresAt.Format(time.RFC3339),
		})
	}
}

// maskToken 截断 token 用于安全日志输出
func maskToken(auth string) string {
	if len(auth) <= 20 {
		return auth
	}
	return auth[:20] + "..." + auth[len(auth)-6:]
}
