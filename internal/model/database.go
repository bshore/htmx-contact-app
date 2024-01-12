package model

import "github.com/pocketbase/pocketbase/daos"

type DBClient struct {
	dao *daos.Dao
}

func NewDBClient() *DBClient {
	return &DBClient{}
}

func (d *DBClient) RegisterDao(dao *daos.Dao) {
	d.dao = dao
}
