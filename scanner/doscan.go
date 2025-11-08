package scanner

import (
	"fmt"
	"log"
	"net"
        "sync"
//        "reflect"
        "math"
        "strings"
        "context"
        "errors"
        "time"
	"net/netip"
        "strconv"

	"github.com/mdlayher/arp"

	"graphite/publisher/db"
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
//    log.Println(addrs)

    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
//            log.Println("ipnet ", ipnet)
            return ifi, ipnet, nil
        }
    }
    return nil, nil, fmt.Errorf("no IPv4 address found on interface %s", name)
}

func arpScan(interfaceName string, response chan db.Address) error {
    log.Println("arp scan running ...")
    ifi, ipnet, err := getLocalInterface(interfaceName)
    if err != nil {
        return err
    }

    binaryIP := convertIPToByteSlices(ipnet.IP.To4())

    // binaryMask := [][]byte{
    //     {1, 1, 1, 1, 1, 1, 1, 1},
    //     {1, 1, 1, 1, 1, 1, 1, 1},
    //     {1, 1, 1, 1, 1, 1, 1, 1},
    //     {0, 0, 0, 0, 0, 0, 0, 0},
    // }

    binaryMask := convertMaskToByteSlices(ipnet.Mask)

    firstIP := subnetFirstAddress(binaryIP , binaryMask)

    var wg sync.WaitGroup
    validHosts := countValidHosts(binaryMask) 
    currentIP := firstIP

    for j := 0; j <= validHosts; j++ {
        wg.Add(1)
        go func(ip [][]byte) {
            defer wg.Done()
            err := checkHostAlive(convertByteSlicesToIP(ip), ifi, response)
            if err != nil {
            }
        }(currentIP)
        currentIP = nextAddress(currentIP)
    }

    wg.Wait()
    close(response)
    return nil
}

func checkHostAlive(ip string, ifi *net.Interface, response chan db.Address) error {
    conn, err := arp.Dial(ifi)
    if err != nil {
        return err
    }
    defer conn.Close()

    ipAddr, err := netip.ParseAddr(ip)
    if err != nil {
        return err
    }

    hwAddr, err := resolveWithTimeout(conn, ipAddr, 5*time.Second)
    if err != nil {
        // no reply or error - host might be unreachable
        return err
    } else {
	msg := db.Address{Mac: hwAddr.String(), IP: ip}
	// log.Println("placing single address on response queue: ", msg)
        response <- msg
    }
    return nil
}

func resolveWithTimeout(conn *arp.Client, ip netip.Addr, timeout time.Duration) (net.HardwareAddr, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    resultCh := make(chan struct {
        hw  net.HardwareAddr
        err error
    }, 1)

    go func() {
        hw, err := conn.Resolve(ip)
        resultCh <- struct {
            hw  net.HardwareAddr
            err error
        }{hw, err}
    }()

    select {
    case res := <-resultCh:
        return res.hw, res.err
    case <-ctx.Done():
        return nil, errors.New("ARP resolution timed out")
    }
}

func convertIPToByteSlices(ip net.IP) [][]byte {
    var result [][]byte
    for _, b := range ip {
        var bits []byte
        for i := 7; i >= 0; i-- {
            bits = append(bits, (b>>i)&1)
        }
        result = append(result, bits)
    }
    return result
}

func bitsToInt(bits []byte) int {
    var result int
    for _, bit := range bits {
        result = (result << 1) | int(bit)
    }
    return result
}

func convertByteSlicesToIP(binaryIP [][]byte) string {
    var ipComponents []string
    for _, subSlice := range binaryIP {
        s := bitsToInt(subSlice)
        ipComponents = append(ipComponents, strconv.Itoa(s))
    }
    ip := strings.Join(ipComponents, ".")
    return ip
}

func convertMaskToByteSlices(mask net.IPMask) [][]byte {
    var result [][]byte
    for _, b := range mask {
        var bits []byte
        for i := 7; i >= 0; i-- {
            bits = append(bits, (b>>i)&1)
        }
        result = append(result, bits)
    }
    return result
}

// Operations
func subnetFirstAddress(ip [][]byte, mask [][]byte) [][]byte {
    var a []byte
    var b []byte
    var subnet [][]byte
    for i := 0; i < len(ip); i++ {
        a = ip[i] 
        b = mask[i]
        c := make([]byte, len(a))
        for k := 0; k < len(a); k++ {
            //c[k] = a[k] ^ b[k]
            c[k] = a[k] & b[k]
        }
        subnet = append(subnet, c)
    }
    return subnet
}

// Not currently used - could be removed but it's nice to have as a reference
func subnetLastAddress(ip [][]byte, mask [][]byte) string {
    var a []byte
    var b []byte
    var binaryLastIP [][]byte
    var count int = 0
    for i := 0; i < len(ip); i++ {
        a = ip[i] 
        b = mask[i]
        c := make([]byte, len(a))
        for k := 0; k < len(a); k++ {
            //c[k] = a[k] ^ b[k]
            if b[k] == 1 {
                count++
            }
            inverse := b[k] ^ 1
            c[k] = a[k] | inverse
        }
        binaryLastIP = append(binaryLastIP, c)
    }
    lastIP := convertByteSlicesToIP(binaryLastIP)
    return lastIP
}

func countValidHosts(mask [][]byte) int {
    hostBits := 0
    for _, byteSlice := range mask {
        for _, bit := range byteSlice {
            if bit == 0 {
                hostBits++
            }
        }
    }
    // Handle edge case: /32 subnet has 0 valid hosts
    if hostBits == 0 {
        return 0
    }
    // Total usable hosts = 2^hostBits - 2 (excluding network & broadcast)
    total := int(math.Pow(2, float64(hostBits))) - 2
    return total
}

func nextAddress(addr [][]byte) [][]byte {
    // Copy input to avoid modifying original
    newAddr := make([][]byte, len(addr))
    for i := range addr {
        newAddr[i] = append([]byte(nil), addr[i]...)
    }
    // Start from the least significant bit (last byte, last bit)
    for i := len(newAddr) - 1; i >= 0; i-- {
        for j := 7; j >= 0; j-- {
            if newAddr[i][j] == 0 {
                newAddr[i][j] = 1
                // All bits after this stay the same
                return newAddr
            }
            // Otherwise, carry over
            newAddr[i][j] = 0
        }
    }
    // If all bits were 1, we wrapped around; return all 0s again
    return newAddr
}

func DoScan(interfaceName string, response chan db.Address) {
    _ = arpScan(interfaceName, response)
    // TODO add error channel to return errors to parent 
}

