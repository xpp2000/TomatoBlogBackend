package service

import (
	"errors"
	"tomatoBlogDB/dao"
	"tomatoBlogDB/dto"
	"tomatoBlogDB/model"
	"tomatoBlogDB/utils"

	"gorm.io/gorm"
)

type AdminService struct {
	adminDao dao.IAdminDao
}

func NewAdminService(adminDao dao.IAdminDao) *AdminService {
	return &AdminService{
		adminDao: adminDao,
	}
}

// ==== Login
// -input: DTO
// -output: Token string, User, err
func (s *AdminService) Login(req dto.AdminLoginReq) (string, *model.Admin, error) {

	// 1. find user via DAO
	admin, err := s.adminDao.GetAdminByName(req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("administrator doesn't exist ")
		}
		return "", nil, err
	}

	// 2. verify password
	if !admin.CheckPassword(req.Password) {
		return "", nil, errors.New("incorrect password")
	}

	// 3. generate token
	token, err := utils.GenerateToken(admin.ID, admin.Name, admin.Role)
	if err != nil {
		return "", nil, errors.New("fail to generate token: " + err.Error())
	}

	return token, admin, nil
}

// ==== Create
func (s *AdminService) CreateAdmin(req dto.AdminAddReq, operatorRole int) error {
	// 1. permission check
	if operatorRole != ROLE_ADMIN {
		return errors.New("permission denied")
	}
	if req.Role == ROLE_ADMIN {
		return errors.New("permission denied")
	}

	// 2. assembly admin
	admin := &model.Admin{
		Name:     req.Name,
		RealName: req.RealName,
		Mobile:   req.Mobile,
		Email:    req.Email,
		Role:     req.Role,
	}
	if err := admin.SetPassword(req.Password); err != nil {
		return err
	}

	// 3.
	// 只有当 Role 为普通作者(1) 时，才为其初始化作者档案
	if admin.Role == 1 {
		admin.Author = &model.Author{
			PenName: req.PenName,
			// 此时不需要填 AdminID，GORM 插入 admin 后会自动把生成的 ID 填到这里！
		}
	}
	// 4.
	return s.adminDao.CreateAdmin(admin)
}

// ==== UpdateAdminStatus
func (s *AdminService) UpdateAdminStatus(targetID uint64, status int, currentAdminID uint64) error {
	// 1. prevent banning self
	if targetID == 1 {
		return errors.New("cannot disable the supreme administrator")
	}
	if targetID == currentAdminID {
		return errors.New("can't execute this operation")
	}
	// 2. status field check
	if status != 1 && status != 2 {
		return errors.New("invalid status")
	}

	return s.adminDao.UpdateAdminStatus(targetID, status)

}

// ==== DeleteAdmin
func (s *AdminService) DeleteAdmin(targetID uint64, currentAdminID uint64) error {
	if targetID == 1 {
		return errors.New("cannot delete the supreme administrator")
	}
	if targetID == currentAdminID {
		return errors.New("you cannot delete your own account")
	}

	// 业务拓展：你甚至可以在这里调用 postDao 检查该作者名下有没有文章，有的话不让删
	// TODO
	return s.adminDao.DeleteAdmin(targetID)
}

// ====
func (s *AdminService) GetAdminList(req dto.AdminListReq) ([]*dto.AdminListResp, int64, error) {
	// 1. default parameters check
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大页大小，防止恶意攻击
	}
	// 2. DAO
	admins, total, err := s.adminDao.GetAdminList(req)
	if err != nil {
		return nil, 0, err
	}

	// 3. data cleaning
	respList := make([]*dto.AdminListResp, 0, len(admins))
	for _, admin := range admins {
		resp := &dto.AdminListResp{
			ID:        admin.ID,
			Name:      admin.Name,
			Mobile:    admin.Mobile,
			Role:      admin.Role,
			Status:    admin.Status,
			CreatedAt: admin.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// 安全地提取 Author 信息（如果存在的话）
		if admin.Author != nil {
			resp.PenName = admin.Author.PenName
			resp.Avatar = admin.Author.Avatar
			if admin.Author.PenEmail != nil {
				resp.Email = *admin.Author.PenEmail
			}
		}

		respList = append(respList, resp)
	}

	return respList, total, nil
}
