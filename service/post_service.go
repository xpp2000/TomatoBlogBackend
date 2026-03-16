package service

import (
	"errors"
	"strconv"
	"time"
	"tomatoBlogDB/dao"
	"tomatoBlogDB/dto"
	"tomatoBlogDB/errcode"
	"tomatoBlogDB/model"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

const ROLE_ADMIN = 999

// TODO: DAO存在复用怎么办？
type PostService struct {
	postDao dao.IPostDao
	tagDao  dao.ITagDao
}

// Dependency Injection design
func NewPostService(postDao dao.IPostDao, tagDao dao.ITagDao) *PostService {
	return &PostService{
		postDao: postDao,
		tagDao:  tagDao,
	}
}

type CategoryService struct {
	categoryDao dao.ICategoryDao
	postDao     dao.IPostDao
}

func NewCategoryService(categoryDao dao.ICategoryDao, postDao dao.IPostDao) *CategoryService {
	return &CategoryService{
		categoryDao: categoryDao,
		postDao:     postDao,
	}
}

type AuthorService struct {
	authorDao dao.IAuthorDao
	postDao   dao.IPostDao
}

func NewAuthorService(authorDao dao.IAuthorDao, postDao dao.IPostDao) *AuthorService {
	return &AuthorService{
		authorDao: authorDao,
		postDao:   postDao,
	}
}

// ==== GetPostDetail
// - slug or id
func (s *PostService) GetPostDetail(param string) (*dto.PostSimple, error) {
	id, err := strconv.ParseUint(param, 10, 64)
	var post *model.Post
	var dbErr error
	// 1. parameter check, and get post
	if err == nil && id > 0 {
		// A. ID search
		post, dbErr = s.postDao.GetPostDetailByID(id)
	} else {
		// B. Slug search
		post, dbErr = s.postDao.GetPostDetailBySlug(param)
	}

	if dbErr != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrPostNotFound
		}
		return nil, errcode.NewSysErr(dbErr)
	}

	// 2. assembly respPost
	var tagNames []string
	for _, tag := range post.Tags {
		tagNames = append(tagNames, tag.Name)
	}
	var author = dto.AuthorSimple{
		ID:       post.Author.ID,
		PenName:  post.Author.PenName,
		Position: post.Author.Position,
		Avatar:   post.Author.Avatar,
	}
	var postSimple = &dto.PostSimple{
		Content:      post.Content,
		ID:           post.ID, // ID 必须给前端
		Title:        post.Title,
		Slug:         post.Slug,
		Subtitle:     post.Subtitle,
		Summary:      post.Summary,
		Keywords:     post.Keywords,
		Cover:        post.Cover,
		ReadCount:    post.ReadCount,
		AuthorID:     post.AuthorID,
		Author:       author, // 预加载的作者名
		CategoryID:   post.CategoryID,
		CategoryName: post.Category.Name,                    // 预加载的分类名
		Tags:         tagNames,                              // 提取好的标签数组
		PublishedAt:  post.PublishedAt.Format(time.RFC3339), // 格式化时间
		Status:       post.Status,
	}

	// 2. post operations
	// -2.1 asynchronously increasing
	go s.postDao.IncrReadCount(post.ID)

	return postSimple, nil
}

