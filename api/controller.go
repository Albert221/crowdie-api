package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gbrlsnchs/jwt"
	"github.com/google/logger"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"wolszon.me/groupie/domain"
	"github.com/gorilla/websocket"
	"sync"
)

type ApiController struct {
	repository   domain.Repository
	tokenManager TokenManager
	upgrader     websocket.Upgrader
	clients      map[string][]*Client // [groupId => [conn, conn]]
	clientsMutex sync.Mutex
}

func NewApiController(repository domain.Repository, manager TokenManager) ApiController {
	return ApiController{
		repository:   repository,
		tokenManager: manager,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[string][]*Client),
	}
}

func (c *ApiController) GetMiddleware() mux.MiddlewareFunc {
	return c.tokenManager.TokenMiddleware
}

func (c *ApiController) NewGroup(w http.ResponseWriter, r *http.Request) {
	creator := createMemberFromRequest(r)

	group, err := c.repository.CreateGroup(creator)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := c.tokenManager.CreateToken(creator.Secret, creator.AndroidId)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := bson.M{
		"yourId": creator.Id,
		"group":  group.Export(),
		"token":  token,
	}

	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

func (c *ApiController) JoinGroup(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]
	member := createMemberFromRequest(r)

	group, err := c.repository.AddMemberToGroup(id, member)
	if err == domain.GroupNotExists {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err == domain.NoSufficientPermissions {
		w.WriteHeader(http.StatusForbidden)
		return
	} else if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := c.tokenManager.CreateToken(member.Secret, member.AndroidId)
	if err != nil {
		logger.Warning(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := bson.M{
		"group":  group.Export(),
		"yourId": member.Id,
		"token":  token,
	}

	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

func (c *ApiController) Endpoint(w http.ResponseWriter, r *http.Request) {
	secPile, err := getSecurityPileFromJWT(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	v := mux.Vars(r)
	id := v["id"]
	_, err = c.repository.GetGroupById(id, secPile)
	if err == domain.GroupNotExists {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err == domain.NoSufficientPermissions {
		w.WriteHeader(http.StatusForbidden)
		return
	} else if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("error while upgrading: %s", err)
		return
	}

	connection := &Client{
		GroupId:      id,
		SecurityPile: secPile,
		Conn:         conn,
		Mutex:        &sync.Mutex{},
	}

	c.clientsMutex.Lock()
	c.clients[id] = append(c.clients[id], connection)
	connectionId := len(c.clients[id]) - 1
	c.clientsMutex.Unlock()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			logger.Errorf("error while reading message: %s", err)
			break
		}

		var requestBody map[string]interface{}
		if bson.UnmarshalJSON(p, &requestBody) != nil {
			continue
		}

		request := connection.CreateRequestContext(requestBody)

		switch request.Action {
		case "get":
			go c.get(request)
		case "send_coordinates":
			go c.sendCoordinates(request)
		case "update_role":
			go c.updateRole(request)
		case "kick":
			go c.kick(request)
		}
	}

	c.clientsMutex.Lock()
	c.clients[id] = append(c.clients[id][:connectionId], c.clients[id][connectionId+1:]...)
	c.clientsMutex.Unlock()
}

func (c *ApiController) get(request RequestContext) {
	fmt.Println("get lol")

	group, err := c.repository.GetGroupById(request.GroupId, request.SecurityPile)
	if err != nil {
		request.HandleError(err)
		return
	}

	go request.SendUpdate(group)
}

func (c *ApiController) sendCoordinates(request RequestContext) {
	if request.Payload == nil {
		return
	}

	memberId := request.Payload["memberId"].(string)
	coords := domain.CoordsBit{
		Lat:  float32(request.Payload["lat"].(float64)),
		Lng:  float32(request.Payload["lng"].(float64)),
		Time: time.Now(),
	}

	group, err := c.repository.UpdateMemberCoordsBit(memberId, coords, request.SecurityPile)
	if err != nil {
		request.HandleError(err)
		return
	}

	go c.sendUpdateToGroup(request.GroupId, group)
}

func (c *ApiController) updateRole(request RequestContext) {
	memberId := request.Payload["memberId"].(string)
	role := int8(request.Payload["role"].(float64))

	group, err := c.repository.UpdateMemberRole(memberId, int8(role), request.SecurityPile)
	if err != nil {
		request.HandleError(err)
		return
	}

	go c.sendUpdateToGroup(request.GroupId, group)
}

func (c *ApiController) kick(request RequestContext) {
	memberId := request.Payload["memberId"].(string)

	group, err := c.repository.KickMember(memberId, request.SecurityPile)
	if err != nil {
		request.HandleError(err)
		return
	}

	go c.sendUpdateToGroup(request.GroupId, group)
}

func (c *ApiController) sendUpdateToGroup(groupId string, group *domain.Group) {
	fmt.Println("sending to whole group")

	if clients, ok := c.clients[groupId]; ok {
		for _, client := range clients {
			client.SendUpdate(group)
		}
	}
}

func createMemberFromRequest(r *http.Request) domain.Member {
	member := make(map[string]interface{})

	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &member)

	name := member["name"].(string)
	androidId := member["androidId"].(string)
	lat := float32(member["lat"].(float64))
	lng := float32(member["lng"].(float64))

	return domain.NewMember(name, androidId, lat, lng)
}

func getSecurityPileFromJWT(r *http.Request) (domain.SecurityPile, error) {
	t := r.Context().Value(ContextJwt)
	if t == nil {
		return domain.SecurityPile{}, fmt.Errorf("can't retrieve JWT from RequestContext")
	}

	token := t.(*jwt.JWT)
	public := token.Public()

	secret, secretExists := public["secret"]
	androidId, androidIdExists := public["androidId"]
	if !secretExists || !androidIdExists {
		return domain.SecurityPile{}, fmt.Errorf("JWT payload lacks secret and/or androidId")
	}

	return domain.SecurityPile{
		Secret:    secret.(string),
		AndroidId: androidId.(string),
	}, nil
}
