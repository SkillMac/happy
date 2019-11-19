package messageProtocol

import (
	"github.com/json-iterator/go"
)

type JsonProtocol struct {
}

func NewJsonProtocol() *JsonProtocol {
	return &JsonProtocol{}
}

func (this *JsonProtocol) Marshal(message interface{}) ([]byte, error) {
	return jsoniter.Marshal(message) //json.Marshal(message)
}

func (this *JsonProtocol) Unmarshal(data []byte, messageType interface{}) error {
	return jsoniter.Unmarshal(data, &messageType) //json.Unmarshal(data, &messageType)
}
