package dao

import (
	"tomatoBlogDB/dto"
	"tomatoBlogDB/model"

	"gorm.io/gorm"
)

/* ===== Interface Define Begin ===== */
type IPostDao interface {
	CreatePost(post *model.Post) error
	GetPostByID(id uint64) (*model.Post, error)
	GetPostBySlug(slug string) (*model.Post, error)
	GetPostDetailByID(id uint64) (*model.Post, error)
	GetPostDetailBySlug(slug string) (*model.Post, error)

	UpdatePost(post *model.Post, tagIDs []uint64) error
	UpdatePostStatus(id uint64, status int) error
	DeletePost(id uint64) error
	GetPostList(req dto.PostListReq) ([]*model.Post, int64, error)
	IncrReadCount(id uint64)
}

type PostDao struct {
	*BaseDao
}

var _ IPostDao = (*PostDao)(nil)

type ICategoryDao interface {
	CreateCategory(cate *model.Category) error
	GetCategoryByID(id uint64) (*model.Category, error)
	GetCategoryByName(name string) (*model.Category, error)
	GetCategoryList(req dto.CategoryListReq) ([]*dto.CategorySimple, int64, error)
}

type CategoryDao struct {
	*BaseDao
}

var _ ICategoryDao = (*CategoryDao)(nil)

type ITagDao interface {
	GetOrCreateByNames(names []string) ([]*model.Tag, error)
}

type TagDao struct {
	*BaseDao
}

var _ ITagDao = (*TagDao)(nil)

/* ===== Interface Define End ===== */

func NewPostDao() *PostDao {
	return &PostDao{BaseDao: NewBaseDao()}
}

func NewCategoryDao() *CategoryDao {
	return &CategoryDao{BaseDao: NewBaseDao()}
}

func NewTagDao() *TagDao {
	return &TagDao{BaseDao: NewBaseDao()}
}

// ==== CreatePost
// GORM will check post.Tags automatically, if only set ID in post.Tags,
// GORM will associate them in middle table rather than creating a new record.
func (d *PostDao) CreatePost(post *model.Post) error {
	return d.Orm.Create(post).Error
}

// ==== GetPostByID
func (d *PostDao) GetPostByID(id uint64) (*model.Post, error) {
	var post model.Post
	// Preload("Tags") 很有必要，因为更新时如果 Tag 没变，我们可能需要回显
	// 这里只查基础信息用于权限判断，其实不 preload 也行，看业务需求
	err := d.Orm.First(&post, id).Error
	return &post, err
}

func (d *PostDao) GetPostDetailByID(id uint64) (*model.Post, error) {
	var post model.Post
	err := d.Orm.
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		First(&post, id).Error
	return &post, err
}

// ==== GetPostBySlug
func (d *PostDao) GetPostBySlug(slug string) (*model.Post, error) {
	var post model.Post
	err := d.Orm.
		Where("slug = ?", slug).
		First(&post).Error
	return &post, err
}

func (d *PostDao) GetPostDetailBySlug(slug string) (*model.Post, error) {
	var post model.Post
	err := d.Orm.Preload("Author").Preload("Category").Preload("Tags").
		Where("slug = ?", slug).
		First(&post).Error
	return &post, err
}

// ==== UpdatePost
func (d *PostDao) UpdatePost(post *model.Post, tagIDs []uint64) error {
	return d.Orm.Transaction(func(tx *gorm.DB) error {
		// 1. 更新基础字段 (Title, Content, etc.)
		// 使用 Save 会保存所有字段，包括零值；使用 Updates 只更新非零值
		// 这里推荐用 Model+Updates 指定 ID 进行更新
		if err := tx.Model(&model.Post{}).Where("id = ?", post.ID).Updates(post).Error; err != nil {
			return err
		}

		// 2. 更新标签关联 (Many-to-Many)
		// 这一步会自动：删除旧关联 + 插入新关联
		if len(tagIDs) > 0 {
			var tags []*model.Tag
			for _, id := range tagIDs {
				tags = append(tags, &model.Tag{ID: id})
			}
			// 替换关联
			if err := tx.Model(post).Association("Tags").Replace(tags); err != nil {
				return err
			}
		} else {
			// 如果传了空数组，意味着清空所有标签
			if err := tx.Model(post).Association("Tags").Clear(); err != nil {
				return err
			}
		}

		return nil
	})
}

