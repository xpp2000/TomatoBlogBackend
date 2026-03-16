package service

import (
	"errors"
	"tomatoBlogDB/dao"
	"tomatoBlogDB/dto"
	"tomatoBlogDB/errcode"
	"tomatoBlogDB/model"
	"tomatoBlogDB/utils"

	"gorm.io/gorm"
)

type AdminService struct {
	adminDao dao.IAdminDao
	postDao  dao.IPostDao
}

func NewAdminService(adminDao dao.IAdminDao, postDao dao.IPostDao) *AdminService {
	return &AdminService{
		adminDao: adminDao,
		postDao:  postDao,
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
			return "", nil, errcode.ErrAdminNotFound
		}
		return "", nil, errcode.NewSysErr(err)
	}

	// 2. verify password
	if !admin.CheckPassword(req.Password) {
		return "", nil, errcode.ErrPassWordIncorrect
	}

	// 3. generate token
	token, err := utils.GenerateToken(admin.ID, admin.Name, admin.Role)
	if err != nil {
		return "", nil, errcode.NewSysErr(err)
	}

	return token, admin, nil
}

// ==== Create admin
func (s *AdminService) CreateAdmin(req dto.AdminAddReq) error {
	// 1. permission check
	// role has been checked on api
	adminRole := 999

	// 2. assembly admin
	finalRealName := req.RealName
	if finalRealName == "" {
		finalRealName = "anonymous"
	}
	admin := &model.Admin{
		Name:     req.Name,
		RealName: req.RealName,
		Mobile:   req.Mobile,
		Email:    req.Email,
		Role:     adminRole,
	}
	if err := admin.SetPassword(req.Password); err != nil {
		return errcode.NewSysErr(err)
	}
	// 3. check duplicated
	if isExist, _ := s.adminDao.CheckUnique("email", req.Email); isExist {
		return errcode.ErrEmailExist
	}
	if isExist, _ := s.adminDao.CheckUnique("name", req.Name); isExist {
		return errcode.ErrNameExist
	}
	if isExist, _ := s.adminDao.CheckUnique("mobil", req.Mobile); isExist {
		return errcode.ErrMobileExist
	}

	// 5.
	err := s.adminDao.CreateAdmin(admin)
	if err != nil {
		return errcode.NewSysErr(err)
	}
	return nil
}

// ==== Create author
func (s *AdminService) CreateAuthor(req dto.AuthorAddReq) error {
	// 1. permission check
	// role has been checked on api

	authorRole := 1

	// 2. assembly admin
	finalRealName := req.RealName
	if finalRealName == "" {
		finalRealName = "anonymous"
	}
	author := &model.Admin{
		Name:     req.Name,
		RealName: req.RealName,
		Mobile:   req.Mobile,
		Email:    req.Email,
		Role:     authorRole,
	}
	if err := author.SetPassword(req.Password); err != nil {
		return errcode.NewSysErr(err)
	}
	// 3. check duplicated
	if isExist, _ := s.adminDao.CheckUnique("email", req.Email); isExist {
		return errcode.ErrEmailExist
	}
	if isExist, _ := s.adminDao.CheckUnique("name", req.Name); isExist {
		return errcode.ErrNameExist
	}
	if isExist, _ := s.adminDao.CheckUnique("mobil", req.Mobile); isExist {
		return errcode.ErrMobileExist
	}

	// 4. create Author item concurrently
	author.Author = &model.Author{
		PenName: req.PenName,
		// 此时不需要填 AdminID，GORM 插入 admin 后会自动把生成的 ID 填到这里！
	}

	// 5.
	err := s.adminDao.CreateAdmin(author)
	if err != nil {
		return errcode.NewSysErr(err)
	}
	return nil
}

// ==== UpdateAdminStatus
func (s *AdminService) UpdateAdminStatus(targetID uint64, status int, currentAdminID uint64) error {
	// 1. prevent banning self
	if targetID == 1 {
		return errcode.ErrMaliciousTarget
	}
	if targetID == currentAdminID {
		return errcode.ErrMaliciousOperation
	}
	// 2. status field check
	if status != 1 && status != 2 {
		return errcode.ErrAdminStatus
	}

	// 3.
	if err := s.adminDao.UpdateAdminStatus(targetID, status); err != nil {
		return errcode.NewSysErr(err)
	}
	return nil

}

// ==== DeleteAdmin
func (s *AdminService) DeleteAdmin(targetID uint64, currentAdminID uint64) error {
	if targetID == 1 {
		return errcode.ErrMaliciousTarget
	}
	if targetID == currentAdminID {
		return errcode.ErrMaliciousTarget
	}

	// 业务拓展：你甚至可以在这里调用 postDao 检查该作者名下有没有文章，有的话不让删
	// TODO
	// 1. 先查出这个即将被删的倒霉蛋
	admin, err := s.adminDao.GetAdminByID(targetID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrAdminNotFound
		}
		return errcode.NewSysErr(err) // 抛出系统错误给 HandleError 记日志
	}
	// 2. 核心约束校验：检查文章依赖
	// 只有当存在 Author 档案时，才需要去查文章表
	if admin.Author != nil {
		postCount, err := s.postDao.CountByAuthorID(admin.Author.ID)
		if err != nil {
			return errcode.NewSysErr(err) // 数据库查挂了
		}

		// 铁律：名下有资产，绝对不让销户
		if postCount > 0 {
			// 409 是标准的 HTTP 冲突状态码 (Conflict)
			return errcode.ErrDeletePostFirst
		}
	}
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
