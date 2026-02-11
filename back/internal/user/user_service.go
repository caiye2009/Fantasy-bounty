package user

import (
	"back/pkg/crypto"
	"context"
	"errors"
	"math/rand"
	"strings"

	"gorm.io/gorm"
)

// generateUsername 生成用户名：8位base62随机字符（0-9a-zA-Z）
func generateUsername() string {
	const base62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)
	for i := range b {
		b[i] = base62[rand.Intn(len(base62))]
	}
	return string(b)
}

// isUniqueConstraintError 检查是否是唯一约束冲突错误
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL 唯一约束错误包含 "duplicate key" 或 "unique constraint"
	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "UNIQUE constraint")
}

type Service interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, page, pageSize int) ([]User, int64, error)
	UpdateLastLogin(ctx context.Context, id string) error
}

type service struct {
	repo   Repository
	crypto *crypto.Crypto
}

func NewService(repo Repository, cryptoService *crypto.Crypto) Service {
	return &service{repo: repo, crypto: cryptoService}
}

// decryptUser 解密用户手机号并填充 Phone 和 PhoneMasked 字段
func (s *service) decryptUser(u *User) error {
	if u == nil || u.PhoneEncrypted == "" {
		return nil
	}
	phone, err := s.crypto.Decrypt(u.PhoneEncrypted)
	if err != nil {
		return err
	}
	u.Phone = phone
	u.PhoneMasked = crypto.MaskPhone(phone)
	return nil
}

func (s *service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// 加密手机号
	phoneEncrypted, err := s.crypto.Encrypt(req.Phone)
	if err != nil {
		return nil, errors.New("手机号加密失败: " + err.Error())
	}

	user := &User{
		Username:       generateUsername(),
		PhoneHash:      s.crypto.Hash(req.Phone),
		PhoneEncrypted: phoneEncrypted,
		Status:         "active",
	}

	// 尝试创建，仅在用户名唯一冲突时重试1次
	// 8位base62有62^8≈218万亿种组合，冲突率极低（百万用户下<0.046%）
	createErr := s.repo.Create(ctx, user)
	if createErr != nil {
		// 只有唯一约束冲突才重试，其他错误直接返回
		if isUniqueConstraintError(createErr) {
			// 用户名冲突，重新生成再试一次
			user.Username = generateUsername()
			createErr = s.repo.Create(ctx, user)
		}

		if createErr != nil {
			return nil, createErr
		}
	}

	// 填充解密后的手机号用于返回
	user.Phone = req.Phone
	user.PhoneMasked = crypto.MaskPhone(req.Phone)

	return user, nil
}

func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 解密手机号
	if err := s.decryptUser(user); err != nil {
		return nil, errors.New("手机号解密失败: " + err.Error())
	}

	return user, nil
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 解密手机号
	if err := s.decryptUser(user); err != nil {
		return nil, errors.New("手机号解密失败: " + err.Error())
	}

	return user, nil
}

func (s *service) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	// 通过手机号哈希查询
	phoneHash := s.crypto.Hash(phone)
	user, err := s.repo.GetByPhoneHash(ctx, phoneHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 解密手机号
	if err := s.decryptUser(user); err != nil {
		return nil, errors.New("手机号解密失败: " + err.Error())
	}

	return user, nil
}

func (s *service) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if req.Status != "" {
		user.Status = req.Status
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) DeleteUser(ctx context.Context, id string) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) ListUsers(ctx context.Context, page, pageSize int) ([]User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, total, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 解密所有用户的手机号
	for i := range users {
		_ = s.decryptUser(&users[i]) // 忽略单个解密错误
	}

	return users, total, nil
}

func (s *service) UpdateLastLogin(ctx context.Context, id string) error {
	return s.repo.UpdateLastLogin(ctx, id)
}
