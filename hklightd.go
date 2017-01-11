package main

import (
	"github.com/brutella/hc"
	"github.com/eager7/go_study/myProject/src/socketClient"
	"github.com/brutella/hc/accessory"
	"log"
	"encoding/json"
	"os"
	"runtime"
)
var sock_zigbee socketClient.SockClient
var device_lists []device_info

type device_info struct{
	device_name string
	device_id uint16
	device_online uint8
	device_mac uint64
}

func turnLightOn() {
	log.Println("Turn Light On")
	type jsoncmd struct {
		Cmd 	uint16 	`json:"command"`
		Seq 	int 	`json:"sequence"`
		Addr 	uint64 	`json:"device_address"`
		Group 	uint8 	`json:"group_id"`
		Mod 	uint8 	`json:"mode"`
	}
	for i := range device_lists{
		if device_lists[i].device_id == 0x0101{
			cmd_on,_ := json.Marshal(jsoncmd{0x0020, 0, device_lists[i].device_mac, 0, 1})
			sock_zigbee.SendMsgWithResp(string(cmd_on))
		}
	}
}

func turnLightOff() {
	log.Println("Turn Light Off")
	type jsoncmd struct {
		Cmd 	uint16 	`json:"command"`
		Seq 	int 	`json:"sequence"`
		Addr 	uint64 	`json:"device_address"`
		Group 	uint8 	`json:"group_id"`
		Mod 	uint8 	`json:"mode"`
	}
	for i := range device_lists{
		if device_lists[i].device_id == 0x0101{
			cmd_on,_ := json.Marshal(jsoncmd{0x0020, 0, device_lists[i].device_mac, 0, 0})
			sock_zigbee.SendMsgWithResp(string(cmd_on))
		}
	}
}

func setLightBrightness(level int){
	log.Println("setLightBrightness")
	type jsoncmd struct {
		Cmd 	uint16 	`json:"command"`
		Seq 	int 	`json:"sequence"`
		Addr 	uint64 	`json:"device_address"`
		Group 	uint8 	`json:"group_id"`
		Level 	int 	`json:"light_level"`
	}
	for i := range device_lists{
		if device_lists[i].device_id == 0x0101{
			cmd_on,_ := json.Marshal(jsoncmd{0x0021, 0, device_lists[i].device_mac, 0, level})
			sock_zigbee.SendMsgWithResp(string(cmd_on))
		}
	}
}

func GetDeviceLists(){
	err := sock_zigbee.Init("127.0.0.1", 6667)
	if err != nil{
		log.Fatal("Error:", err)
		os.Exit(1)
	}

	type searchCmd struct {
		Command int `json:"command"`
		Sequence int `json:"sequence"`
	}
	cmd, _ := json.Marshal(searchCmd{0x0011,0})
	ret, sta := sock_zigbee.SendMsgWithResp(string(cmd))
	if sta != 0{
		log.Fatal("recv msg error")
		os.Exit(1)
	}

	type description struct{
		Device_name string `json:"device_name"`
		Device_id uint16 `json:"device_id"`
		Device_online uint8 `json:"device_online"`
		Device_mac uint64 `json:"device_mac_address"`
	}
	type resp struct {
		Status uint8 `json:"status"`
		Sequence int `json:"sequence"`
		Desc []description `json:"description"`
	}
	var r resp;
	err = json.Unmarshal([]byte(ret), &r)
	log.Println(err)
	if _, _, line, _ := runtime.Caller(0); err != nil {
		log.Println(line, err)
	}

	if r.Status == 0{
		for _,dev := range r.Desc{
			device_lists = append(device_lists, device_info{dev.Device_name,
				dev.Device_id, dev.Device_online, dev.Device_mac})
		}
	}
	log.Println(device_lists)
}

func main() {
	GetDeviceLists()
	info := accessory.Info{
		Name:         "Personal Light Bulb",
		Manufacturer: "Matthias",
	}

	acc := accessory.NewLightbulb(info)

	acc.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			turnLightOn()
		} else {
			turnLightOff()
		}
	})

	acc.Lightbulb.Brightness.OnValueRemoteUpdate(func(brightness int) {
		setLightBrightness(brightness)
	})

	acc.Lightbulb.Hue.OnValueRemoteUpdate(func(hue float64) {
		log.Println("hue:", hue)
	})

	acc.Lightbulb.Saturation.OnValueRemoteUpdate(func(sat float64) {
		log.Println("sat:", sat)
	})

	t, err := hc.NewIPTransport(hc.Config{Pin: "32191123"}, acc.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

	t.Start()
}
