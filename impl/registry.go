package impl

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/config"
	"net/http"
)

var harborURL = "https://" + config.HarborURL + "/api"

func ListProjects(c *gin.Context) {
	responseData := common.ResponseData{}
	url := harborURL + "/projects/"
	request := HttpRequest(url)

	var someOne []common.Projects
	if err := json.Unmarshal(request, &someOne); err == nil {
		responseData.Msg = ""
		responseData.Data = someOne
		responseData.Code = http.StatusOK
	} else {
		responseData.Msg = err.Error()
		responseData.Data = ""
		responseData.Code = http.StatusInternalServerError
		log.Errorf("harbor/projects get: %s", err)
	}
	c.JSON(responseData.Code, responseData)
}

func ListImages(c *gin.Context) {
	id := c.Param("id")
	url := harborURL + "/repositories?project_id=" + id
	responseData := common.ResponseData{}
	request := HttpRequest(url)
	var someOne []common.ImageList
	if err := json.Unmarshal(request, &someOne); err == nil {
		responseData.Msg = ""
		responseData.Data = someOne
		responseData.Code = http.StatusOK
	} else {
		responseData.Msg = err.Error()
		responseData.Data = ""
		responseData.Code = http.StatusInternalServerError
		log.Errorf("harbor/images get :%s", err)
	}
	c.JSON(responseData.Code, responseData)
}

func ListTags(c *gin.Context) {
	project := c.Param("project")
	image := c.Param("image")
	url := harborURL + "/repositories/" + project + "/" + image + "/tags"
	responseData := common.ResponseData{}
	request := HttpRequest(url)
	var someOne []common.ImageTags
	if err := json.Unmarshal(request, &someOne); err == nil {
		responseData.Msg = ""
		responseData.Data = someOne
		responseData.Code = http.StatusOK
	} else {
		responseData.Msg = err.Error()
		responseData.Data = ""
		responseData.Code = http.StatusInternalServerError
		log.Errorf("harbor/images get :%s", err)
	}
	c.JSON(responseData.Code, responseData)
}

func HttpRequest(url string) []byte {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}
