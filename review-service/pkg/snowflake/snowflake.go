package snowflake

import (
	"errors"
	"fmt"
	"time"

	sf "github.com/bwmarrin/snowflake"
)

/* 该pkg使用雪花算法生成全局ID
提供GenID方法,可以生成分布式全局不重复的ID
雪花算法初始化需要提供一个应用上线时间戳startTime
*/

// 错误码
var (
	ErrInvalidInitParam  = errors.New("snowflake初始化失败,无效的startTime或machineID")
	ErrInvalidTimeFormat = errors.New("snowflake初始化错误,无效的startTime格式")
)

var node *sf.Node //雪花节点

func Init(startTime string, machineID int64) (err error) {
	// 参数校验,配置文件不支持valiadte?
	if len(startTime) == 0 || machineID <= 0 {
		return ErrInvalidInitParam
	}
	var st time.Time //应用上线时间
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return ErrInvalidTimeFormat
	}
	sf.Epoch = st.UnixNano() / 1000000
	fmt.Println(sf.Epoch)
	node, err = sf.NewNode(machineID) //基于MachineID确定该节点
	return
}

func GenID() int64 {
	return node.Generate().Int64()
}
