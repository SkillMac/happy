package components

import (
	"math/rand"
	"time"
)

//房间障碍物信息
type RoomCrystalInfo struct {
	//方块总数 随机5-8个
	Id     int //方块id数，随机1-8
	Num    int //叠加方块的数量随机1-3
	Angle  int //角度随机0-360
	Shape  int //形状1-3随机    1:5, 2 3: 6
	Sizeof int //大中小1-3随机
}

type RoomComponent struct {
	BaseComponent
	RoomCrystalInfo interface{}
	RoomID          int
	Sid             *[]string
}

func (this *RoomComponent) getSid() []string {
	return *this.Sid
}

//生成障碍物
func (this *RoomComponent) CrystalInfo() []RoomCrystalInfo {
	var slice = make([]RoomCrystalInfo, 0)
	crystall := &RoomCrystalInfo{
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
		shape := RandNum(3 + 1)
		angle := 0
		for _, val := range slice {
			if id == val.Id {
				goto loop
			}
		}
		if shape == 1 {
			angle = RandNum(5 + 1)
		}

		if shape == 2 || shape == 3{
			angle = RandNum(6 + 1)
		}
		crystall.Id = id
		crystall.Num = RandNum(3 + 1)
		crystall.Angle = angle
		crystall.Shape =shape
		crystall.Sizeof = RandNum(3 + 1)
		slice = append(slice, *crystall)
		idNum++
	}
	return slice

}
