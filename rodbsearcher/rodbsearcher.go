package rodbsearcher

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	internal "github.com/invxp/rodbsearcher/internal/http"
	"github.com/invxp/rodbsearcher/internal/util/convert"
	"github.com/invxp/rodbsearcher/internal/util/cron"
	"github.com/invxp/rodbsearcher/internal/util/log"
	"github.com/invxp/rodbsearcher/internal/util/mysql"
	"github.com/invxp/rodbsearcher/internal/util/redis"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//RODBSearcher   - 面向函数编程, 任何对象皆通过New创建
//conf   	- 配置文件
//logger 	- 日志
//redis  	- 缓存/KV
//mysql  	- 数据库/SQL
//mysqlConf - 自定义数据库配置(kv)
//server    - HTTP服务器
//router    - HTTP路由(gin)
//cron      - 定时任务(cron)
//close  	- 关闭应用的channel
type RODBSearcher struct {
	conf      *Config
	logger    *log.Log
	redis     *redis.Redis
	mysql     *mysql.MySQL
	mysqlConf map[string]string
	server    *http.Server
	router    *gin.Engine
	cron      *cron.Cron
	close     chan struct{}
	items     map[int]ItemDB
}

type ItemDBInfo struct {
	Body []ItemDB `yaml:"Body"`
}

type ItemDB struct {
	Id          int    `yaml:"Id"`
	AegisName   string `yaml:"AegisName"`
	Name        string `yaml:"Name"`
	Type        string `yaml:"Type"`
	SubType     string `yaml:"SubType"`
	Buy         int    `yaml:"Buy"`
	Sell        int    `yaml:"Sell"`
	Weight      int    `yaml:"Weight"`
	Attack      int    `yaml:"Attack"`
	MagicAttack int    `yaml:"MagicAttack"`
	Defense     int    `yaml:"Defense"`
	Range       int    `yaml:"Range"`
	Slots       int    `yaml:"Slots"`
	//Jobs
	//Classes
	//Gender
	////Locations
	WeaponLevel   int    `yaml:"WeaponLevel"`
	ArmorLevel    int    `yaml:"ArmorLevel"`
	EquipLevelMin int    `yaml:"EquipLevelMin"`
	EquipLevelMax int    `yaml:"EquipLevelMax"`
	Refineable    bool   `yaml:"Refineable"`
	View          int    `yaml:"View"`
	AliasName     string `yaml:"AliasName"`
	//Flags
	//BuyingStore
	//DeadBranch
	//Container
	//UniqueId
	//BindOnEquip
	//DropAnnounce
	//NoConsume
	//DropEffect
	//Delay
	//Duration
	//Status
	//Stack
	//Amount
	//Inventory
	//Cart
	//Storage
	//GuildStorage
	//NoUse
	//Override
	//Sitting
	//Trade
	//Override
	//NoDrop
	//NoTrade
	//TradePartner
	//NoSell
	//NoCart
	//NoStorage
	//NoGuildStorage
	//NoMail
	//NoAuction
	Script        string `yaml:"Script"`
	EquipScript   string `yaml:"EquipScript"`
	UnEquipScript string `yaml:"UnEquipScript"`
}

type GroupItemList struct {
	Item []GroupItem `yaml:"List"`
}

type GroupItem struct {
	Item      string `yaml:"Item"`
	Rate      int    `yaml:"Rate"`
	Amount    int    `yaml:"Amount"`
	Duration  int    `yaml:"Duration"`
	Announced bool   `yaml:"Announced"`
}

//New 新建一个应用, Options 动态传入配置(可选)
func New(opts ...Options) (*RODBSearcher, error) {
	searcher := &RODBSearcher{close: make(chan struct{}), cron: cron.New(), items: make(map[int]ItemDB, 0)}

	//加载默认Config
	searcher.conf = defaultConfig

	//遍历传入的Options方法
	for _, opt := range opts {
		opt(searcher)
	}

	//加载日志
	if err := searcher.loadLogger(); err != nil {
		return nil, err
	}

	searcher.logf("load log success: %v", searcher.logger)

	return searcher, nil
}

//AddCron 添加一个定时任务
//此处与Linux不同, 格式为: 秒/分/时/月/日/年
func (searcher *RODBSearcher) AddCron(spec string, f func()) {
	searcher.cron.MustAdd(spec, f)
}

//StartCron 开始定时任务
func (searcher *RODBSearcher) StartCron() {
	searcher.cron.Start()
}

//Serv 开启HTTPServer
func (searcher *RODBSearcher) Serv() error {
	if !searcher.conf.HTTP.Enable {
		return fmt.Errorf("http service was not enabled")
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	searcher.router = gin.Default()
	searcher.server = &http.Server{
		Addr:    searcher.conf.HTTP.Address,
		Handler: searcher.router}

	//加载日志中间件(可选)
	searcher.router.Use(searcher.httpStatics())
	//加载校验中间件(可选)
	searcher.router.Use(searcher.httpAuth())

	//默认AnyRouter
	searcher.router.Any(":any", searcher.httpAny)

	//APIRouter
	api := searcher.router.Group(internal.APIPath)
	{
		api.Any(internal.RouteCron, searcher.httpCron)
		api.POST(internal.RouteCreate, searcher.httpCreate)
	}

	searcher.logf("start http service on: %s", searcher.conf.HTTP.Address)

	return searcher.server.ListenAndServe()
}

//Close 优雅的退出应用
func (searcher *RODBSearcher) Close(timeout time.Duration) {
	searcher.cron.Stop()

	close(searcher.close)

	if searcher.server != nil {
		searcher.logf("shutting down http server...")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := searcher.server.Shutdown(ctx); err != nil {
			searcher.logf("http server forced to shutdown: %v", err)
		} else {
			searcher.logf("http server close success")
		}
	}

	if searcher.redis != nil {
		searcher.logf("close redis client: %v", searcher.redis.Close())
	}

	if searcher.mysql != nil {
		searcher.logf("close mysql client: %v", searcher.mysql.Close())
	}
}

//LoadItemDB 加载ItemDB
func (searcher *RODBSearcher) LoadItemDB(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		searcher.panic(err)
	}

	item := &ItemDBInfo{}
	err = yaml.Unmarshal(data, item)

	if err != nil {
		searcher.panic(err)
	}

	for _, item := range item.Body {
		searcher.items[item.Id] = item
	}

	searcher.logf("%s %d item loaded...", filename, len(item.Body))
}

func (searcher *RODBSearcher) ConvertGroupItem(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		searcher.panic(err)
	}

	lines := strings.Split(string(data), "\n")

	gil := &GroupItemList{}
	for _, line := range lines {
		gi := GroupItem{}
		itemArg := strings.Split(line, ",")
		if len(itemArg) == 8 {
			continue
		}
		gi.Item = searcher.items[convert.Atoi(itemArg[0])].AegisName
		if gi.Item == "" {
			searcher.panic(fmt.Sprintf("wrong item: %s", line))
			continue
		}
		gi.Rate = convert.Atoi(itemArg[1])
		gi.Amount = convert.Atoi(itemArg[2])
		gi.Duration = convert.Atoi(itemArg[3])
		if itemArg[4] == "1" {
			gi.Announced = true
		}
		gil.Item = append(gil.Item, gi)
	}
	output, err := yaml.Marshal(gil)
	if err != nil {
		searcher.panic(err)
	}

	searcher.logf("\n%s", string(output))
}
