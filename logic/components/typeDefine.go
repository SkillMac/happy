package components

import (
	"math/rand"
	"time"
)

// Match 返回结果
type MatchPlayInfo struct {
	NickName    string
	HeadUrl     string
	Lv          int
	RoomId      int
	IsShootBall bool
	CrystalInfo interface{}
}
type matchCrystalInfo struct {
	//方块总数 随机5-8个
	Id     int //方块id数，随机1-8
	Num    int //叠加方块的数量随机1-3
	Angle  int //角度随机0-360
	Shape  int //形状1-3随机
	Sizeof int //大中小1-3随机
}

//生成障碍物
func (this *MatchPlayInfo) MatchCrystalInfo() []matchCrystalInfo {
	var slice = make([]matchCrystalInfo, 0)
	crystall := &matchCrystalInfo{
		Id:     0,
		Num:    0,
		Angle:  0,
		Shape:  0,
		Sizeof: 0,
	}
	totalNum := RandNumScope(5, 8+1) //方块总数
	idNum := 0                       //实际方块数
	rand.Seed(time.Now().UnixNano()) //设置随机数种子
	for {
	loop:
		if idNum == totalNum {
			break
		}
		id := RandNum(8 + 1)

		for _, val := range slice {
			//fmt.Println("id,val.id=== ", id, val.id)
			if id == val.Id {
				goto loop
			}
		}
		crystall.Id = id
		crystall.Num = RandNum(3 + 1)
		crystall.Angle = RandNum(360 + 1)
		crystall.Shape = RandNum(3 + 1)
		crystall.Sizeof = RandNum(3 + 1)
		slice = append(slice, *crystall)
		idNum++
	}
	return slice

}
