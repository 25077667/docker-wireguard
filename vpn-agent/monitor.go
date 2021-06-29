package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var cpu_info_lock sync.RWMutex

var cpu_rate float64

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

func main() {
	go get_cpu()
	app := fiber.New()
	app.Get("/stat", func(c *fiber.Ctx) error {
		return c.Send(get_sys_info())
	})
	app.Get("/peer:id?", func(c *fiber.Ctx) error {
		id := c.Params("id") // Suppose it is a string of int
		path := fmt.Sprintf("/config/peer%s/peer%s.conf", id, id)
		return c.SendFile(path) // give Sendfile to check if the file is exist
	})

	app.Listen(":8080")
}
