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

func NewAdminService(d ...dao.IAdminDao) *AdminService {
	var finalDao dao.IAdminDao

	if len(d) > 0 {
		finalDao = d[0]
	} else {
		finalDao = dao.NewAdminDao()
	}
	return &AdminService{
		adminDao: finalDao,
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