// ==== CreatePost
// authorID: fetch from Controller layer via Token
// NOTE: For writer
func (s *PostService) CreatePost(req dto.PostAddReq, authorID uint64, operatorRole int) error {
	// 1. slug
	finalSlug := req.Slug
	if finalSlug == "" {
		finalSlug = slug.Make(req.Title)

	}

	// 2. 处理 Tags（去数据库查找或创建标签）
	var tags []*model.Tag // 提前声明，如果没有传入标签，它就是 nil
	if len(req.Tags) > 0 {
		var err error
		tags, err = s.tagDao.GetOrCreateByNames(req.Tags)
		if err != nil {
			return errcode.NewSysErr(err) // 如果处理标签时数据库报错，直接抛出
		}
	}

	// 3. published_at
	// 如果前端传了时间字符串
	now := time.Now()
	FinalPublishedAt := now
	if req.PublishedAt != "" {
		// 尝试按 RFC3339 解析
		if t, err := time.Parse(time.RFC3339, req.PublishedAt); err == nil {
			FinalPublishedAt = t
		} else if t2, err2 := time.Parse(time.DateTime, req.PublishedAt); err2 == nil {
			// 2. 尝试按 "YYYY-MM-DD HH:MM:SS" 解析
			FinalPublishedAt = t2
		} else if t3, err3 := time.Parse(time.DateOnly, req.PublishedAt); err3 == nil {
			// 3. 尝试按 "YYYY-MM-DD" 解析 (Go 1.20+)
			// 解析结果的默认时分秒会是 00:00:00
			FinalPublishedAt = t3
		} else {
			// [强烈建议] 这里最好抛出错误或者打日志，防止前端传了乱码导致被默默吞掉并赋值为 now
			// m.HandleError(ctx, fmt.Errorf("invalid time format: %s", req.PublishedAt))
			// return
		}
	}

	// 处理“早于当前时间”的逻辑
	if FinalPublishedAt.Before(now) {
		FinalPublishedAt = now
	}
	// 4. assemble model
	//  4.1 back-door for admin
	var finalAuthorID uint64 = authorID
	if operatorRole == 999 {
		finalAuthorID = req.TargetAuthorID
	}
	// 4.2
	post := model.Post{
		Title:       req.Title,
		Slug:        finalSlug,
		Subtitle:    req.Subtitle,
		Summary:     req.Summary,
		Keywords:    req.Keywords,
		Content:     req.Content,
		Cover:       req.Cover,
		AuthorID:    finalAuthorID,
		CategoryID:  req.CategoryID,
		Tags:        tags,
		Status:      1,
		PublishedAt: FinalPublishedAt,
	}

	// 3.
	err := s.postDao.CreatePost(&post)
	if err != nil {
		return errcode.NewSysErr(err)
	}
	return nil
}

// ==== UpdatePost
// Description ---- for writer
// operatorID:
// operatorRole:
func (s *PostService) UpdatePost(req dto.PostUpdateReq, operatorID uint64, operatorRole int) error {
	// 1. check post
	post, err := s.postDao.GetPostByID(req.ID)
	if err != nil {
		return errcode.ErrPostNotFound
	}

	// 2. permission check
	if post.AuthorID != operatorID {
		if operatorRole != ROLE_ADMIN {
			return errcode.ErrAuthorNotMatch
		}
	}

	// 3. prepare data to be updated
	// 注意：这里只赋值需要更新的字段，AuthorID 不需要赋值（禁止转让作者）

	// 3.1
	updateData := make(map[string]interface{})
	{
		if req.Title != nil {
			updateData["title"] = *req.Title
		}
		if req.Subtitle != nil {
			updateData["subtitle"] = *req.Subtitle
		}
		if req.Summary != nil {
			updateData["summary"] = *req.Summary
		}
		if req.Keywords != nil {
			updateData["keywords"] = *req.Keywords
		}
		if req.Content != nil {
			updateData["content"] = *req.Content
		}
		if req.Cover != nil {
			updateData["cover"] = *req.Cover
		}
		if req.CategoryID != nil {
			updateData["category_id"] = *req.CategoryID
		}
		if req.Status != nil {
			updateData["status"] = *req.Status // 即使前端传了 0，也会被正确存入 map
		}
		finalSlug := req.Slug
		if finalSlug != nil {
			updateData["slug"] = *req.Slug
		} else if req.Title != nil && post.Slug == "" {
			// 只有当标题更新了，且原来没有 slug 时，才自动生成
			updateData["slug"] = slug.Make(*req.Title)
		}
	}
	// 3.2 Tags
	var tagsID []uint64
	if req.Tags != nil {
		tags, err := s.tagDao.GetOrCreateByNames(req.Tags)
		if err != nil {
			return errcode.NewSysErr(err)
		}
		for _, tag := range tags {
			tagsID = append(tagsID, tag.ID)
		}
	}

	// 4.
	if len(updateData) == 0 && req.Tags == nil {
		return nil
	}

	if err := s.postDao.UpdatePost(req.ID, updateData, tagsID, req.Tags != nil); err != nil {
		return errcode.NewSysErr(err)
	}
	return nil
}

