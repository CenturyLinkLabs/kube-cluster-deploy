package utils

import "net"

func WaitForSSH(publicIP string) error {
    for {
        conn, e := net.Dial("tcp", publicIP+":22")
        if e != nil {
            return e
        }
        defer conn.Close()
        if _, e = conn.Read(make([]byte, 1)); e != nil {
            continue
        }
        break
    }
    return nil
}

