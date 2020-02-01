// LianDi - 链滴笔记，链接点滴
// Copyright (c) 2020-present, b3log.org
//
// Lute is licensed under the Mulan PSL v1.
// You can use this software according to the terms and conditions of the Mulan PSL v1.
// You may obtain a copy of Mulan PSL v1 at:
//     http://license.coscl.org.cn/MulanPSL
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v1 for more details.

package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/go-ps"
	"gopkg.in/olahol/melody.v1"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	InitLog()
	InitConf()
	InitMount()
	InitSearch()

	go ParentExited()
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	m := melody.New()
	m.Config.MaxMessageSize = 1024 * 1024 * 2
	r.GET("/ws", func(c *gin.Context) {
		if err := m.HandleRequest(c.Writer, c.Request); nil != err {
			Logger.Errorf("处理命令失败：%s", err)
		}
	})

	r.POST("/upload", Upload)

	m.HandleConnect(func(s *melody.Session) {
		SetPushChan(s)
		Logger.Debug("websocket connected")
	})

	m.HandleDisconnect(func(s *melody.Session) {
		Logger.Debugf("websocket disconnected")
	})

	m.HandleError(func(s *melody.Session, err error) {
		Logger.Debugf("websocket on error: %s", err)
	})

	m.HandleClose(func(s *melody.Session, i int, str string) error {
		Logger.Debugf("websocket on close: %v, %v", i, str)
		return nil
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		Logger.Debugf("request [%s]", msg)
		request := map[string]interface{}{}
		if err := json.Unmarshal(msg, &request); nil != err {
			result := NewResult()
			result.Code = -1
			result.Msg = "Bad Request"
			responseData, _ := json.Marshal(result)
			Push(responseData)
			return
		}

		cmdStr := request["cmd"].(string)
		cmdId := request["reqId"].(float64)
		param := request["param"].(map[string]interface{})
		cmd := NewCommand(cmdStr, cmdId, param)
		if nil == cmd {
			result := NewResult()
			result.Code = -1
			result.Msg = "查找命令 [" + cmdStr + "] 失败"
			Push(result.Bytes())
			return
		}
		Exec(cmd)
	})

	handleSignal()

	addr := "127.0.0.1:" + ServerPort
	Logger.Infof("内核进程 [v%s] 正在启动，监听端口 [%s]", Ver, "http://"+addr)
	if err := r.Run(addr); nil != err {
		Logger.Errorf("启动链滴笔记内核失败 [%s]", err)
	}
}

func handleSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		s := <-c
		Logger.Infof("收到系统信号 [%s]，退出内核进程", s)

		Close()
		os.Exit(0)
	}()
}

var ppid = os.Getppid()

func ParentExited() {
	for range time.Tick(2 * time.Second) {
		process, e := ps.FindProcess(ppid)
		if nil == process || nil != e {
			Logger.Info("UI 进程已经退出，现在退出内核进程")
			Close()
			os.Exit(0)
		}
	}
}