package components

// Match 返回结果
type MatchPlayInfo struct {
	NickName    string
	HeadUrl     string
	Lv          int
	RoomId      int
	IsShootBall bool
	CrystalInfo RoomCrystalInfo
}

//
////生成障碍物
//func (this *RoomComponent) CrystalInfoForMatch() []RoomCrystalInfo {
//	var slice = make([]RoomCrystalInfo, 0)
//	crystall := &RoomCrystalInfo{
//		Id:     0,
//		Num:    0,
//		Angle:  0,
//		Shape:  0,
//		Sizeof: 0,
//	}
//	totalNum := RandNumScope(5, 8)   //方块总数
//	idNum := 0                       //实际方块数
//	rand.Seed(time.Now().UnixNano()) //设置随机数种子
//	for {
//		if idNum == totalNum {
//			break
//		}
//		crystall.Id = RandNum(8)
//		crystall.Num = RandNum(3)
//		crystall.Angle = RandNum(360)
//		crystall.Shape = RandNum(3)
//		crystall.Sizeof = RandNum(3)
//		slice = append(slice, *crystall)
//		idNum++
//	}
//	return slice
//
//}
