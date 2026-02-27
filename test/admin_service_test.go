package test

import (
	"errors"
	"testing"
	"tomatoBlogDB/dto"
	"tomatoBlogDB/model"
	"tomatoBlogDB/service"
	mock_dao "tomatoBlogDB/test/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAdminService_Login_Success(t *testing.T) {
	// 1. 初始化 Mock 控制器
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // 测试结束时清理

	// 2. 创建一个假的 DAO (Mock 对象)
	mockUserDao := mock_dao.NewMockIAdminDao(ctrl)

	// 3. 准备测试数据
	passwordRaw := "123456"
	// 因为 Service 里会校验 hash，所以我们在测试数据里要先把密码加密好
	hash, _ := bcrypt.GenerateFromPassword([]byte(passwordRaw), bcrypt.DefaultCost)

	mockUser := &model.Admin{
		BaseModel: model.BaseModel{ID: 1},
		Name:      "admin",
		Password:  string(hash), // 数据库里存的是加密后的
	}

	// 4. 【核心】设定 Mock 的行为 (打桩 / Stubbing)
	// 意思是：当 Service 调用 GetUserByName("admin") 时，Mock 必须返回 mockUser 和 nil (无错误)
	mockUserDao.EXPECT().
		GetAdminByName("admin"). // 期望参数是 "admin"
		Return(mockUser, nil).   // 模拟返回数据库查到的 user
		Times(1)                 // 期望被调用 1 次

	// 5. 初始化 Service (注入假的 DAO)
	adminService := service.NewAdminService(mockUserDao)

	// 6. 调用被测方法
	req := dto.AdminLoginReq{
		Name:     "admin",
		Password: passwordRaw, // 传入明文
	}
	token, user, err := adminService.Login(req)

	// 7. 断言结果 (验证结果是否符合预期)
	assert.NoError(t, err)              // 应该没有错误
	assert.NotNil(t, user)              // 用户对象不为空
	assert.Equal(t, uint64(1), user.ID) // ID 应该是 1
	assert.NotEmpty(t, token)           // Token 不应该为空
}

func TestAdminService_Login_Fail_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDao := mock_dao.NewMockIAdminDao(ctrl)

	// 场景：模拟数据库找不到用户
	mockUserDao.EXPECT().
		GetAdminByName("unknown").
		Return(nil, errors.New("record not found")). // 模拟数据库报错
		Times(1)

	adminService := service.NewAdminService(mockUserDao)

	req := dto.AdminLoginReq{
		Name:     "unknown",
		Password: "123",
	}
	_, _, err := adminService.Login(req)

	// 断言：应该报错
	assert.Error(t, err)
	assert.Equal(t, "record not found", err.Error()) // 假设你的 Service 处理了错误转换
}
