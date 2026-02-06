package user

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
		PhoneHash:      crypto.Hash(req.Phone),
		PhoneEncrypted: phoneEncrypted,
		Status:         "active",
	}

	// 尝试创建，用户名冲突时重试（最多10次）
	var createErr error
	for i := 0; i < 10; i++ {
		createErr = s.repo.Create(ctx, user)
		if createErr == nil {
			break
		}
		user.Username = generateUsername()
	}
	if createErr != nil {
		return nil, errors.New("创建用户失败，用户名生成冲突")
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
	phoneHash := crypto.Hash(phone)
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