// ==== UpdateStatus
// Description ---- for writer and super admin

func (s *PostService) UpdateStatus(req dto.PostStatusReq, operatorID uint64, operatorRole int) error {
	// 1. search the post
	post, err := s.postDao.GetPostByID(req.ID)
	if err != nil {
		return errcode.ErrPostNotFound
	}

	// 2. permission check
	if operatorRole == ROLE_ADMIN {
		// = admin
	} else {
		// = user
		// B1.
		if post.AuthorID != operatorID {
			return errcode.ErrAuthorNotMatch
		}
		// B2.
		if post.Status == 2 {
			return errcode.ErrPostLocked
		}
	}

	// 3.  execute
	err = s.postDao.UpdatePostStatus(req.ID, req.Status)
	if err != nil {
		return errcode.NewSysErr(err)
	}

	return nil
}

// ==== ChangePostAuthor
// TODO!

// ==== DeletePost
func (s *PostService) DeletePost(id uint64, operatorID uint64, operatorRole int) error {
	// 1. check post
	post, err := s.postDao.GetPostByID(id)
	if err != nil {
		return errcode.ErrPostNotFound
	}

	// 2. permission check
	if operatorID != ROLE_ADMIN && post.AuthorID != operatorID {
		return errcode.ErrPermissionDenied
	}
	// 3. delete
	return s.postDao.DeletePost(id)
}

// ==== GetPostList
func (s *PostService) GetPostList(req dto.PostListReq) ([]*dto.PostSimple, int64, error) {
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
	posts, total, err := s.postDao.GetPostList(req)
	if err != nil {
		return nil, 0, err
	}

	// 3. postSimple Slice
	respList := make([]*dto.PostSimple, 0, len(posts))
	for _, p := range posts {
		// 提取 Tag 名称组成简单的字符串数组
		var tagNames []string
		for _, tag := range p.Tags {
			tagNames = append(tagNames, tag.Name)
		}

		var author = dto.AuthorSimple{
			ID:       p.Author.ID,
			PenName:  p.Author.PenName,
			Position: p.Author.Position,
			Avatar:   p.Author.Avatar,
		}

		// 组装 DTO (不含content)
		resp := &dto.PostSimple{
			Content:      "",
			ID:           p.ID, // ID 必须给前端
			Title:        p.Title,
			Slug:         p.Slug,
			Subtitle:     p.Subtitle,
			Summary:      p.Summary,
			Keywords:     p.Keywords,
			Cover:        p.Cover,
			ReadCount:    p.ReadCount,
			AuthorID:     p.AuthorID,
			Author:       author, // 预加载的作者名
			CategoryID:   p.CategoryID,
			CategoryName: p.Category.Name,                    // 预加载的分类名
			Tags:         tagNames,                           // 提取好的标签数组
			PublishedAt:  p.PublishedAt.Format(time.RFC3339), // 格式化时间
			Status:       p.Status,
		}

		respList = append(respList, resp)
	}

	// 4. return
	return respList, total, nil
}

// ==== GetCategoryDetail
func (s *CategoryService) CreateCategory(req dto.CategoryAddReq, operatorRole int) error {
	// 1. permission check
	if operatorRole != ROLE_ADMIN {
		return errcode.ErrPermissionDenied
	}

	// 2. call dao
	cate := &model.Category{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.categoryDao.CreateCategory(cate); err != nil {
		return errcode.NewSysErr(err)
	}
	return nil

}

// - name or id
func (s *CategoryService) GetCategoryDetail(param string) (*model.Category, error) {
	id, err := strconv.ParseUint(param, 10, 64)
	var cate *model.Category
	var dbErr error
	// 1. parameter check, and get category
	if err == nil && id > 0 {
		// A. ID search
		cate, dbErr = s.categoryDao.GetCategoryByID(id)
	} else {
		// B. Name search
		cate, dbErr = s.categoryDao.GetCategoryByName(param)
	}

	if dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrCategoryNotFound
		}
		return nil, errcode.NewSysErr(dbErr)
	}
	// 2. category operations
	// -2.1 asynchronously increasing

	return cate, nil
}

