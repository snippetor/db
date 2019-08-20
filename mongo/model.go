package mongo

type Model interface {
	GetId() uint32
}

type BaseModel struct {
	Id uint32 `bson:"_id"`
}

func (m *BaseModel) GetId() uint32 {
	return m.Id
}
