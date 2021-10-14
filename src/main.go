package main

import (
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"go-chat/modules"
)

const socketBufferSize = 1024

var (
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  socketBufferSize,
		WriteBufferSize: socketBufferSize,
	}
)

func init() {
	// 렌더러 생성
	modules.Renderer = render.New()

	s, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		panic(err)
	}

	modules.MongoSession = *s
}

func main() {
	// 라우터 생성
	router := httprouter.New()

	// 핸들러 정의
	router.GET("/", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// 렌더러를 사용하여 템플릿 렌더링
		modules.Renderer.HTML(w, http.StatusOK, "index", map[string]string{"title": "Simple Chat!"})
	})

	router.GET("/login", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// 로그인 페이지 렌더링
		modules.Renderer.HTML(w, http.StatusOK, "login", nil)
	})

	router.GET("/logout", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// 세션에서 사용자 정보 제거 후 로그인 페이지로 이동
		//sessions.GetSession(r).Delete(keyCurrentUser)
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	// 소셜 로그인 기능
	router.GET("/auth/:action/:provider", modules.LoginHandler)

	// 채팅방
	router.POST("/rooms", modules.CreateRoom)
	router.GET("/rooms", modules.RetrieveRooms)
	router.GET("/rooms/:id/messages", modules.RetrieveMessages)

	router.GET("/ws/:room_id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		socket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal("ServeHTTP: ", err)
			return
		}
		modules.NewClient(socket, ps.ByName("room_id"), modules.GetCurrentUser(r))
	})

	// negroni 미들웨어 생성
	n := negroni.Classic()
	store := cookiestore.New([]byte(sessionSecret))
	n.Use(sessions.Sessions(sessionKey, store))

	// 인증 핸들러 등록
	n.Use(modules.LoginRequired("/login", "/auth"))

	// negroni에 router를 핸들러로 등록
	n.UseHandler(router)

	// 웹 서버 실행
	n.Run(":3000")
}

const (
	// application에서 사용할 세션의 키 정보
	sessionKey    = "simple_chat_session"
	sessionSecret = "simple_chat_session_secret"
)
