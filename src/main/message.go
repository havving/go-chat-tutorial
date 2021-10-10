package main

import (
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

const messageFetchSize = 10

type Message struct {
	ID        bson.ObjectId `bson:"_id" json:"id"`
	RoomId    bson.ObjectId `bson:"room_id" json:"room_id"`
	Content   string        `bson:"content" json:"content"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	User      *User         `bson:"user" json:"user"`
}

// 웹소켓으로 메시지 생성
func (m *Message) create() error {
	// 몽고DB 세션 생성
	session := mongoSession.Copy()
	// 몽고DB 세션을 닫는 코드를 defer로 등록
	defer session.Close()
	// 몽고DB ID 생성
	m.ID = bson.NewObjectId()
	// 메시지 생성 시간 기록
	m.CreatedAt = time.Now()
	// 메시지 정보 저장을 위한 몽고DB 컬렉션 객체 생성
	c := session.DB("test").C("messages")

	// messages 컬렉션에 메시지 정보 저장
	if err := c.Insert(m); err != nil {
		return err
	}

	return nil
}

// REST API 액션으로 사용할 함수
func retrieveMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 몽고DB 세션 생성
	session := mongoSession.Copy()
	// 몽고DB 세션을 닫는 코드를 defer로 등록
	defer session.Close()

	// 쿼리 매개변수로 전달된 limit 값 확인
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		// 정상적인 limit 값이 전달되지 않으면 limit를 messageFetchSize로 세팅
		limit = messageFetchSize
	}

	var messages []Message
	// _id 역순으로 정렬하여 limit 수만큼 message 조회
	err = session.DB("test").C("messages").
		Find(bson.M{"room_id": bson.ObjectIdHex(ps.ByName("id"))}).
		Sort("-_id").Limit(limit).All(&messages)

	if err != nil {
		// 오류 발생 시 500에러 반환
		renderer.JSON(w, http.StatusInternalServerError, err)
		return
	}

	// message 조회 결과 반환
	renderer.JSON(w, http.StatusOK, messages)
}
