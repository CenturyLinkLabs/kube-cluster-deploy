package utils

import (
    "net"
    "encoding/json"
    "io/ioutil")

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

func LoadJsonConfig() (map[string]string, error) {
    var m map[string]string
    c, e :=  ioutil.ReadFile("./config.json")
    if e != nil {
        return nil, e
    }
    if e = json.Unmarshal(c, &m); e != nil {
        return nil, e
    }
    return m, nil
}

