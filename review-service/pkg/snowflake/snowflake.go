package snowflake

import (
	"errors"
	"time"

	bw "github.com/bwmarrin/snowflake"
)

// Generator 封装 github.com/bwmarrin/snowflake，nodeID 取值 0–1023，多实例部署须保证唯一。
var node *bw.Node

// 初始化雪花算法
func Init(start string, machineID int64) (err error) {
	if len(start) == 0 || machineID < 0 {
		return errors.New("start or machineID is invalid")
	}
	var st time.Time
	st, err = time.Parse(time.RFC3339, start)
	if err != nil {
		return err
	}
	bw.Epoch = st.UnixNano() / 1000000
	node, err = bw.NewNode(machineID)
	return
}

// 生成一个id
func GenID() int64 {
	return node.Generate().Int64()
}
