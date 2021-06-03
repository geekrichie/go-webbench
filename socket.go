package main

import (
	"net"
	"strconv"
)

func Socket(addr string , port int) (net.Conn, error){
	strport := strconv.Itoa(port)
	c, err := net.Dial("tcp", addr+ ":"+strport)
	if err != nil {
		return nil, err
	}
	return c, nil
}
