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

	UpdatePost(postID uint64, updateData map[string]interface{}, tagIDs []uint64, updateTags bool) error
	UpdatePostStatus(id uint64, status int) error
	DeletePost(id uint64) error
	GetPostList(req dto.PostListReq) ([]*model.Post, int64, error)

	IncrReadCount(id uint64)
	CountByCategoryID(categoryID uint64) (int64, error)
	CountByAuthorID(authorID uint64) (int64, error)
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
	DeleteCategory(id uint64) error
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

type IAuthorDao interface {
	CreateAuthor(author *model.Author) error
	GetAuthorByID(id uint64) (*model.Author, error)
	GetAuthorByName(name string) (*model.Author, error)
	GetAuthorList(req dto.AuthorListReq) ([]*dto.AuthorSimple, int64, error)
	DeleteAuthor(id uint64) error
}

type AuthorDao struct {
	*BaseDao
}

var _ IAuthorDao = (*AuthorDao)(nil)

/* ===== Interface Define End ===== */

func NewPostDao(base *BaseDao) *PostDao {
	return &PostDao{BaseDao: base}
}

func NewCategoryDao(base *BaseDao) *CategoryDao {
	return &CategoryDao{BaseDao: base}
}

func NewTagDao(base *BaseDao) *TagDao {
	return &TagDao{BaseDao: base}
}

func NewAuthorDao(base *BaseDao) *AuthorDao {
	return &AuthorDao{BaseDao: base}
}

/* ===== Post Dao begin ===== */
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
// UpdatePost 动态更新文章及其标签
// postID: 目标文章ID
// updateData: 需要更新的基础字段（利用 map 解决零值被忽略的问题）
// tagIDs: 新的标签ID列表
// updateTags: 标识是否需要更新标签（用于区分“未传标签”和“传入空数组主动清空标签”）
func (d *PostDao) UpdatePost(postID uint64, updateData map[string]interface{}, tagIDs []uint64, updateTags bool) error {
	return d.Orm.Transaction(func(tx *gorm.DB) error {

		// 1. 更新基础字段 (Title, Content, Status 等)
		// 只有当 map 里有数据时才执行主表更新
		if len(updateData) > 0 {
			// 使用 Map 更新：GORM 会忠实地将 map 里的 "" 或 0 更新到数据库，而不会忽略
			if err := tx.Model(&model.Post{}).Where("id = ?", postID).Updates(updateData).Error; err != nil {
				return err
			}
		}

		// 2. 更新标签关联 (Many-to-Many)
		// 只有当明确要求更新标签时（前端传了 Tags 字段，即 updateTags == true）才执行
		if updateTags {
			// 构造一个只包含 ID 的结构体实例，供 GORM 识别主键以进行关联操作
			// （假设你的模型主键字段名为 ID，如果有嵌套如 BaseModel 请按需调整）
			post := &model.Post{BaseModel: model.BaseModel{ID: postID}}

			if len(tagIDs) > 0 {
				var tags []*model.Tag
				for _, id := range tagIDs {
					tags = append(tags, &model.Tag{ID: id})
				}
				// Replace: 会自动查出现有绑定，删除不需要的，插入新增的
				if err := tx.Model(post).Association("Tags").Replace(tags); err != nil {
					return err
				}
			} else {
				// 如果传了空数组（且 updateTags 为 true），说明前端要主动清空所有标签
				if err := tx.Model(post).Association("Tags").Clear(); err != nil {
					return err
				}
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
func (d *PostDao) CountByCategoryID(categoryID uint64) (int64, error) {
	var count int64
	err := d.Orm.Model(&model.Post{}).
		Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}

func (d *PostDao) CountByAuthorID(authorID uint64) (int64, error) {
	var count int64
	err := d.Orm.Model(&model.Post{}).
		Where("author_id = ?", authorID).Count(&count).Error
	return count, err
}

/* ===== Post Dao end ===== */

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

// ==== DeleteCategory
// - soft delete
func (d *CategoryDao) DeleteCategory(id uint64) error {
	// GORM 默认是软删除 (Soft Delete)，因为 model 里有 DeletedAt 字段
	// 如果想要硬删除（物理删除），可以使用 d.Orm.Unscoped().Delete(...)
	return d.Orm.Delete(&model.Category{}, id).Error
}

// ==== GetOrCreateByNames
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

// ==== CreateAuthor
func (d *AuthorDao) CreateAuthor(author *model.Author) error {
	return d.Orm.Create(author).Error
}

// ==== GetCreateAuthor
func (d *AuthorDao) GetAuthorByID(id uint64) (*model.Author, error) {
	var author model.Author
	err := d.Orm.First(&author, id).Error
	return &author, err
}

func (d *AuthorDao) GetAuthorByName(name string) (*model.Author, error) {
	var author model.Author
	err := d.Orm.
		Where("name = ?", name).
		First(&author).Error
	return &author, err
}

func (d *AuthorDao) GetAuthorList(req dto.AuthorListReq) ([]*dto.AuthorSimple, int64, error) {
	var authors []*dto.AuthorSimple
	var total int64

	// 1. init
	db := d.Orm.Model(&model.Author{})

	// possible conditional search

	// 2. count candidates
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 3. 分页、排序并查询指定字段，结果自动映射到 cates (DTO) 中
	offset := (req.Page - 1) * req.PageSize
	err := db.
		Order("name ASC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&authors).Error
	return authors, total, err
}

// ==== DeleteCategory
// - soft delete
func (d *AuthorDao) DeleteAuthor(id uint64) error {
	// GORM 默认是软删除 (Soft Delete)，因为 model 里有 DeletedAt 字段
	// 如果想要硬删除（物理删除），可以使用 d.Orm.Unscoped().Delete(...)
	return d.Orm.Delete(&model.Author{}, id).Error
}
