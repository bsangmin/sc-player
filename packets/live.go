package packets

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/sys/windows"
)

const winSIO_RCVALL = windows.IOC_IN | windows.IOC_VENDOR | 1

// Handle gopacket.PacketDataSource interface structure
type Handle struct {
	blockForever bool
	device       string
	deviceIndex  int
	mu           sync.Mutex
	socket       windows.Handle
	timeout      time.Duration
	snaplen      int32
}

// Close closes the underlying socket handle.
func (p *Handle) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	windows.Close(p.socket)
}

// ReadPacketData gopacket.PacketDataSource interface method
func (p *Handle) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	p.mu.Lock()
	data = make([]byte, 65536)
	n, _, err := windows.Recvfrom(p.socket, data, 0)

	// log.Println(n)

	if err != nil {
		log.Printf("Error:Recvfrom() - %v", err)
	}
	ci = gopacket.CaptureInfo{Timestamp: time.Now(), CaptureLength: len(data), Length: n, InterfaceIndex: p.deviceIndex}

	p.mu.Unlock()

	return
}

// LinkType returns pcap_datalink, as a layers.LinkType.
func (p *Handle) LinkType() layers.LinkType {
	return layers.LinkTypeIPv4
}

// OpenLive is caputer live packet
func OpenLive(snaplen int32, timeout time.Duration) (handle *Handle, ip4 net.IP, err error) {
	p := &Handle{}
	p.blockForever = timeout < 0
	p.timeout = timeout
	p.snaplen = snaplen

	var d windows.WSAData

	log.Println("Initialising Winsock...")
	err = windows.WSAStartup(uint32(0x202), &d)
	if err != nil {
		return nil, nil, fmt.Errorf("Error: WSAStartup - %v", err)
	}
	log.Println("Initialised")

	//Create a RAW Socket
	log.Println("Creating RAW Socket...")
	fd, err := windows.Socket(windows.AF_INET, windows.SOCK_RAW, windows.IPPROTO_IP)
	if err != nil {
		return nil, nil, fmt.Errorf("Error: socket - %v", err)
	}
	p.socket = fd
	log.Println("Created.")

	// Retrieve the local hostname
	hostname, err := os.Hostname()

	if err != nil {
		return nil, nil, fmt.Errorf("Error: Hostname() - %v", err)
	}
	log.Printf("\nHost name : %s \n", hostname)

	//Retrieve the available IPs of the local host
	log.Printf("Available Network Interfaces : \n\n")
	_, err = windows.GetHostByName(hostname)

	if err != nil {
		return nil, nil, fmt.Errorf("Error: GetHostByName() - %v", err)
	}

	myip4, iFcindex, err := externalIP()
	if err != nil {
		return nil, nil, fmt.Errorf("Error: externalIP() - %v", err)
	}
	p.deviceIndex = iFcindex

	la := new(windows.SockaddrInet4)
	la.Port = int(0)

	for i := 0; i < net.IPv4len; i++ {
		la.Addr[i] = myip4[i]
	}

	if err := windows.Bind(fd, la); err != nil {
		return nil, nil, fmt.Errorf("Error:Bind - %v", err)
	}

	inbuf := uint32(1)
	sizebuf := uint32(unsafe.Sizeof(inbuf))
	ret := uint32(0)

	err = windows.WSAIoctl(fd, winSIO_RCVALL, (*byte)(unsafe.Pointer(&inbuf)), sizebuf, nil, 0, &ret, nil, 0)

	if err != nil {
		return nil, nil, fmt.Errorf("Error:WSAIoctl() failed - %v", err)
	}

	ip4 = myip4

	return p, ip4, nil

}

func htons(n int) int {
	return int(int16(byte(n))<<8 | int16(byte(n>>8)))
}

func externalIP() (IPBYTE []byte, ifaceIndex int, err error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

	IPBYTE = make([]byte, 4)
	ifaceIndex = 0

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			IPBYTE = ip.To4()
			ifaceIndex = iface.Index
			if IPBYTE == nil {
				continue // not an ipv4 address
			}
			//~ err = nil

			log.Printf("Active Network Interfaces %v : %v ", iface.Index, ip.String())
			return IPBYTE, ifaceIndex, nil
		}
	}

	err = errors.New("are you connected to the network?")
	return
}
