package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rainbowism/gin-pongo2"
	"net"
	"net/http"
	"os"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func vncStreamHandler(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}

	address := fmt.Sprintf("%s:5901", c.Param("ip"))

	fmt.Println("proxying ", address)

	// proxy
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		fmt.Println("cannot resolve address")
		return
	}

	tcpConn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println("cannot connect to address")
		return
	}
	defer tcpConn.Close()

	go func() {
		for {
			data := make([]byte, 100)

			n, err := tcpConn.Read(data)
			if err != nil {
				fmt.Println("error while reading data from tcp", err)
				return
			}

			conn.WriteMessage(websocket.BinaryMessage, data[0:n])
		}
	}()

	for {
		t, msg, err := conn.ReadMessage()
		if t != websocket.BinaryMessage {
			fmt.Println("received wrong websocket message type")
			return
		}

		if err != nil {
			fmt.Println("error while reading data from WS", err)
			return
		}

		tcpConn.Write(msg)
	}
}

func main() {
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)

	switch os.Getenv("MODE") {
	case "RELEASE":
		gin.SetMode(gin.ReleaseMode)

	default:
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	if gin.IsDebugging() {
		r.HTMLRender = render.NewDebug("templates")
	} else {
		r.HTMLRender = render.NewProduction("templates")
	}

	r.Static("/static", "./static")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/new", func(c *gin.Context) {
		container, err := client.CreateContainer(docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: "caas",
			},
		})

		fmt.Println(err)

		err = client.StartContainer(container.ID, nil)

		fmt.Println(err, container)

		container, err = client.InspectContainer(container.ID)

		ip := container.NetworkSettings.IPAddress

		c.HTML(http.StatusOK, "new.html", render.Context{
			"ip": ip,
		})
	})

	r.GET("/ws/:ip", vncStreamHandler)

	r.Run() // listen and server on 0.0.0.0:8080
}
