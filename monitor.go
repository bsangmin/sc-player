package main

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	live "github.com/bsangmin/sc-player/packets"
	sc "github.com/bsangmin/sc-player/structures"
	"time"
	// "io/ioutil"
)

var players = make([]*sc.Player, 200)
var mapInfo = sc.MapInfo{[]byte("대기중..."), 0, 0}
var fixInfo = false

func getPackets() {
	var (
		buf     int32         = 65536
		timeout time.Duration = 30 * time.Second
	)

	var (
		// eth layers.Ethernet
		ip4 layers.IPv4
		tcp layers.TCP
		udp layers.UDP
		pay gopacket.Payload
	)

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &ip4, &tcp, &udp, &pay)
	decoded := []gopacket.LayerType{}

	hnd, myIP, err := live.OpenLive(buf, timeout)

	if err != nil {
		log.Fatal(err)
	}
	defer hnd.Close()

	packetSource := gopacket.NewPacketSource(hnd, hnd.LinkType())

	for packetData := range packetSource.Packets() {
		err := parser.DecodeLayers(packetData.Data(), &decoded)

		if packetData == nil || err != nil {
			continue
		}

		for _, layerType := range decoded {
			switch layerType {
			case layers.LayerTypeUDP:
				scPacketFilter(&ip4, &udp, &pay, myIP)
			}
		}
	}
}

type bzSvrInfo struct {
	ip   net.IP
	mask net.IPMask
}

var bzSvrs = map[int]bzSvrInfo{
	0: {[]byte{59, 153, 40, 0}, net.IPv4Mask(255, 255, 252, 0)},
	1: {[]byte{37, 244, 0, 0}, net.IPv4Mask(255, 255, 192, 0)},
	2: {[]byte{158, 115, 192, 0}, net.IPv4Mask(255, 255, 224, 0)},
}

var hostIP net.IP

func scPacketFilter(ip4 *layers.IPv4, udp *layers.UDP, payload *gopacket.Payload, myip net.IP) {

	if int(udp.DstPort) == 6112 || int(udp.SrcPort) == 6112 {
		scp := &sc.SCprotocol{}
		if err := scp.DecodeStructFromBytes(payload.Payload()); err != nil {
			// log.Println(err)
			return
		}

		if scp.Sign == 0x0801 {
			if scp.IsUserInfo() {
				player := &sc.Player{}
				player.DecodeStructFromBytes(scp.Payload)

				if player.Num == 0 && myip.Equal(ip4.SrcIP) {
					return
				}

				if player.Num == 0 && hostIP != nil && !player.Me {
					copy(player.IP, hostIP)
					player.Flag = true
					hostIP = nil
				}

				if players[player.Num] == nil {
					players[player.Num] = player
				} else {
					copy(players[player.Num].IP, player.IP)
				}

				log.Println("Player Info", player.Num, player.IP.String())

			} else if scp.IsJoin() && ip4.SrcIP.Equal(myip) { // 입장
				reset()
				log.Println("Enter room")

			} else if scp.IsMapInfo() && !fixInfo {
				now := time.Now().Unix()
				if now-mapInfo.Timestamp > 1 {
					mapInfo.Name = scp.Payload[22:]
					mapInfo.Status = 1
					mapInfo.Timestamp = now

					log.Println("Map Info", mapInfo.StringName(), mapInfo.StringStatus())
				}

			} else if scp.IsBattleCode() {
				pNum := uint8(scp.Payload[0])
				batcode := scp.Payload[8:108]
				name := scp.Payload[208:]

				if players[pNum] == nil {
					player := &sc.Player{}
					player.Num = pNum
					player.Name = name
					player.Flag = true

					if myip.Equal(ip4.SrcIP) {
						player.Me = true
					}

					if bzSvr(ip4.SrcIP) || player.Me {
						player.IP = []byte{0, 0, 0, 0}
					} else {
						player.IP = ip4.SrcIP
					}

					players[pNum] = player
				}

				players[pNum].Batcode = batcode

				log.Println("Battle code Info", pNum, players[pNum].Batcode.ToString(), players[pNum].IP.String())

			} else if scp.IsOutRoom() {
				pNum := uint8(scp.Payload[0])

				if fixInfo && players[pNum] != nil { // 게임중
					if players[pNum].Me { // 내가 나갈때
						mapInfo.Status = 3
						log.Println("Im out", mapInfo.StringName(), mapInfo.StringStatus())

					} else if !players[pNum].Out {
						players[pNum].Out = true
						log.Println("Out room", pNum)
					}

					if countPlayers() <= 1 {
						mapInfo.Status = 3
						log.Println("Nobody", mapInfo.StringName(), mapInfo.StringStatus())
					}

				} else if players[pNum] != nil { // 대기실
					if pNum == 0 || players[pNum].Me {
						mapInfo.Name = []byte("대기중...")
						mapInfo.Status = 0
						reset()

						log.Println("Out room me or host", pNum)

					} else {
						players[pNum] = nil
						log.Println("Out room", pNum)
					}

				}

			} else if scp.IsStart() {
				mapInfo.Status = 2
				fixInfo = true

				log.Println("Start", mapInfo.StringName(), mapInfo.StringStatus())
			}

		} else if scp.Sign == 0x0802 {
			mapInfo.Name = []byte("Host 입니다")
			mapInfo.Status = 1
			reset()

			log.Println("I am a host")
		}
	}

	if int(udp.SrcPort) == 6112 && int(udp.DstPort) == 6112 && myip.Equal(ip4.SrcIP) {
		holePunching := true
		for _, p := range players {
			if p != nil && p.IP.Equal(ip4.DstIP) {
				holePunching = false
				break
			}
		}

		if holePunching {
			if players[0] != nil {
				players[0].IP = ip4.DstIP
				players[0].Flag = true
			} else {
				hostIP = ip4.DstIP
			}
		}
	}
}

func countPlayers() uint8 {
	var count uint8 = 0

	for i := 0; i < 10; i++ {
		if players[i] != nil && !players[i].Out {
			count++
		}
	}

	for i := 120; i < 140; i++ {
		if players[i] != nil && !players[i].Out {
			count++
		}
	}

	return count

}

func reset() {
	fixInfo = false

	for i := 0; i < 10; i++ {
		players[i] = nil
	}

	for i := 120; i < 140; i++ {
		players[i] = nil
	}

	log.Println("reset")
}

func bzSvr(ip net.IP) bool {
	for _, info := range bzSvrs {
		if info.ip.Equal(ip.Mask(info.mask)) {
			return true
		}
	}
	return false
}
