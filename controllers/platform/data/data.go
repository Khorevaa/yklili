package data

import (
	"github.com/sinmahod/yklili/util"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"

	"github.com/astaxie/beego"
)

type DataController struct {
	beego.Controller
	//每页行数
	PageSize int
	//请求的页码
	PageIndex int
	//排序字段
	OrderColumn string
	//升降序
	OrderSord string
	//方法名
	MethodName string
	//存放字段和值的数据
	RequestData map[string]interface{}
	//存放上传的文件
	FileMap map[string][]*multipart.FileHeader
	//返回数据
	ResponseData map[string]interface{}
}

func (c *DataController) Get() {
	//得到方法名，利用反射机制获取结构体
	value := reflect.ValueOf(c.AppController)
	//判断结构中是否存在方法，存在则执行
	if v := value.MethodByName(c.MethodName); v.IsValid() {
		v.Call(nil)
	} else {
		c.methodNotFind()
	}
}

/**
*   准备方法，得到页码等信息
**/
func (c *DataController) Prepare() {
	c.PageSize, _ = c.GetInt("rows")
	c.PageIndex, _ = c.GetInt("page")
	c.OrderColumn = c.GetString("sidx")
	c.OrderSord = c.GetString("sord")
	c.MethodName = c.GetString(":method")

	if err := c.Ctx.Request.ParseForm(); err != nil {
		beego.Info(err)
	}
	c.RequestData = make(map[string]interface{})
	for k, v := range c.Ctx.Request.Form {
		if len(v) > 0 {
			beego.Info(k, "=========", v)
			c.RequestData[k] = v[0]
		}
	}

	c.RequestData["User"] = c.GetSession("User")

	c.Data["Request"] = c.RequestData

	//文件上传
	c.FileMap = make(map[string][]*multipart.FileHeader)
	if c.Ctx.Request.MultipartForm != nil && c.Ctx.Request.MultipartForm.File != nil {
		filemap := c.Ctx.Request.MultipartForm.File
		for k, v := range filemap {
			beego.Info("===========", k, v)
			reg, _ := regexp.Compile(`\[[\d]+?\]`)
			k = reg.ReplaceAllString(k, "")

			if files, ok := c.FileMap[k]; ok {
				c.FileMap[k] = append(files, v...)
			} else {
				c.FileMap[k] = v
			}
		}
	}

	c.ResponseData = make(map[string]interface{})
}

func (c *DataController) methodNotFind() {
	http.Error(c.Ctx.ResponseWriter, c.MethodName+" 方法未找到", 404)
}

func (c *DataController) paramIsNull() {
	http.Error(c.Ctx.ResponseWriter, "参数为空", 510)
}

const (
	Script string = "<script>$('[data-rel=tooltip]').tooltip({container:'body'});</script>"
)

//在模板结尾追加js
func (c *DataController) addScript() error {
	rb, err := c.RenderBytes()
	if err != nil {
		return err
	}

	//替换标签 增加校验
	html, err := util.AnalysisGoTag(string(rb))

	if err != nil {
		return err
	}

	c.Ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	c.Ctx.Output.Body([]byte(html))
	return nil
}

//普通的put方法
func (c *DataController) put(key string, val interface{}) {
	c.ResponseData[key] = val
}

//成功
func (c *DataController) success(message string) {
	c.ResponseData["STATUS"] = 1
	c.ResponseData["MESSAGE"] = message
}

//失败
func (c *DataController) fail(message string) {
	c.ResponseData["STATUS"] = 0
	c.ResponseData["MESSAGE"] = message
}

//重写ServeJson，增加将ResponseData写入json的操作
func (c *DataController) ServeJSON(encoding ...bool) {
	if c.Data["json"] == nil {
		c.Data["json"] = c.ResponseData
	}
	var (
		hasIndent   = true
		hasEncoding = false
	)
	if beego.BConfig.RunMode == beego.PROD {
		hasIndent = false
	}
	if len(encoding) > 0 && encoding[0] == true {
		hasEncoding = true
	}
	c.Ctx.Output.JSON(c.Data["json"], hasIndent, hasEncoding)
}

type Validator struct {
	Valid bool `json:"valid"`
}
