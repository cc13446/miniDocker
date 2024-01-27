package network

import (
	"testing"
)

func TestBridgeInit(t *testing.T) {
	d := BridgeNetworkDriver{}
	bridge, err := d.Create("192.168.0.1/24", "test_bridge")
	if err != nil {
		t.Errorf("Create err: %v", err)
	}
	if err := d.Delete(*bridge); err != nil {
		t.Errorf("Delete err: %v", err)
	}
}

func TestBridgeConnect(t *testing.T) {
	ep := Endpoint{
		ID: "test_container",
	}

	n := Network{
		Name: "test_bridge",
	}

	d := BridgeNetworkDriver{}
	err := d.Connect(&n, &ep)
	t.Logf("err: %v", err)
}

//func TestNetworkConnect(t *testing.T) {
//
//	cInfo := &container.ContainerInfo{
//		Id:  "testcontainer",
//		Pid: "15438",
//	}
//
//	d := BridgeNetworkDriver{}
//	n, err := d.Create("192.168.0.1/24", "testbridge")
//	t.Logf("err: %v", n)
//
//	Init()
//
//	networks[n.Name] = n
//	err = Connect(n.Name, cInfo)
//	t.Logf("err: %v", err)
//}
//
//func TestLoad(t *testing.T) {
//	n := Network{
//		Name: "testbridge",
//	}
//
//	n.load("/var/run/mydocker/network/network/testbridge")
//
//	t.Logf("network: %v", n)
//}
