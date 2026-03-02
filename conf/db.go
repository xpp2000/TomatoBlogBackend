package conf

import (
	"time"
	"tomatoBlogDB/global"
	"tomatoBlogDB/model"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB() (*gorm.DB, error) {
	logMode := logger.Info

	if !viper.GetBool("mode.develop") {
		logMode = logger.Error
	}

	db, err := gorm.Open(postgres.Open(viper.GetString("db.dsn")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "sys_",
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logMode),
	})

	if err != nil {
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(viper.GetInt("db.maxIdleConn"))
	sqlDB.SetMaxOpenConns(viper.GetInt("db.maxOpenConn"))
	sqlDB.SetConnMaxLifetime(time.Hour)

	// = migrate struct
	db.AutoMigrate(&model.Admin{})
	db.AutoMigrate(&model.Author{})
	db.AutoMigrate(&model.Category{})
	db.AutoMigrate(&model.Tag{})

	db.AutoMigrate(&model.Post{})
	// = seed data
	initAdmin(db)
	initPosts(db)
	return db, nil

}
func initAdmin(db *gorm.DB) {
	// ==========================================
	// 1. 初始化超级管理员 (Role = 999)
	// ==========================================
	var superAdmin model.Admin
	err := db.Where("name=?", viper.GetString("dataInit.InitAdminName")).First(&superAdmin).Error

	if err == gorm.ErrRecordNotFound {
		global.Logger.Info("Default super admin doesn't exist, creating......")

		newAdmin := model.Admin{
			Name:   viper.GetString("dataInit.InitAdminName"),
			Mobile: "99999999999",
			Email:  "admin@admin.com",
			Role:   999,
			// 超级管理员一般不需要暴露在前端展示，所以不需要关联 Author 档案
		}

		if err := newAdmin.SetPassword(viper.GetString("dataInit.InitAdminPsw")); err != nil {
			global.Logger.Panic("fail to encrypt password", err)
		}

		if err := db.Create(&newAdmin).Error; err != nil {
			global.Logger.Panic("fail to create default super admin", err)
		}
		global.Logger.Info("create super admin successfully")
	} else if err == nil {
		global.Logger.Info("Default super admin has existed")
	}

	// ==========================================
	// 2. 初始化默认作者 (Role = 1，并附带 Author 档案)
	// ==========================================
	var defaultAuthorAdmin model.Admin
	err = db.Where("name=?", viper.GetString("dataInit.InitAuthorName")).First(&defaultAuthorAdmin).Error

	if err == gorm.ErrRecordNotFound {
		global.Logger.Info("Default author doesn't exist, creating......")

		// 注意：如果你之前把 Author.Email 改成了 *string，这里需要先声明一个变量取地址
		email := "displaiedEmail@qq.com"

		newAuthorAdmin := model.Admin{
			Name:   viper.GetString("dataInit.InitAuthorName"),
			Mobile: "13135511982",
			Email:  "tomato@tomato.com",
			Role:   1, // 身份是普通作者

			// ⭐️ GORM 魔法：直接嵌套声明 Author 档案
			Author: &model.Author{
				PenName:  "Tomato Leo", // 之前提到的笔名/展示名
				Position: "Creator",    // 或者是 Bio
				Avatar:   "http://localhost/avatar1.jpg",
				PenEmail: &email, // 传入指针以防触发唯一索引空值报错
				// 绝对不需要手写 AdminID，GORM 会自动打通！
			},
		}

		if err := newAuthorAdmin.SetPassword(viper.GetString("dataInit.InitAuthorPsw")); err != nil {
			global.Logger.Panic("fail to encrypt author password", err)
		}

		// 一键落库，GORM 会自动开事务并向 admin 和 author 两张表插入数据
		if err := db.Create(&newAuthorAdmin).Error; err != nil {
			global.Logger.Panic("fail to create default author", err)
		}
		global.Logger.Info("create author successfully")
	} else if err == nil {
		global.Logger.Info("Default author has existed")
	}
}
func initPosts(db *gorm.DB) {
	// 1. 查找刚才创建的默认作者 (通过笔名查找)
	var author model.Author
	if err := db.Where("pen_name = ?", "Tomato Leo").First(&author).Error; err != nil {
		global.Logger.Info("Default author not found, skipping post seed.")
		return
	}

	// 2. 初始化一些默认分类 (Categories)
	cateTech := model.Category{Name: "技术探索", Description: "代码与架构的沉淀"}
	cateStudy := model.Category{Name: "学习笔记", Description: "考研与学习日常"}

	// FirstOrCreate: 如果 Name 不存在则创建，存在则直接将查到的数据查出并赋给变量 (带上了 ID)
	db.FirstOrCreate(&cateTech, model.Category{Name: "技术探索"})
	db.FirstOrCreate(&cateStudy, model.Category{Name: "学习笔记"})

	// 3. 初始化一些默认标签 (Tags)
	tagGo := model.Tag{Name: "Golang", Description: "Go 语言后端开发"}
	tagReact := model.Tag{Name: "React", Description: "前端工程化"}
	tagMath := model.Tag{Name: "数学", Description: "考研数学"}
	tagLatex := model.Tag{Name: "LaTeX", Description: "排版与公式渲染"}
	tagEnglish := model.Tag{Name: "英语", Description: "词汇与长难句"}

	db.FirstOrCreate(&tagGo, model.Tag{Name: "Golang"})
	db.FirstOrCreate(&tagReact, model.Tag{Name: "React"})
	db.FirstOrCreate(&tagMath, model.Tag{Name: "数学"})
	db.FirstOrCreate(&tagLatex, model.Tag{Name: "LaTeX"})
	db.FirstOrCreate(&tagEnglish, model.Tag{Name: "英语"})

	// 4. 准备 5 条测试文章数据
	now := time.Now()
	posts := []model.Post{
		{
			Title:       "TomatoBlog 架构设计指南",
			Subtitle:    "基于 Go 与 GORM 的全栈 CMS",
			Slug:        "tomatoblog-architecture-guide",
			Summary:     "探讨如何利用 GORM 处理复杂的表关联，以及如何优雅地实现依赖注入。",
			Keywords:    "Go, GORM, CMS, 后端架构",
			Content:     "## 架构概览\n本项目采用了严格的三层架构，并结合了极其灵活的 DI（依赖注入）机制...\n\n### 数据库设计\n核心围绕 Admin 与 Author 的 1:1 分离设计展开...",
			Cover:       "http://localhost/cover_blog.jpg",
			AuthorID:    author.ID,                       // 挂载作者
			CategoryID:  cateTech.ID,                     // 挂载分类
			Tags:        []*model.Tag{&tagGo, &tagReact}, // 魔法：直接传入标签的指针地址，GORM 会自动写入 post_tags 中间表！
			Status:      1,                               // 1-published
			PublishedAt: now,
		},
		{
			Title:       "TanStack Table 结合 Tailwind 打造高性能表格",
			Subtitle:    "前端 Headless UI 组件实践",
			Slug:        "tanstack-table-tailwind-practice",
			Summary:     "Headless 组件库只提供逻辑不提供 UI，这篇文章教你如何用 Tailwind 为其赋予灵魂。",
			Keywords:    "React, TanStack, Tailwind, 前端",
			Content:     "## 为什么选择 Headless UI?\n传统的表格组件往往过于臃肿，而 TanStack Table 给了我们最大的自定义自由度...",
			Cover:       "http://localhost/cover_react.jpg",
			AuthorID:    author.ID,
			CategoryID:  cateTech.ID,
			Tags:        []*model.Tag{&tagReact},
			Status:      1,
			PublishedAt: now.Add(-24 * time.Hour), // 昨天发布的
		},
		{
			Title:       "考研数学：定积分的核心题型解析",
			Subtitle:    "掌握计算曲线下方图形面积的基石",
			Slug:        "definite-integral-notes",
			Summary:     "梳理定积分的基础概念与常见考研题型的解题套路。",
			Keywords:    "数学, 考研, 定积分",
			Content:     "## 定积分的几何意义\n在 LaTeX 中，我们通常使用 $\\int_{a}^{b} f(x) dx$ 来表示定积分...\n\n### 常见替换技巧\n三角代换是解决复杂根号积分的利器...",
			Cover:       "http://localhost/cover_math.jpg",
			AuthorID:    author.ID,
			CategoryID:  cateStudy.ID,
			Tags:        []*model.Tag{&tagMath, &tagLatex},
			Status:      1,
			PublishedAt: now.Add(-48 * time.Hour),
		},
		{
			Title:       "LaTeX 笔记排版入门到进阶",
			Subtitle:    "优雅输出数学公式与学术文档",
			Slug:        "latex-typesetting-guide",
			Summary:     "告别 Word 的排版烦恼，用代码的方式书写结构化的高质量笔记。",
			Keywords:    "LaTeX, 笔记, 效率工具",
			Content:     "## 环境搭建\n推荐使用 TeX Live 配合 VS Code...\n\n### 数学环境\n使用 `equation` 环境可以自动为公式编号...",
			Cover:       "http://localhost/cover_latex.jpg",
			AuthorID:    author.ID,
			CategoryID:  cateStudy.ID,
			Tags:        []*model.Tag{&tagLatex},
			Status:      1,
			PublishedAt: now.Add(-72 * time.Hour),
		},
		{
			Title:       "考研英语核心词汇记忆指南",
			Subtitle:    "如何科学背诵长难句词汇",
			Slug:        "english-vocabulary-study-guide",
			Summary:     "探讨词根词缀记忆法与艾宾浩斯遗忘曲线在英语单词背诵中的应用。",
			Keywords:    "英语, 考研, 单词",
			Content:     "## 为什么要背单词？\n词汇量是阅读理解的基础...\n\n### 记忆策略\n通过多次重复而非单次长时间注视来加深印象...",
			Cover:       "http://localhost/cover_english.jpg",
			AuthorID:    author.ID,
			CategoryID:  cateStudy.ID,
			Tags:        []*model.Tag{&tagEnglish},
			Status:      0, // 0-draft 设为草稿状态，用于测试列表过滤
			PublishedAt: now.Add(-96 * time.Hour),
		},
	}

	// 5. 循环插入文章数据 (如果 Slug 已存在则跳过，防止重复执行报错)
	var insertCount int
	for _, post := range posts {
		var count int64
		db.Model(&model.Post{}).Where("slug = ?", post.Slug).Count(&count)

		if count == 0 {
			if err := db.Create(&post).Error; err != nil {
				global.Logger.Error("Fail to insert seed post: "+post.Slug, err)
			} else {
				insertCount++
			}
		}
	}

	if insertCount > 0 {
		global.Logger.Info("Create post seeds successfully")
	} else {
		global.Logger.Info("Post seeds already exist")
	}
}
