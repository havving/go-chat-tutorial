package modules

import (
	"github.com/julienschmidt/httprouter"
	"github.com/mholt/binding"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

var (
	MongoSession mgo.Session
	Renderer     *render.Render
)

type Room struct {
	ID   bson.ObjectId `bson:"_id" json:"id"`
	Name string        `bson:"name" json:"name"`
}

func (r *Room) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{&r.Name: "name"}
}

// 채팅방 정보 생성
func CreateRoom(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	// binding 패키지로 room 생성 요청 정보를 Room 타입 값으로 변환
	r := new(Room)
	errs := binding.Bind(req, r)
	if errs.Handle(w) {
		return
	}

	// 몽고DB 세션 생성
	session := MongoSession.Copy()
	// 몽고DB 세션을 닫는 코드를 defer로 등록
	defer session.Close()

	// 몽고DB ID 생성
	r.ID = bson.NewObjectId()
	// room 정보 저장을 위한 몽고DB 컬렉션 객체 생성
	c := session.DB("test").C("rooms")

	// rooms 컬렉션에 room 정보 저장
	if err := c.Insert(r); err != nil {
		// 오류 발생 시, 500 에러 반환
		Renderer.JSON(w, http.StatusInternalServerError, err)
		return
	}

	// 처리 결과 반환
	Renderer.JSON(w, http.StatusCreated, r)
}

// 채팅방 정보 조회
func RetrieveRooms(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	// 몽고DB 세션 생성
	session := MongoSession.Copy()
	// 몽고DB 세션을 닫는코드를 defer로 등록
	defer session.Close()

	var rooms []Room
	// 모든 room 정보 조회
	err := session.DB("test").C("rooms").Find(nil).All(&rooms)
	if err != nil {
		// 오류 발생 시 500 에러 반환
		Renderer.JSON(w, http.StatusInternalServerError, err)
		return
	}

	// room 조회 결과 반환
	Renderer.JSON(w, http.StatusOK, rooms)
}
