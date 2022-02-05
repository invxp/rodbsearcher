package rodbsearcher

type Options func(*RODBSearcher)

//WithMySQLConfig 自定义MySQL配置(KV-参考调用示例)
//https://github.com/go-sql-driver/mysql
func WithMySQLConfig(conf map[string]string) Options {
	return func(searcher *RODBSearcher) { searcher.mysqlConf = conf }
}
