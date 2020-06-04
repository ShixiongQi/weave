package ipamplugin

import (
	"encoding/json"
	"fmt"
	"net"

	"log"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/weaveworks/weave/common"
)

func (i *Ipam) CmdAdd(args *skel.CmdArgs) error {
	logFileName := "/users/sqi009/weave-start-time.log"
	logFile, _  := os.OpenFile(logFileName,os.O_RDWR|os.O_APPEND|os.O_CREATE,0644)
	// logFile, _  := os.Create(logFileName)
	defer logFile.Close()
	debugLog := log.New(logFile,"[Info: weave-ipam.go]",log.Lmicroseconds)
	debugLog.Println("[weave-ipam] cmdAdd start")

	common.Log.Debugln("[ipam cni.go] CmdAdd")
	var conf types.NetConf
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}
	debugLog.Println("[weave-ipam] i.Allocate(args) start")
	result, err := i.Allocate(args)
	if err != nil {
		return err
	}
	debugLog.Println("[weave-ipam] i.Allocate(args) finish")
	debugLog.Println("[weave-ipam] cmdAdd finish")	
	return types.PrintResult(result, conf.CNIVersion)
}

func (i *Ipam) Allocate(args *skel.CmdArgs) (types.Result, error) {
	logFileName := "/users/sqi009/weave-start-time.log"
	logFile, _  := os.OpenFile(logFileName,os.O_RDWR|os.O_APPEND|os.O_CREATE,0644)
	// logFile, _  := os.Create(logFileName)
	defer logFile.Close()
	debugLog := log.New(logFile,"[Info: weave-ipam.go]",log.Lmicroseconds)
	debugLog.Println("[weave-ipam] inside Allocate")

	// extract the things we care about
	common.Log.Debugln("[ipam cni.go] Allocate")
	debugLog.Println("[weave-ipam] loadIPAMConf(args.StdinData) start")
	conf, err := loadIPAMConf(args.StdinData)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &ipamConf{}
	}
	containerID := args.ContainerID
	if containerID == "" {
		return nil, fmt.Errorf("Weave CNI Allocate: blank container name")
	}
	var ipnet *net.IPNet

	if conf.Subnet == "" {
		debugLog.Println("[weave-ipam] AllocateIP start")
		ipnet, err = i.weave.AllocateIP(containerID, false)
	} else {
		var subnet *net.IPNet
		subnet, err = types.ParseCIDR(conf.Subnet)
		if err != nil {
			return nil, fmt.Errorf("subnet given in config, but not parseable: %s", err)
		}
		debugLog.Println("[weave-ipam] AllocateIPInSubnet start")
		ipnet, err = i.weave.AllocateIPInSubnet(containerID, subnet, false)
	}

	if err != nil {
		return nil, err
	}
	result := &current.Result{
		IPs: []*current.IPConfig{{
			Version: "4",
			Address: *ipnet,
			Gateway: conf.Gateway,
		}},
		Routes: conf.Routes,
	}
	debugLog.Println("[weave-ipam] leave Allocate")
	return result, nil
}

func (i *Ipam) CmdDel(args *skel.CmdArgs) error {
	return i.Release(args)
}

func (i *Ipam) Release(args *skel.CmdArgs) error {
	return i.weave.ReleaseIPsFor(args.ContainerID)
}

type ipamConf struct {
	Subnet  string         `json:"subnet,omitempty"`
	Gateway net.IP         `json:"gateway,omitempty"`
	Routes  []*types.Route `json:"routes"`
}

type netConf struct {
	IPAM *ipamConf `json:"ipam"`
}

func loadIPAMConf(stdinData []byte) (*ipamConf, error) {
	var conf netConf
	return conf.IPAM, json.Unmarshal(stdinData, &conf)
}
