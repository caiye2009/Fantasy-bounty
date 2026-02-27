package proxy

// =============================================================================
// 通用响应基础结构
// =============================================================================

// BaseResponse 所有接口的通用响应格式
type BaseResponse struct {
	IsSucceed  bool        `json:"isSucceed"`
	Message    *string     `json:"message"`
	Time       int         `json:"time"`
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
	OutData    interface{} `json:"outData"`
}

// =============================================================================
// 1. 微信端绑定供应商
// =============================================================================

type BindWeChatRequest struct {
	Openid       string `json:"Openid"       example:"oXxx123abc"`
	CustomerCode string `json:"CustomerCode" example:"PD0201"`
	Mobile       string `json:"Mobile"       example:"13800138000"`
	Type         string `json:"Type"         example:"1"` // 0:公众号 1:小程序
}

type BindWeChatOutData struct {
	IntRetVal  int    `json:"intRetVal"`
	StrMessage string `json:"strMessage"`
}

type BindWeChatResponse struct {
	IsSucceed  bool               `json:"isSucceed"`
	Message    *string            `json:"message"`
	Time       int                `json:"time"`
	StatusCode int                `json:"statusCode"`
	Data       interface{}        `json:"data"`    // TODO: 待补充
	OutData    BindWeChatOutData  `json:"outData"`
}

// =============================================================================
// 2. 根据Openid获取供应商信息
// =============================================================================

type GetByWeChatRequest struct {
	Openid string `json:"Openid" example:"oXxx123abc"`
	Type   string `json:"Type"   example:"1"` // 0:公众号 1:小程序
}

type GetByWeChatOutData struct {
	IntRetVal int `json:"intRetVal"`
}

type GetByWeChatResponse struct {
	IsSucceed  bool               `json:"isSucceed"`
	Message    *string            `json:"message"`
	Time       int                `json:"time"`
	StatusCode int                `json:"statusCode"`
	Data       interface{}        `json:"data"`    // TODO: 待补充（空数组=未绑定，有数据=已绑定）
	OutData    GetByWeChatOutData `json:"outData"`
}

// =============================================================================
// 3. 查询供应商可报价的招标信息
// =============================================================================

type InquiryQueryRequest struct {
	Supplier  string `json:"Supplier"  example:"C001"`
	PChnName  string `json:"P_chnName" example:""`        // 产品名称，模糊查询，可选
	BeginDate string `json:"BeginDate" example:""`        // 发布开始时间，可选
	EndDate   string `json:"EndDate"   example:""`        // 发布结束时间，可选
	IncludeEnd string `json:"IncludeEnd" example:"0"`     // 是否包含已截止 1:包含 0:不包含
}

type InquiryQueryResponse struct {
	IsSucceed  bool        `json:"isSucceed"`
	Message    *string     `json:"message"`
	Time       int         `json:"time"`
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`    // TODO: 待补充
	OutData    interface{} `json:"outData"`
}

// =============================================================================
// 4. 供应商查看采购需求详情
// =============================================================================

type InquiryDetailRequest struct {
	InquiryId string `json:"InquiryId" example:"INQ-2024-001"`
	Supplier  string `json:"Supplier"  example:"C001"`
}

type InquiryDetailOutData struct {
	StrMessage string `json:"strMessage"`
}

type InquiryDetailResponse struct {
	IsSucceed  bool                 `json:"isSucceed"`
	Message    *string              `json:"message"`
	Time       int                  `json:"time"`
	StatusCode int                  `json:"statusCode"`
	Data       interface{}          `json:"data"`    // TODO: 待补充
	OutData    InquiryDetailOutData `json:"outData"`
}

// =============================================================================
// 5. 撤回或删除供应商报价
// =============================================================================

type QuoteDeleteRequest struct {
	QuoteId     string `json:"QuoteId"     example:"QT-2024-001"`
	Supplier    string `json:"Supplier"    example:"C001"`
	OperateType string `json:"OperateType" example:"Cancel"` // Cancel:撤回 Delete:删除
}

type QuoteDeleteOutData struct {
	StrMessage string `json:"strMessage"`
}

type QuoteDeleteResponse struct {
	IsSucceed  bool               `json:"isSucceed"`
	Message    *string            `json:"message"`
	Time       int                `json:"time"`
	StatusCode int                `json:"statusCode"`
	Data       interface{}        `json:"data"`    // TODO: 待补充
	OutData    QuoteDeleteOutData `json:"outData"`
}

// =============================================================================
// 6. 供应商报价保存
// =============================================================================

type QuoteSaveRequest struct {
	Id                string `json:"Id"                example:""`              // 空=新增
	InquiryId         string `json:"InquiryId"         example:"INQ-2024-001"`
	Supplier          string `json:"Supplier"          example:"C001"`
	Price             string `json:"Price"             example:"13.00"`
	Unit              string `json:"Unit"              example:"个"`
	IsTax             string `json:"isTax"             example:"1"`
	TaxRate           string `json:"TaxRate"           example:"0.13"`
	Payment           string `json:"Payment"           example:"货到付款"`
	IsShipping        string `json:"isShipping"        example:"1"`
	LeadDays          int    `json:"LeadDays"          example:"7"`
	LeadTimeStartDate string `json:"LeadTimeStartDate" example:"1"`
	StockStatus       string `json:"StockStatus"       example:"1"`
	QuoteJson         string `json:"QuoteJson"         example:""`
	Remark            string `json:"Remark"            example:""`
	QStatus           string `json:"QStatus"           example:"0"` // 0:草稿 1:已撤回 2:已提交
}

type QuoteSaveOutData struct {
	ReturnId   string `json:"ReturnId"`
	StrMessage string `json:"strMessage"`
}

type QuoteSaveResponse struct {
	IsSucceed  bool             `json:"isSucceed"`
	Message    *string          `json:"message"`
	Time       int              `json:"time"`
	StatusCode int              `json:"statusCode"`
	Data       interface{}      `json:"data"`    // TODO: 待补充
	OutData    QuoteSaveOutData `json:"outData"`
}
