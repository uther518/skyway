package DAO

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"skyway/library/DataSource"
	"skyway/managerapi/model"
	"strconv"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ApiDAO struct {
	client *DataSource.EtcdClient
	apis   map[string]*model.Api
}

func NewApiDao() *ApiDAO {
	return &ApiDAO{
		client: DataSource.GetInstance(),
	}
}

const (
	API_PREFIX     = "API_"
	API_KEY_FORMAT = "API_%d"
)

func getApiKey(apiId int) string {
	return fmt.Sprintf(API_KEY_FORMAT, apiId)
}

/**
 * 注册或更新API
 */
func (apiDao *ApiDAO) RegisterApi(api *model.Api) bool {
	data, err := json.Marshal(api)
	if err == nil {
		apiKey := getApiKey(api.ApiId)
		ret := apiDao.client.Put(apiKey, string(data))
		if ret == true {
			//apiDao.apis[apiKey] = api
		}
		return ret
	}
	return false
}

/**
 * 获取全部API列表
 */
func (apiDao *ApiDAO) GetApis() (map[string]*model.Api, error) {
	if apiDao.apis != nil {
		return apiDao.apis, nil
	}

	apis, err := apiDao.client.GetAll(API_PREFIX)
	if err == nil {
		apiModels := make(map[string]*model.Api)
		for k, v := range apis {
			api := model.NewApi()
			err := json.UnmarshalFromString(v, api)
			if err == nil {
				apiModels[k] = api
			}
		}
		apiDao.apis = apiModels
	}
	return apiDao.apis, err
}

/**
 * 删除指定API
 */
func (apiDao *ApiDAO) DelApi(apiId int) (int64, error) {
	effect, err := apiDao.client.Delete(getApiKey(apiId))
	return effect, err
}

/**
 * 删除所有API
 */
func (apiDao *ApiDAO) DelAllApis() (int64, error) {
	effect, err := apiDao.client.DeleteAll(API_PREFIX)
	return effect, err
}

func (apiDao *ApiDAO) GetMaxId() (int, error) {
	maxKey, err := apiDao.client.GetMaxKey(API_PREFIX)
	if err != nil {
		return -1, err
	}

	idstr := strings.Replace(maxKey, API_PREFIX, "", 1)
	return strconv.Atoi(idstr)
}
