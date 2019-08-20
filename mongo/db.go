package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
)

type DataBase interface {
	DefaultDB() *mongo.Database
	DefaultClient() *mongo.Client
	Collection(string, ...*options.CollectionOptions) *mongo.Collection
	CollectionWithModel(interface{}, ...*options.CollectionOptions) *mongo.Collection

	CreateOne(Model) (interface{}, error)
	CreateMany(...interface{}) ([]interface{}, error)
	FindAll(interface{}, ...*options.FindOptions) error
	FindMany(interface{}, Model, ...*options.FindOptions) error
	FindOne(Model, ...*options.FindOneOptions) error
	DeleteAll(Model, ...*options.DeleteOptions) (int64, error)
	DeleteMany(Model, ...*options.DeleteOptions) (int64, error)
	DeleteOne(Model, ...*options.DeleteOptions) (int64, error)
	UpdateMany(Model, Model, ...*options.UpdateOptions) (int64, interface{}, error)
	UpdateOne(Model, Model, ...*options.UpdateOptions) (int64, interface{}, error)

	Close() error
}

func findStruct(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		return findStruct(t.Elem())
	} else if t.Kind() == reflect.Struct {
		return t
	}
	return t
}

func getStruct(i interface{}) reflect.Type {
	return findStruct(reflect.TypeOf(i))
}

func getStructName(i interface{}) string {
	return getStruct(i).Name()
}

func Connect(addr, user, pwd, defaultDb string) (DataBase, error) {
	//[mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	//mongodb://myuser:mypass@localhost:40001,otherhost:40001/
	var uri string
	if user != "" && pwd != "" {
		uri = "mongodb://" + user + ":" + pwd + "@" + addr + "/" + defaultDb
	} else {
		uri = "mongodb://" + addr + "/" + defaultDb
	}
	c, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &database{c, c.Database(defaultDb)}, nil
}

type database struct {
	client    *mongo.Client
	defaultDb *mongo.Database
}

func (m *database) DefaultDB() *mongo.Database {
	return m.defaultDb
}

func (m *database) DefaultClient() *mongo.Client {
	return m.client
}

func (m *database) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return m.defaultDb.Collection(name, opts...)
}

// @models: single model or slice, or point to model or slice
func (m *database) CollectionWithModel(models interface{}, opts ...*options.CollectionOptions) *mongo.Collection {
	return m.defaultDb.Collection(getStructName(models), opts...)
}

func (m *database) CreateOne(model Model) (interface{}, error) {
	ret, err := m.CollectionWithModel(model).InsertOne(context.TODO(), model)
	if err != nil {
		return nil, err
	}
	return ret.InsertedID, nil
}

// @models must same type
func (m *database) CreateMany(models ...interface{}) ([]interface{}, error) {
	ret, err := m.CollectionWithModel(models).InsertMany(context.TODO(), models)
	if err != nil {
		return nil, err
	}
	return ret.InsertedIDs, nil
}

// @models: must *[]interface
func (m *database) FindAll(models interface{}, opts ...*options.FindOptions) error {
	return m.FindMany(models, nil, opts...)
}

// @models: must *[]interface
func (m *database) FindMany(models interface{}, filter Model, opts ...*options.FindOptions) error {
	t := getStruct(models)
	c := m.Collection(t.Name())
	var bs []byte
	if filter != nil {
		var err error
		bs, err = bson.Marshal(filter)
		if err != nil {
			return err
		}
	}
	cur, err := c.Find(context.TODO(), bs, opts...)
	if err != nil {
		return err
	}
	valuePtr := reflect.ValueOf(models)
	value := valuePtr.Elem()
	for cur.Next(context.TODO()) {
		m := reflect.New(t).Interface()
		if err := cur.Decode(m); err != nil {
			return err
		}
		value.Set(reflect.Append(value, reflect.ValueOf(m).Elem()))
	}
	return nil
}

func (m *database) FindOne(model Model, opts ...*options.FindOneOptions) error {
	bs, err := bson.Marshal(model)
	if err != nil {
		return err
	}
	err = m.CollectionWithModel(model).FindOne(context.TODO(), bs, opts...).Decode(model)
	if err != nil {
		return err
	}
	return nil
}

func (m *database) DeleteAll(model Model, opts ...*options.DeleteOptions) (int64, error) {
	return m.DeleteMany(model, opts...)
}

func (m *database) DeleteMany(model Model, opts ...*options.DeleteOptions) (int64, error) {
	bs, err := bson.Marshal(model)
	if err != nil {
		return 0, err
	}
	ret, err := m.CollectionWithModel(model).DeleteMany(context.TODO(), bs, opts...)
	if err != nil {
		return 0, err
	}
	return ret.DeletedCount, nil
}

func (m *database) DeleteOne(model Model, opts ...*options.DeleteOptions) (int64, error) {
	bs, err := bson.Marshal(model)
	if err != nil {
		return 0, err
	}
	ret, err := m.CollectionWithModel(model).DeleteOne(context.TODO(), bs, opts...)
	return ret.DeletedCount, err
}

func (m *database) UpdateMany(filterModel, updateModel Model, opts ...*options.UpdateOptions) (int64, interface{}, error) {
	filter, err := bson.Marshal(filterModel)
	if err != nil {
		return 0, 0, err
	}
	update, err := bson.Marshal(updateModel)
	if err != nil {
		return 0, 0, err
	}
	ret, err := m.CollectionWithModel(filterModel).UpdateMany(context.TODO(), filter, update, opts...)
	if err != nil {
		return 0, 0, err
	}
	return ret.UpsertedCount, ret.UpsertedID, nil
}

func (m *database) UpdateOne(filterModel, updateModel Model, opts ...*options.UpdateOptions) (int64, interface{}, error) {
	filter, err := bson.Marshal(filterModel)
	if err != nil {
		return 0, 0, err
	}
	update, err := bson.Marshal(updateModel)
	if err != nil {
		return 0, 0, err
	}
	ret, err := m.CollectionWithModel(filterModel).UpdateOne(context.TODO(), filter, update, opts...)
	if err != nil {
		return 0, 0, err
	}
	return ret.UpsertedCount, ret.UpsertedID, nil
}

func (m *database) Close() error {
	if m.client != nil {
		return m.client.Disconnect(context.TODO())
	}
	return nil
}
