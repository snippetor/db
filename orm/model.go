package orm

type Model interface {
	GetId() uint32
}

type BaseModel struct {
	Id uint32 `gorm:"primary_key"`
}

func (m *BaseModel) GetId() uint32 {
	return m.Id
}
