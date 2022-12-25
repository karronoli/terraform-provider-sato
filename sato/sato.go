package sato

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
	"unsafe"
)

type RawResponse struct {
	STX            byte
	HardwareAddr   [6]byte
	Delimiter1     byte
	IPAddress      [4]byte
	Delimiter2     byte
	SubnetMask     [4]byte
	Delimiter3     byte
	GatewayAddress [4]byte
	Delimiter4     byte
	Name           [32]byte
	Delimiter5     byte
	DHCP           byte
	RARP           byte
	ETX            byte
}

type Response struct {
	HardwareAddress net.HardwareAddr
	IPAddress       net.IP
	SubnetMask      net.IP
	GatewayAddress  net.IP
	Name            string
	DHCP            bool
	RARP            bool
}

const ASCII_SOH = 0x01
const ASCII_STX = 0x02
const ASCII_ETX = 0x03
const ASCII_ESC = 0x1b
const REQUEST_PORT = 19541

func assign_static_ip(hw_addr net.HardwareAddr, _ip net.IP, _subnet net.IP, _gateway net.IP) {
	mac := strings.Replace(hw_addr.String(), ":", "", -1)
	ip := _ip.To4()
	subnet := _subnet.To4()
	gateway := _gateway.To4()

	esc := string(rune(ASCII_ESC))
	WJ0 := esc + "WJ0"
	WI0 := esc + "WI0"
	W1 := esc + fmt.Sprintf("W1%03d%03d%03d%03d", ip[0], ip[1], ip[2], ip[3])
	W2 := esc + fmt.Sprintf("W2%03d%03d%03d%03d", subnet[0], subnet[1], subnet[2], subnet[3])
	W3 := esc + fmt.Sprintf("W3%03d%03d%03d%03d", gateway[0], gateway[1], gateway[2], gateway[3])
	operations := []string{
		WJ0,
		WI0,
		W1,
		W2,
		W3,
		mac,
	}

	payload := strings.Join(operations, ",")
	assign(payload)

}

func assign_dhcp(hw_addr net.HardwareAddr) {
	mac := strings.Replace(hw_addr.String(), ":", "", -1)

	esc := string(rune(ASCII_ESC))
	WJ1 := esc + "WJ1"
	WI1 := esc + "WI1"
	operations := []string{
		WJ1,
		WI1,
		mac,
	}

	payload := strings.Join(operations, ",")
	assign(payload)
}

func assign(payload string) {
	remote_endpoint := net.UDPAddr{IP: net.IPv4bcast, Port: REQUEST_PORT}
	client, err := net.DialUDP("udp", nil, &remote_endpoint)

	if err != nil {
		panic(err)
	}

	defer client.Close()

	// for PrtSetTool_LespritV.exe parameter
	sleep_interval := 15 * time.Millisecond
	request_count := 3

	for i := 0; i < request_count; i++ {
		_, err = client.Write([]byte(payload))

		if err != nil {
			panic(err)
		}

		time.Sleep(sleep_interval)
	}
}

func search() []Response {
	remote_endpoint := net.UDPAddr{IP: net.IPv4bcast, Port: REQUEST_PORT}
	client, err := net.DialUDP("udp", nil, &remote_endpoint)

	if err != nil {
		panic(err)
	}

	defer client.Close()
	_, err = client.Write([]byte{ASCII_SOH, 'L', 'A'})

	if err != nil {
		panic(err)
	}

	// Release source port, use for listen port
	client.Close()
	udp_addr, _ := net.ResolveUDPAddr("udp", client.LocalAddr().String())
	server_endpoint := net.UDPAddr{IP: net.IPv4zero, Port: udp_addr.Port}
	server, err := net.ListenUDP("udp", &server_endpoint)

	if err != nil {
		panic(err)
	}

	defer server.Close()

	start := time.Now()
	end := start.Add(time.Second * 1)
	server.SetReadDeadline(end.Add(time.Millisecond * 100))

	responses := []Response{}

	for {
		if time.Now().After(end) {
			//fmt.Println("read wait timeout")
			break
		}

		raw_response := RawResponse{}
		buffer := make([]byte, unsafe.Sizeof(raw_response))
		length, _, _ := server.ReadFrom(buffer)

		switch uintptr(length) {
		case unsafe.Sizeof(raw_response):
			reader := bytes.NewReader(buffer)
			binary.Read(reader, binary.LittleEndian, &raw_response)
		case 0:
			//fmt.Println("skip")
			continue
		default:
			panic("unknown response")
		}

		if raw_response.STX != ASCII_STX || raw_response.ETX != ASCII_ETX {
			panic("bad response")
		}

		response := Response{
			HardwareAddress: net.HardwareAddr(raw_response.HardwareAddr[:]),
			IPAddress:       net.IP(raw_response.IPAddress[:]).To4(),
			SubnetMask:      net.IP(raw_response.SubnetMask[:]).To4(),
			GatewayAddress:  net.IP(raw_response.GatewayAddress[:]).To4(),
			Name: strings.TrimRight(string(
				bytes.TrimRight([]byte(raw_response.Name[:]), "\x00"),
			), " "),
			DHCP: raw_response.DHCP == 1,
			RARP: raw_response.RARP == 1,
		}

		responses = append(responses, response)
	}

	return responses
}
