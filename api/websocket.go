package api

import (
	"wolszon.me/groupie/domain"
	"github.com/gorilla/websocket"
	"sync"
	"gopkg.in/mgo.v2/bson"
)

type (
	Client struct {
		GroupId      string
		SecurityPile domain.SecurityPile
		Conn         *websocket.Conn
		Mutex        *sync.Mutex
	}

	RequestContext struct {
		*Client
		Action  string
		Payload map[string]interface{}
	}
)

func (c *Client) SendUpdate(group *domain.Group) {
	c.send(bson.M{
		"type":    "update",
		"payload": group.Export(),
	})
}

func (c *Client) HandleError(err error) {
	if err == domain.NoSufficientPermissions {
		c.SendError(1, err.Error())
	} else if err == domain.MemberNotExists {
		c.SendError(2, err.Error())
	} else if err == domain.GroupNotExists {
		c.SendError(3, err.Error())
	} else {
		c.SendInternalServerError()
	}
}

func (c *Client) SendInternalServerError() {
	c.SendError(-1, "Internal server error")
}

func (c *Client) SendError(code int, message string) {
	c.send(bson.M{
		"type": "error",
		"payload": bson.M{
			"code":    code,
			"message": message,
		},
	})
}

func (c *Client) send(body interface{}) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.Conn.WriteJSON(body)
}

func (c *Client) CreateRequestContext(body map[string]interface{}) RequestContext {
	request := RequestContext{
		Action: body["action"].(string),
		Client: c,
	}

	if payload, ok := body["payload"]; ok {
		request.Payload = payload.(map[string]interface{})
	}

	return request
}
