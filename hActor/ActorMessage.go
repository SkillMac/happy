package hActor

type ActorMessage struct {
	Service    string
	Data       []interface{}
	IsWaitCall bool
}

func NewActorMessage(service string, args ...interface{}) *ActorMessage {
	return &ActorMessage{
		Service:    service,
		Data:       args,
		IsWaitCall: false,
	}
}
