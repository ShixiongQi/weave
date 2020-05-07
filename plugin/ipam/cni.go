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

	logFileName := "/users/sqi009/weave_ipam_cmdAdd_info.log"
	logFile, _  := os.Create(logFileName)
	defer logFile.Close()
	debugLog := log.New(logFile,"[Info: cni.go]",log.Lmicroseconds)
	debugLog.Println("[weave-ipam] cmdAdd start")

	common.Log.Debugln("[ipam cni.go] CmdAdd")
	var conf types.NetConf
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}
	debugLog.Println("[weave-ipam] ipam allocate start")
	result, err := i.Allocate(args)
	if err != nil {
		return err
	}
	debugLog.Println("[weave-ipam] ipam allocate finish")
	debugLog.Println("[weave-ipam] cmdAdd finish")	
	return types.PrintResult(result, conf.CNIVersion)
}

func (i *Ipam) Allocate(args *skel.CmdArgs) (types.Result, error) {
	// extract the things we care about
	common.Log.Debugln("[ipam cni.go] Allocate")
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
		ipnet, err = i.weave.AllocateIP(containerID, false)
	} else {
		var subnet *net.IPNet
		subnet, err = types.ParseCIDR(conf.Subnet)
		if err != nil {
			return nil, fmt.Errorf("subnet given in config, but not parseable: %s", err)
		}
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
