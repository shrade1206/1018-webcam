package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	fmt.Println("Go WebSocket")

	setRoutes()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homepage(w http.ResponseWriter, r *http.Request) {
	d, err := os.ReadFile("index.html")
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", http.DetectContentType(d))
	w.Write(d)
}

func reader(conn *websocket.Conn) {
	var a []byte

	for {
		messageType, p, err := conn.ReadMessage()

		if string(p) == "run" {
			webcam, err := gocv.VideoCaptureDevice(0)
			if err != nil {
				log.Println(err)
			}

			time.Sleep(time.Second)

			img := gocv.NewMat()

			webcam.Read(&img)

			defer webcam.Close()

			buf, err := gocv.IMEncode(".jpg", img)
			defer buf.Close()

			a = buf.GetBytes()

			// d, _ := os.ReadFile(a)

			data := base64.StdEncoding.EncodeToString(a)

			// fmt.Println(data)

			if err := conn.WriteMessage(messageType, []byte(data)); err != nil {
				log.Println(err)
				return
			}

		}
		if string(p) == "save" {

			os.WriteFile("demo.jpg", a, os.ModePerm)

		}

		if err != nil {
			log.Println(err)
			return
		}
		log.Println("使用者訊息: " + string(p))
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("使用者已連線")

	reader(ws)
}

func setRoutes() {
	http.Handle("/", http.FileServer(http.Dir("./public")))

	http.HandleFunc("/ws", wsEndpoint)
}
