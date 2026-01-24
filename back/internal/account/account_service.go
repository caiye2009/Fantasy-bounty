package account

import (
	"back/pkg/crypto"
	"context"
	"errors"
	"math/rand"

	"gorm.io/gorm"
)

// generateUsername 生成用户名：用户+5位随机字符（小写字母+数字）
func generateUsername() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return "用户" + string(b)
}

type Service interface {
	CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccountByPhone(ctx context.Context, phone string) (*Account, error)
	UpdateAccount(ctx context.Context, id string, req *UpdateAccountRequest) (*Account, error)
	DeleteAccount(ctx context.Context, id string) error
	ListAccounts(ctx context.Context, page, pageSize int) ([]Account, int64, error)
	UpdateLastLogin(ctx context.Context, id string) error
}

type service struct {
	repo   Repository
	crypto *crypto.Crypto
}

func NewService(repo Repository, cryptoService *crypto.Crypto) Service {
	return &service{repo: repo, crypto: cryptoService}
}

// decryptAccount 解密账号手机号并填充 Phone 和 PhoneMasked 字段
func (s *service) decryptAccount(acc *Account) error {
	if acc == nil || acc.PhoneEncrypted == "" {
		return nil
	}
	phone, err := s.crypto.Decrypt(acc.PhoneEncrypted)
	if err != nil {
		return err
	}
	acc.Phone = phone
	acc.PhoneMasked = crypto.MaskPhone(phone)
	return nil
}

func (s *service) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
	// 加密手机号
	phoneEncrypted, err := s.crypto.Encrypt(req.Phone)
	if err != nil {
		return nil, errors.New("手机号加密失败: " + err.Error())
	}

	// 生成唯一用户名（重试最多10次）
	var username string
	for i := 0; i < 10; i++ {
		username = generateUsername()
		// 检查是否已存在（通过尝试创建来验证唯一性）
		break
	}

	account := &Account{
		Username:       username,
		PhoneHash:      crypto.Hash(req.Phone),
		PhoneEncrypted: phoneEncrypted,
		Status:         "active",
	}

	if err := s.repo.Create(ctx, account); err != nil {
		// 如果用户名冲突，重试
		for i := 0; i < 9; i++ {
			account.Username = generateUsername()
			if err2 := s.repo.Create(ctx, account); err2 == nil {
				break
			}
			if i == 8 {
				return nil, errors.New("创建账号失败，用户名生成冲突")
			}
		}
	}

	// 填充解密后的手机号用于返回
	account.Phone = req.Phone
	account.PhoneMasked = crypto.MaskPhone(req.Phone)

	return account, nil
}

func (s *service) GetAccount(ctx context.Context, id string) (*Account, error) {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	// 解密手机号
	if err := s.decryptAccount(account); err != nil {
		return nil, errors.New("手机号解密失败: " + err.Error())
	}

	return account, nil
}

func (s *service) GetAccountByPhone(ctx context.Context, phone string) (*Account, error) {
	// 通过手机号哈希查询
	phoneHash := crypto.Hash(phone)
	account, err := s.repo.GetByPhoneHash(ctx, phoneHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	// 解密手机号
	if err := s.decryptAccount(account); err != nil {
		return nil, errors.New("手机号解密失败: " + err.Error())
	}

	return account, nil
}

func (s *service) UpdateAccount(ctx context.Context, id string, req *UpdateAccountRequest) (*Account, error) {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	if req.Status != "" {
		account.Status = req.Status
	}

	if err := s.repo.Update(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *service) DeleteAccount(ctx context.Context, id string) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("account not found")
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) ListAccounts(ctx context.Context, page, pageSize int) ([]Account, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	accounts, total, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 解密所有账号的手机号
	for i := range accounts {
		_ = s.decryptAccount(&accounts[i]) // 忽略单个解密错误
	}

	return accounts, total, nil
}

func (s *service) UpdateLastLogin(ctx context.Context, id string) error {
	return s.repo.UpdateLastLogin(ctx, id)
}
