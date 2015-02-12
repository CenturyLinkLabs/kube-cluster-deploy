package utils

import (
    "net"
    "encoding/json"
    "io/ioutil"
    "os"
    "bufio"
    "io"
    "strings"
    )

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
    c, e := ioutil.ReadFile("./config.json")
    if e != nil {
        return nil, e
    }
    if e = json.Unmarshal(c, &m); e != nil {
        return nil, e
    }
    return m, nil
}

func LoadStdinToEnvAndKeys() error {
    rd := bufio.NewReader(os.Stdin)
    for {
        ln := ""
        ln, e := rd.ReadString('\n')
        if e == io.EOF {
            break
        } else if e != nil {
            return e
        } else if strings.Contains(ln, "=") {
            kv := strings.SplitN(ln, "=", 2)
            SetKey(kv[0], kv[1])
            os.Setenv(kv[0], kv[1])
        }
    }
    return nil
}