func (s *CategoryService) GetCategoryList(req dto.CategoryListReq) ([]*dto.CategorySimple, int64, error) {
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
	cates, total, err := s.categoryDao.GetCategoryList(req)
	if err != nil {
		return nil, 0, err
	}

	// 3. post operations

	// 4. return
	return cates, total, nil

}

// ==== DeleteCategory
func (s *CategoryService) DeleteCategory(id uint64, role int) error {
	// 1. validate foreign key
	count, err := s.postDao.CountByCategoryID(id)
	if err != nil {
		return errcode.NewSysErr(err) // 数据库查询出错
	}
	// restrict
	if count > 0 {
		return errcode.ErrCategoryDeleteRefuse
	}

	// 2. check post
	_, err = s.categoryDao.GetCategoryByID(id)
	if err != nil {
		return errcode.ErrCategoryNotFound
	}
	// 3. permission check
	if role != ROLE_ADMIN {
		return errcode.ErrPermissionDenied
	}
	// 4.
	if errDB := s.categoryDao.DeleteCategory(id); errDB != nil {
		return errcode.NewSysErr(errDB)
	}
	return nil
}

// ==== CreateAuthor
// !TODO: 不能单独创建作者了，作者是Admin中ROLE=2的管理员，能够登录后台
// func (s *AuthorService) CreateAuthorAbolished(req dto.AuthorAddReq, operatorRole int) error {
// 	// 1. permission check
// 	if operatorRole != ROLE_ADMIN {
// 		return errors.New("permission denied")
// 	}

// 	// 2. call dao
// 	author := &model.Author{
// 		Name:     req.Name,
// 		Position: req.Position,
// 		Avatar:   req.Avatar,
// 		Mobile:   req.Mobile,
// 		Email:    req.Email,
// 	}
// 	return s.authorDao.CreateAuthor(author)
// }

// - name or id
func (s *AuthorService) GetAuthorDetail(param string) (*model.Author, error) {
	id, err := strconv.ParseUint(param, 10, 64)
	var author *model.Author
	var dbErr error
	// 1. parameter check, and get author
	if err == nil && id > 0 {
		// A. ID search
		author, dbErr = s.authorDao.GetAuthorByID(id)
	} else {
		// B. Name search
		author, dbErr = s.authorDao.GetAuthorByName(param)
	}
	if dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrPostNotFound
		}
		return nil, errcode.NewSysErr(dbErr)
	}

	// 2. author operations
	// -2.1 asynchronously increasing

	return author, nil
}

func (s *AuthorService) GetAuthorList(req dto.AuthorListReq) ([]*dto.AuthorSimple, int64, error) {
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
	authors, total, err := s.authorDao.GetAuthorList(req)
	if err != nil {
		return nil, 0, errcode.NewSysErr(err)
	}

	// 3. post operations

	// 4. return
	return authors, total, nil
}

// ==== DeleteAuthor
// !TODO: 不能单独删除作者了
func (s *AuthorService) DeleteAuthor(id uint64, role int) error {
	// 1. validate foreign key
	count, err := s.postDao.CountByAuthorID(id)
	if err != nil {
		return err // 数据库查询出错
	}
	// restrict
	if count > 0 {
		return errors.New("cannot delete: there are posts associated with this author")
	}

	// 2. check Author
	_, err = s.authorDao.GetAuthorByID(id)
	if err != nil {
		return errors.New("the author doesn't exist")
	}
	// 3. permission check
	if role != ROLE_ADMIN {
		return errors.New("permission denied")
	}
	// 4.
	return s.authorDao.DeleteAuthor(id)
}
