package scanner

import (
	"fmt"
	"log"
	"net"
        "reflect"
	"net/netip"
	"time"
        "strconv"

	"github.com/mdlayher/arp"
)

func getLocalInterface(name string) (*net.Interface, *net.IPNet, error) {
    ifi, err := net.InterfaceByName(name)
    if err != nil {
        return nil, nil, err
    }

    addrs, err := ifi.Addrs()
    if err != nil {
        return nil, nil, err
    }
    log.Println(addrs)

    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
            log.Println("ipnet ", ipnet)
            return ifi, ipnet, nil
        }
    }
    return nil, nil, fmt.Errorf("no IPv4 address found on interface %s", name)
}

func arpScan(interfaceName string) (map[string]string, error) {
    ifi, ipnet, err := getLocalInterface(interfaceName)
    if err != nil {
        return nil, err
    }

    conn, err := arp.Dial(ifi)
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    log.Println(ipnet.IP)
    log.Println(ipnet.Mask)

    devices := make(map[string]string)

    ip := ipnet.IP.To4()

    convertToByteSlices(ipnet.IP.To4())
    test := [][]byte{
        {1, 1, 0, 0, 0, 0, 0, 0},
        {1, 0, 1, 0, 1, 0, 0, 0},
        {0, 0, 0, 0, 0, 0, 0, 0},
        {0, 0, 0, 0, 1, 0, 1, 1},
    }
    convertByteSlicesToIP(test)


    for i := 1; i <= 254; i++ {
        ipAddr := net.IPv4(ip[0], ip[1], ip[2], byte(i))
        if ipAddr.Equal(ip) {
            continue // skip self
        }

        target, ok := netip.AddrFromSlice(ipAddr)
        if !ok {
            continue
        }

        // Send ARP request
        if err := conn.Request(target); err != nil {
            continue
        }

        // Set a short timeout
        _ = conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))

        // Read ARP response
        pkt, _, err := conn.Read()
        if err == nil {
            devices[pkt.SenderHardwareAddr.String()] = pkt.SenderIP.String()
        }
    }

    return devices, nil
}
// 
func bitsToInt(bits []byte) int {
    var result int
    for _, bit := range bits {
        result = (result << 1) | int(bit)
    }
    return result
}

func convertByteSlicesToIP(test [][]byte) {
    log.Println(test)
    var ipComponents []string
    for _, subSlice := range test {
        s := bitsToInt(subSlice)
        ipComponents = append(ipComponents, strconv.Itoa(s))
    }
    log.Println(ipComponents)

    a := []byte{0, 1, 1, 0, 1, 0, 0, 0}
    b := []byte{1, 1, 1, 1, 0, 0, 0, 0}
    c := make([]byte, len(a))

    for i := 0; i < len(a); i++ {
        c[i] = a[i] ^ b[i]
    }

}

func convertMask(mask net.IPMask) int {
    ones, _ := mask.Size()
    return ones
}

func convertToByteSlices(ip net.IP) {
    log.Println(ip)
    log.Println(reflect.TypeOf(ip))
    var result [][]byte
    for _, b := range ip {
        var bits []byte
        for i := 7; i >= 0; i-- {
            bits = append(bits, (b>>i)&1)
        }
        result = append(result, bits)
    }
    log.Println(result)
}

func convertMaskToByteSlices(mask net.IPMask) {
}

func DoScan() {
    log.Printf("Do Scan")
    devices, err := arpScan("eno1") // replace with your interface name
    if err != nil {
        log.Fatal(err)
    }
    for mac, ip := range devices {
        log.Printf("IP: %s MAC: %s", ip, mac)
    }
}

