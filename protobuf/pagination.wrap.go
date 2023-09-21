package protobuf

import "gorm.io/gorm"

func PageWrap(p *Pagination) *Pagination {
	if p == nil {
		p = &Pagination{}
	}
	// 如果 Current 传入的参数为空，那么就使用默认值 1
	if p.Current <= 0 {
		p.Current = 1
	}
	// 如果 PageSize 传入的参数为空，那么就使用默认值 20
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return &Pagination{
		Total:    p.Total,
		Current:  p.Current,
		PageSize: p.PageSize,
	}
}

func (p *Pagination) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(p.Limit()).Offset(p.Offset())
	}
}

// Offset 返回分页查询的偏移量, 已自动处理分页数。 min:1, max:unlimited
func (p *Pagination) Offset() int {
	return int((p.Current - 1) * p.PageSize)
}

// Limit 返回分页查询的限制数, 是 PageSize 的别名
func (p *Pagination) Limit() int {
	return int(p.PageSize)
}

// Count 返回分页查询的总记录数 int64 指针
func (p *Pagination) Count() *int64 {
	if p.RawTotal == nil {
		p.RawTotal = new(int64)
	}
	return p.RawTotal
}

func (p *Pagination) Resp() *Pagination {
	p.Total = int32(*p.RawTotal)
	// reset to nil value.
	p.RawTotal = nil
	return p
}
