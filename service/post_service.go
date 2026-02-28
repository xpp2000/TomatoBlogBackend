package service

import (
	"errors"
	"strconv"
	"time"
	"tomatoBlogDB/dao"
	"tomatoBlogDB/dto"
	"tomatoBlogDB/model"

	"github.com/gosimple/slug"
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
}

func NewCategoryService(CategoryDao dao.ICategoryDao) *CategoryService {
	return &CategoryService{
		categoryDao: CategoryDao,
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
		return nil, dbErr
	}

	// 2. assembly respPost
	var tagNames []string
	for _, tag := range post.Tags {
		tagNames = append(tagNames, tag.Name)
	}
	var author = dto.AuthorSimple{
		ID:       post.Author.ID,
		Name:     post.Author.Name,
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
// NOTE: For writer and super admin
func (s *PostService) CreatePost(req dto.PostAddReq, authorID uint64) error {
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
			return err // 如果处理标签时数据库报错，直接抛出
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
		} else {
			// 如果你前端喜欢传 "2006-01-02 15:04:05" 这种普通格式，可以在这里加个 fallback
			if t2, err2 := time.Parse(time.DateTime, req.PublishedAt); err2 == nil {
				FinalPublishedAt = t2
			}
		}
	}

	// 处理“早于当前时间”的逻辑
	if FinalPublishedAt.Before(now) {
		FinalPublishedAt = now
	}
	// 4. assemble model
	post := model.Post{
		Title:       req.Title,
		Slug:        finalSlug,
		Subtitle:    req.Subtitle,
		Summary:     req.Summary,
		Keywords:    req.Keywords,
		Content:     req.Content,
		Cover:       req.Cover,
		AuthorID:    authorID,
		CategoryID:  req.CategoryID,
		Tags:        tags,
		Status:      1,
		PublishedAt: FinalPublishedAt,
	}

	// 3.
	return s.postDao.CreatePost(&post)
}

// ==== UpdatePost
// operatorID:
// operatorRole:
func (s *PostService) UpdatePost(req dto.PostUpdateReq, operatorID uint64, operatorRole int) error {
	// 1. check post
	post, err := s.postDao.GetPostByID(req.ID)
	if err != nil {
		return errors.New("the post doesn't exist or has been deleted")
	}

	// 2. permission check
	if operatorRole != ROLE_ADMIN && post.AuthorID != operatorID {
		return errors.New("insufficient permission")
	}

	// 3. prepare data to be updated
	// 注意：这里只赋值需要更新的字段，AuthorID 不需要赋值（禁止转让作者）

	// 3.1 处理 TagsID（去数据库查找或创建标签）
	var tags []*model.Tag // 提前声明，如果没有传入标签，它就是 nil
	if len(req.Tags) > 0 {
		var err error
		tags, err = s.tagDao.GetOrCreateByNames(req.Tags)

		if err != nil {
			return err // 如果处理标签时数据库报错，直接抛出
		}
	}
	var tagsID []uint64
	for _, tag := range tags {
		tagsID = append(tagsID, tag.ID)
	}

	finalSlug := req.Slug
	if finalSlug == "" {
		finalSlug = slug.Make(req.Title)
	}
	updateData := model.Post{
		BaseModel:  model.BaseModel{ID: req.ID}, // 必须指定 ID
		Title:      req.Title,
		Subtitle:   req.Subtitle,
		Summary:    req.Summary,
		Keywords:   req.Keywords,
		Content:    req.Content,
		Cover:      req.Cover,
		Slug:       finalSlug, // 如果允许修改 Slug
		CategoryID: req.CategoryID,
		Status:     req.Status,
	}

	return s.postDao.UpdatePost(&updateData, tagsID)
}

// ==== UpdateStatus
func (s *PostService) UpdateStatus(req dto.PostStatusReq, operatorID uint64, operatorRole int) error {
	// 1. search the post
	post, err := s.postDao.GetPostByID(req.ID)
	if err != nil {
		return errors.New("The post doesn't exist")
	}

	// 2. permission check
	if operatorRole == ROLE_ADMIN {
		// = admin
	} else {
		// = user
		// B1.
		if post.AuthorID != operatorID {
			return errors.New("permission deny")
		}
		// B2.
		if post.Status == 2 {
			return errors.New("the post has been locked, please let admin unlock it")
		}
	}

	// 3.  execute
	err = s.postDao.UpdatePostStatus(req.ID, req.Status)
	if err != nil {
		return err
	}

	return nil
}

// ==== DeletePost
func (s *PostService) DeletePost(id uint64, operatorID uint64, operatorRole int) error {
	// 1. check post
	post, err := s.postDao.GetPostByID(id)
	if err != nil {
		return errors.New("the post doesn't exist")
	}

	// 2. permission check
	if operatorID != ROLE_ADMIN && post.AuthorID != operatorID {
		return errors.New("permission denied")
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
			Name:     p.Author.Name,
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
		return nil, dbErr
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
