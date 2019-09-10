package hActor

import (
	"custom/happy/hECS"
	"custom/happy/hLog"
)

type ActorBase struct {
	MessageHandler map[string]func(message *ActorMessageInfo) error
	actor          *ActorComponent
	actorType      ActorType
	parent         *hEcs.Object
}

func (this *ActorBase) ActorInit(parent *hEcs.Object, actorType ...ActorType) {
	this.parent = parent
	if len(actorType) == 0 {
		this.actorType = ACTOR_TYPE_DEFAULT
	} else {
		this.actorType = actorType[0]
	}
}

func (this *ActorBase) Actor() *ActorComponent {
	this.panic()
	if this.actor == nil && this.parent != nil {
		err := this.parent.Find(&this.actor)
		if err != nil {
			actor := NewActorComponent(this.actorType)
			this.parent.AddComponent(actor)
			this.actor = actor
		}
	}
	return this.actor
}

func (this *ActorBase) MessageHandlers() map[string]func(message *ActorMessageInfo) error {
	this.panic()
	return this.MessageHandler
}

func (this *ActorBase) AddHandler(service string, handler func(message *ActorMessageInfo) error, isService ...bool) {
	this.panic()
	if this.MessageHandler == nil {
		this.MessageHandler = map[string]func(message *ActorMessageInfo) error{}
	}
	this.MessageHandler[service] = handler
	actor := this.Actor()
	if actor != nil && len(isService) > 0 {
		err := actor.RegisterService(service)
		if err != nil {
			hLog.Error(err)
		}
	}
}

func (this *ActorBase) panic() {
	if this.parent == nil {
		panic("actor base must be initialized first")
	}
}