// ==== UpdatePostStatus
func (d *PostDao) UpdatePostStatus(id uint64, status int) error {
	// UpdateColumn 生成的 SQL 类似: UPDATE posts SET status = ? WHERE id = ?
	// 它不会更新 UpdatedAt 时间戳（如果业务要求更新时间戳，改用 Updates）
	return d.Orm.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("status", status).Error
}

// ==== DeletePost
// - soft delete
func (d *PostDao) DeletePost(id uint64) error {
	// GORM 默认是软删除 (Soft Delete)，因为 model 里有 DeletedAt 字段
	// 如果想要硬删除（物理删除），可以使用 d.Orm.Unscoped().Delete(...)
	return d.Orm.Delete(&model.Post{}, id).Error
}

// ==== GetPostList
func (d *PostDao) GetPostList(req dto.PostListReq) ([]*model.Post, int64, error) {
	var posts []*model.Post
	var total int64

	// 1. init
	db := d.Orm.Model(&model.Post{})

	// 2. prepare query conditions
	// -2.1 status(allow 0)
	if req.Status != nil {
		db = db.Where("sys_post.status = ?", *req.Status)
	}

	// -2.2 category
	if req.CategoryID > 0 {
		db = db.Where("sys_post.category_id = ?", req.CategoryID)
	}

	// -2.3 keywords (search in Title and summary)
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("sys_post.title LIKE ? OR sys_post.summary LIKE ?", keyword, keyword)
	}

	// -2.4 tags (JOIN post_tags table)
	if req.TagID > 0 {
		db = db.Joins("JOIN post_tags pt ON pt.post_id = sys_post.id").
			Where("pt.tag_id = ?", req.TagID)
	}

	// 3. count candidates
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 4. performance optimization
	// 列表页绝对不需要加载 Content (几千字)，只查必要的字段
	// 这样能减少 MySQL IO 和网络传输带宽，提升 10倍+ 性能
	db = db.Omit("Content")

	// 5. Preload
	db = db.Preload("Author").Preload("Category").Preload("Tags")

	// 6. paginate and order
	offset := (req.Page - 1) * req.PageSize
	err := db.Order("sys_post.created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&posts).Error
	return posts, total, err
}

// ==== IncrReadCount
// - atomic operation
// - don't update updated_at filed
func (d *PostDao) IncrReadCount(id uint64) {
	// SQL: UPDATE sys_post SET read_count = read_count + 1 WHERE id = ?
	d.Orm.Model(&model.Post{}).
		Where("id = ?", id).
		UpdateColumn("read_count", gorm.Expr("read_count + ?", 1))
}

// ==== CreateCategory
func (d *CategoryDao) CreateCategory(cate *model.Category) error {
	return d.Orm.Create(cate).Error
}

// ==== GetCategoryByID
func (d *CategoryDao) GetCategoryByID(id uint64) (*model.Category, error) {
	var cate model.Category
	err := d.Orm.First(&cate, id).Error
	return &cate, err
}

func (d *CategoryDao) GetCategoryByName(name string) (*model.Category, error) {
	var cate model.Category
	err := d.Orm.
		Where("name = ?", name).
		First(&cate).Error
	return &cate, err
}

// ==== GetCategoryList
func (d *CategoryDao) GetCategoryList(req dto.CategoryListReq) ([]*dto.CategorySimple, int64, error) {
	var cates []*dto.CategorySimple
	var total int64

	// 1. init
	db := d.Orm.Model(&model.Category{})

	// 如果你有搜索条件，应该在这里加上，例如：
	// if req.Keyword != "" {
	//     db = db.Where("name LIKE ?", "%"+req.Keyword+"%")
	// }

	// 2. count candidates
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 3. 分页、排序并查询指定字段，结果自动映射到 cates (DTO) 中
	offset := (req.Page - 1) * req.PageSize
	err := db.Select("id", "name", "description").
		Order("name ASC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&cates).Error
	return cates, total, err
}

func (d *TagDao) GetOrCreateByNames(names []string) ([]*model.Tag, error) {
	var tags []*model.Tag
	for _, name := range names {
		// 过滤空字符串防呆
		if name == "" {
			continue
		}

		var tag model.Tag
		// 纯粹的数据库操作留在 DAO 层
		err := d.Orm.Where(model.Tag{Name: name}).FirstOrCreate(&tag).Error
		if err != nil {
			return nil, err
		}

		tags = append(tags, &tag)
	}

	return tags, nil
}
