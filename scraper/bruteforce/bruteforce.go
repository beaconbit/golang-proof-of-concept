package bruteforce

func bruteForce(deviceMAC string, msgCh chan<- DeviceMsg) {
    replyCh := make(chan any)
    msgCh <- DeviceMsg{
        Cmd:     "get",
        Device:  Device{MAC: deviceMAC},
        ReplyCh: replyCh,
    }
    result := <-replyCh
    device := result.(Device)
    // proceed with auth/scraper logic
}

