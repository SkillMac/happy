package components

type RoomComponent struct {
	BaseComponent
	RoomID int
	Sid    *[]string
}

func (this *RoomComponent) getSid() []string {
	return *this.Sid
}
