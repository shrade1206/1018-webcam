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

// func homepage(w http.ResponseWriter, r *http.Request) {
// 	d, err := os.ReadFile("index.html")
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	w.Header().Set("Content-Type", http.DetectContentType(d))
// 	w.Write(d)
// }
//設定接收與回傳
func reader(conn *websocket.Conn) {
	var newImg []byte

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if string(p) == "run" {
			func() {
				//設定視訊鏡頭，0 = 預設鏡頭
				webcam, err := gocv.VideoCaptureDevice(0)
				if err != nil {
					log.Println(err)
				}

				time.Sleep(time.Second)
				img := gocv.NewMat()
				defer img.Close()

				webcam.Read(&img)
				defer webcam.Close()
				//設定副檔名、來源
				buf, err := gocv.IMEncode(".jpg", img)
				defer buf.Close() //nolint
				//設定變數取得暫存檔案
				newImg = buf.GetBytes()

				// d, _ := os.ReadFile(a)
				//轉換成base64的字串型別
				data := base64.StdEncoding.EncodeToString(newImg)
				//把轉換好的字串傳送到前端，前端接收在轉換回圖片
				if err := conn.WriteMessage(messageType, []byte(data)); err != nil {
					log.Println(err)
					return
				}
			}() //func裡的func記得加 ()

		}
		if string(p) == "save" {
			//用來生成新文件使用(檔名、來源、)
			os.WriteFile("demo.jpg", newImg, os.ModePerm)
		}

		log.Println("使用者訊息: " + string(p))
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	// 透過http請求程序調用upgrader.Upgrade，來獲取*Conn (代表WebSocket連接)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("使用者已連線")

	reader(ws)
}

//設定PATH
func setRoutes() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	//進入/ws之後執行func wsEndpoint()
	http.HandleFunc("/ws", wsEndpoint)
}
