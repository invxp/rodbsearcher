package main

import (
	"flag"
	"github.com/invxp/rodbsearcher/rodbsearcher"
	"log"
	"os"
	"os/signal"
)

const (
	version = "0.0.1-alpha"
)

const (
	flagGroupItem = "gi"
)

var groupItem = flag.String(flagGroupItem, "group.txt", "set a group item file")

//waitQuit 阻塞等待应用退出(ctrl+c / kill)
func waitQuit() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	log.Printf("rodbsearcher exit")
}

//main 程序入口(尽量少做事)
//1. 解析参数
//2. 判断是否后台执行
//3. 加载配置
//4. 启动应用
//5. 定时任务
//6. 等待退出dd
func main() {
	flag.Parse()

	srv, err := rodbsearcher.New(
		rodbsearcher.WithMySQLConfig(map[string]string{"timeout": "5s"}))

	if err != nil {
		log.Panic(err)
	}

	log.Println("start program version", version)

	srv.LoadItemDB("item_db_etc.yml")
	srv.LoadItemDB("item_db_usable.yml")
	srv.LoadItemDB("item_db_equip.yml")

	srv.ConvertGroupItem(*groupItem)

	waitQuit()
}
