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
	Shape  int //形状1-3随机
	Sizeof int //大中小1-3随机
}

type RoomComponent struct {
	BaseComponent
	RoomCrystalInfo RoomCrystalInfo
	RoomID int
	Sid    *[]string
}

func RandNum(num int) int {

	if r := rand.Intn(num); r == 0 {
		return r + 1
	} else {
		return r
	}
}

func RandNumScope(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
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
	totalNum := RandNumScope(5, 8)   //方块总数
	idNum := 0                       //实际方块数
	rand.Seed(time.Now().UnixNano()) //设置随机数种子
	for {
		if idNum == totalNum {
			break
		}
		crystall.Id = RandNum(8)
		crystall.Num = RandNum(3)
		crystall.Angle = RandNum(360)
		crystall.Shape = RandNum(3)
		crystall.Sizeof = RandNum(3)
		slice = append(slice, *crystall)
		idNum++
	}
	return slice

}
