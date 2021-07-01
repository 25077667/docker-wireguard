package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var cpu_info_lock sync.RWMutex
var cpu_rate float64

var logger = log.New(os.Stdout, "[DEBUG]", log.Ltime)

func get_cpu() {
	for {
		me, _ := cpu.Percent(time.Second, true)
		rate := float64(0)
		for i := 0; i < len(me); i++ {
			rate += me[i] / float64(len(me))
		}

		cpu_info_lock.Lock()
		cpu_rate = rate
		cpu_info_lock.Unlock()
	}
}

func get_sys_info() []byte {
	type Percent struct {
		Mem float64 `json:Mem`
		CPU float64 `json:CPU`
	}

	res := &Percent{
		Mem: func() float64 {
			v, _ := mem.VirtualMemory()
			return v.UsedPercent
		}(),
		CPU: func() float64 {
			cpu_info_lock.RLock()
			me := cpu_rate
			cpu_info_lock.RUnlock()
			return me
		}(),
	}

	s, _ := json.Marshal(res)
	return s
}

func init() {
	logfile, _ := os.OpenFile("vpnLog", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	logger.SetOutput(logfile)
	go get_cpu()
}

func main() {
	app := fiber.New()
	app.Get("/stat", func(c *fiber.Ctx) error {
		return c.Send(get_sys_info())
	})
	app.Get("/peer:id?", func(c *fiber.Ctx) error {
		id := c.Params("id") // Suppose it is a string of int
		path := fmt.Sprintf("/config/peer%s/peer%s.conf", id, id)
		ctx, err := ioutil.ReadFile(path) // give Sendfile to check if the file is exist
		if err != nil {
			logger.Println(err)
		}
		return c.Send(ctx)
	})

	app.Listen(":8080")
}
