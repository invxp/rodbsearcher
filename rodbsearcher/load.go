package rodbsearcher

import (
	"fmt"
	"github.com/invxp/rodbsearcher/internal/util/log"
	"reflect"
	"strings"
)

//loadLogger 加载日志服务, 带自动轮转
func (searcher *RODBSearcher) loadLogger() error {
	if !searcher.conf.Log.Enable {
		return nil
	}

	var err error
	searcher.logger, err = log.New(
		searcher.conf.Log.Path,
		fmt.Sprintf("%s.log", strings.ToLower(reflect.TypeOf(*searcher).Name())),
		searcher.conf.Log.MaxAgeHours,
		searcher.conf.Log.MaxRotationMegabytes)

	return err
}
