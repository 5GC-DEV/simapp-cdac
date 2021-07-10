// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/free5gc/logger_util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"time"
	"strconv"
)

type Config struct {
	Info          *Info               `yaml:"info"`
	Configuration *Configuration      `yaml:"configuration"`
	Logger        *logger_util.Logger `yaml:"logger"`
}

type Info struct {
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type Configuration struct {
	ConfigSlice  bool            `yaml:"provision-network-slice,omitempty"`
	DevGroup     []*DevGroup     `yaml:"device-groups,omitempty"`
	NetworkSlice []*NetworkSlice `yaml:"network-slices,omitempty"`
	Subscriber []*Subscriber `yaml:"subscribers,omitempty"`
	SubProvisionEndpt *SubProvisionEndpt `yaml:"sub-provision-endpt,omitempty"`
}

type DevGroup struct {
	Name         string    `yaml:"name,omitempty"`
	SiteInfo     string    `yaml:"site-info,omitempty" json:"site-info,omitempty"`
	Imsis        []string  `yaml:"imsis,omitempty" json:"imsis,omitempty"`
	IpDomainName string    `yaml:"ip-domain-name,omitempty" json:"ip-domain-name,omitempty"`
	IpDomain     *IpDomain `yaml:"ip-domain-expanded,omitempty" json:"ip-domain-expanded,omitempty"`
}

type IpDomain struct {
	Dnn        string `yaml:"dnn,omitempty" json:"dnn,omitempty"`
	DnsPrimary string `yaml:"dns-primary,omitempty" json:"dns-primary,omitempty"`
	Mtu        int    `yaml:"mtu,omitempty" json:"mtu,omitempty"`
	UePool     string `yaml:"ue-ip-pool,omitempty" json:"ue-ip-pool,omitempty"`
}

type Subscriber struct {
	UeId            string
	UeIdStart       string   `yaml:"ueId-start,omitempty" json:"ueId-start,omitempty"`
	UeIdEnd         string   `yaml:"ueId-end,omitempty" json:"ueId-end,omitempty"`
	PlmnId          string   `yaml:"plmnId,omitempty" json:"plmnId,omitempty"`
	OPc             string   `yaml:"opc,omitempty" json:"opc,omitempty"`
	Key             string   `yaml:"key,omitempty" json:"key,omitempty"`
	SequenceNumber  string   `yaml:"sequenceNumber,omitempty" json:"sequenceNumber,omitempty"`
}

type SubProvisionEndpt struct {
	Addr  string `yaml:"addr,omitempty" json:"addr,omitempty"`
	Port  string `yaml:"port,omitempty" json:"port,omitempty"`
}

type NetworkSlice struct {
	Name       string     `yaml:"name,omitempty" json:"name,omitempty"`
	SliceId    *SliceId   `yaml:"slice-id,omitempty" json:"slice-id,omitempty"`
	Qos        *QosInfo   `yaml:"qos,omitempty" json:"qos,omitempty"`
	DevGroups  []string   `yaml:"site-device-group,omitempty" json:"site-device-group,omitempty"`
	SiteInfo   *SiteInfo  `yaml:"site-info,omitempty" json:"site-info,omitempty"`
	DenyApps   []string   `yaml:"deny-applications,omitempty" json:"deny-applications,omitempty"`
	PermitApps []string   `yaml:"permit-applications,omitempty" json:"permit-applications,omitempty"`
	AppInfo    []*AppInfo `yaml:"applications-information,omitempty" json:"applications-information,omitempty"`
}

type SliceId struct {
	Sst string `yaml:"sst,omitempty" json:"sst,omitempty"`
	Sd  string `yaml:"sd,omitempty" json:"sd,omitempty"`
}

type QosInfo struct {
	Uplink       int    `yaml:"uplink,omitempty" json:"uplink,omitempty"`
	Downlink     int    `yaml:"downlink,omitempty" json:"downlink,omitempty"`
	TrafficClass string `yaml:"traffic-class,omitempty" json:"traffic-class,omitempty"`
}

type SiteInfo struct {
	SiteName string `yaml:"site-name,omitempty" json:"site-name,omitempty"`
	Gnb      []*Gnb `yaml:"gNodeBs,omitempty" json:"gNodeBs,omitempty"`
	Plmn     *Plmn  `yaml:"plmn,omitempty"   json:"plmn,omitempty"`
	Upf      *Upf   `yaml:"upf,omitempty" json:"upf,omitempty"`
}

type Gnb struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Tac  int    `yaml:"tac,omitempty" json:"tac,omitempty"`
}

type Plmn struct {
	Mcc string `yaml:"mcc,omitempty" json:"mcc,omitempty"`
	Mnc string `yaml:"mnc,omitempty" json:"mnc,omitempty"`
}

type Upf struct {
	UpfName string `yaml:"upf-name,omitempty" json:"upf-name,omitempty"`
	UpfPort int    `yaml:"upf-port,omitempty" json:"upf-port,omitempty"`
}

type AppInfo struct {
	AppName   string `yaml:"app-name,omitempty" json:"app-name,omitempty"`
	EndPort   int    `yaml:"end-port,omitempty" json:"end-port,omitempty"`
	EndPoint  string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Protocol  int    `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	StartPort int    `yaml:"start-port,omitempty" json:"start-port,omitempty"`
}

const (
	device_group = iota
	network_slice
	subscriber
)

type configMessage struct {
	msgPtr  *bytes.Buffer
	msgType int
    name    string
}

var SimappConfig Config

func InitConfigFactory(f string, configMsgChan chan configMessage, subProvisionEndpt *SubProvisionEndpt) error {
	fmt.Println("Function called ", f)
	if content, err := ioutil.ReadFile(f); err != nil {
		fmt.Println("Readfile failed called ", err)
		return err
	} else {
		SimappConfig = Config{}

		if yamlErr := yaml.Unmarshal(content, &SimappConfig); yamlErr != nil {
			fmt.Println("yaml parsing failed ", yamlErr)
			return yamlErr
		}
	}
	if SimappConfig.Configuration == nil {
		fmt.Println("Configuration Parsing Failed ", SimappConfig.Configuration)
		return nil
	}

	fmt.Println("Number of subscriber ranges", len(SimappConfig.Configuration.Subscriber))
	for o := 0; o < len(SimappConfig.Configuration.Subscriber); o++ {
		subscribers := SimappConfig.Configuration.Subscriber[o]
		fmt.Println("Subscribers:")
		fmt.Println("    UeIdStart", subscribers.UeIdStart)
		fmt.Println("    UeIdEnd", subscribers.UeIdEnd)
		fmt.Println("    PlmnId", subscribers.PlmnId)
		fmt.Println("    OPc", subscribers.OPc)
		fmt.Println("    Key", subscribers.Key)
		fmt.Println("    SequenceNumber", subscribers.SequenceNumber)

		start, err := strconv.ParseUint(subscribers.UeIdStart, 0, 64)
		if err != nil {
				fmt.Println("error in ParseUint with UeIdStart", err)
				continue
			}
		end, err := strconv.ParseUint(subscribers.UeIdEnd, 0, 64)
		if err != nil {
				fmt.Println("error in ParseUint with UeIdEnd", err)
				continue
			}
		for i := start; i <= end; i++ {
			subscribers.UeId = strconv.FormatUint(i, 10)
			fmt.Println("    UeId", subscribers.UeId)
			if err != nil {
				fmt.Println("error in FormatUint with UeId", err)
				continue
			}
			subscribers.UeIdStart = ""
			subscribers.UeIdEnd = ""
			b, err := json.Marshal(subscribers)
			if err != nil {
				fmt.Println("error in marshal with subscribers", err)
				continue
			}
			reqMsgBody := bytes.NewBuffer(b)
			var msg configMessage
			msg.msgPtr = reqMsgBody
			msg.msgType = subscriber
			msg.name    = subscribers.UeId
			configMsgChan <- msg
		}
	}

	fmt.Println("Subscriber Provision Endpoint:")
	fmt.Println("Address ", SimappConfig.Configuration.SubProvisionEndpt.Addr)
	fmt.Println("Port ", SimappConfig.Configuration.SubProvisionEndpt.Port)
	subProvisionEndpt.Addr = SimappConfig.Configuration.SubProvisionEndpt.Addr
	subProvisionEndpt.Port = SimappConfig.Configuration.SubProvisionEndpt.Port

	fmt.Println("Number of device Groups ", len(SimappConfig.Configuration.DevGroup))
	for g := 0; g < len(SimappConfig.Configuration.DevGroup); g++ {
		group := SimappConfig.Configuration.DevGroup[g]
		fmt.Println("Group Name ", group.Name)
		fmt.Println("  Site Name ", group.SiteInfo)
		fmt.Println("  Imsis ", group.Imsis)
		for im := 0; im < len(group.Imsis); im++ {
			fmt.Println("  IMSI ", group.Imsis[im])
		}
		fmt.Println("  IpDomainName ", group.IpDomainName)
		ipDomain := group.IpDomain
		if group.IpDomain != nil {
			fmt.Println("  IpDomain Dnn ", ipDomain.Dnn)
			fmt.Println("  IpDomain Dns Primary ", ipDomain.DnsPrimary)
			fmt.Println("  IpDomain Mtu ", ipDomain.Mtu)
			fmt.Println("  IpDomain UePool ", ipDomain.UePool)
		}
		b, err := json.Marshal(group)
		if err != nil {
			fmt.Println("error in marshal ", err)
			continue
		}
		reqMsgBody := bytes.NewBuffer(b)
		if SimappConfig.Configuration.ConfigSlice == false {
			fmt.Println("Don't configure network slice ")
			continue
		}
		var msg configMessage
		msg.msgPtr = reqMsgBody
		msg.msgType = device_group
		msg.name    = group.Name
		configMsgChan <- msg
	}

	fmt.Println("Number of network Slices ", len(SimappConfig.Configuration.NetworkSlice))
	for s := 0; s < len(SimappConfig.Configuration.NetworkSlice); s++ {
		slice := SimappConfig.Configuration.NetworkSlice[s]
		fmt.Println("  Slice Name : ", slice.Name)
		fmt.Printf("  Slice sst %v, sd %v", slice.SliceId.Sst, slice.SliceId.Sd)
		fmt.Println("  QoS information ", slice.Qos)
		fmt.Println("  Slice site info ", slice.SiteInfo)
		site := slice.SiteInfo
		fmt.Println("  Slice site name ", site.SiteName)
		fmt.Println("  Slice gNB ", len(site.Gnb))
		for e := 0; e < len(site.Gnb); e++ {
			fmt.Printf("  Slice gNB[%v] = %v  \n", e, site.Gnb[e])
		}
		fmt.Println("  Slice Plmn ", site.Plmn)
		fmt.Println("  Slice Upf ", site.Upf)

		fmt.Println("  Slice Device Groups ", slice.DevGroups)
		for im := 0; im < len(slice.DevGroups); im++ {
			fmt.Println("  Attached Device Groups  ", slice.DevGroups[im])
		}

		fmt.Println("  Permit Apps ", slice.PermitApps)
		for im := 0; im < len(slice.PermitApps); im++ {
			fmt.Println("  Permit Apps  ", slice.PermitApps[im])
		}

		fmt.Println("  Deny Apps ", slice.DenyApps)
		for im := 0; im < len(slice.DenyApps); im++ {
			fmt.Println("  Deny Apps  ", slice.DenyApps[im])
		}
		fmt.Println("  Application information ", slice.AppInfo)
		for im := 0; im < len(slice.AppInfo); im++ {
			fmt.Println("    Application Information ", slice.AppInfo[im])
		}

		b, err := json.Marshal(slice)
		if err != nil {
			fmt.Println("error in marshal ", err)
			continue
		}
		reqMsgBody := bytes.NewBuffer(b)

		if SimappConfig.Configuration.ConfigSlice == false {
			fmt.Println("Don't configure network slice ")
			continue
		}
		var msg configMessage
		msg.msgPtr = reqMsgBody
		msg.msgType = network_slice
        msg.name    = slice.Name
		configMsgChan <- msg
	}
	return nil
}

func main() {
	configMsgChan := make(chan configMessage, 10)
	var subProvisionEndpt SubProvisionEndpt

	fmt.Println("SimApp started")
	InitConfigFactory("./config/simapp.yaml", configMsgChan, &subProvisionEndpt)
	go sendMessage(configMsgChan, subProvisionEndpt)
	for {
		time.Sleep(100 * time.Second)
	}
}

func sendMessage(msgChan chan configMessage, subProvisionEndpt SubProvisionEndpt ) {
	var devGroupHttpend string
	var networkSliceHttpend string
	var subscriberHttpend string

	fmt.Println("Subscriber Provision Endpoint in sendMessage:")
	fmt.Println("Address ", subProvisionEndpt.Addr)
	fmt.Println("Port ", subProvisionEndpt.Port)


	for {
		ip, err := net.ResolveIPAddr("ip", subProvisionEndpt.Addr)
		if err != nil {
			fmt.Println("failed to resolve name")
			time.Sleep(10 * time.Second)
			continue
		}
		fmt.Println("webui running at ", ip.String())
		devGroupHttpend = "http://" + ip.String() + ":9089/config/v1/device-group/"
		fmt.Println("device trigger  http endpoint ", devGroupHttpend)
		networkSliceHttpend = "http://" + ip.String() + ":9089/config/v1/network-slice/"
		fmt.Println("network slice http endpoint ", devGroupHttpend)
		subscriberHttpend = "http://" + ip.String() + ":" + subProvisionEndpt.Port + "/api/subscriber/imsi-"
		fmt.Println("subscriber http endpoint ", subscriberHttpend)
		break
	}
	for {
		select {
		case msg := <-msgChan:
			var httpend string
			fmt.Println("Received Message from Channel")
			switch msg.msgType {
			case device_group:
				httpend = devGroupHttpend + msg.name
			case network_slice:
				httpend = networkSliceHttpend + msg.name
			case subscriber:
				httpend = subscriberHttpend + msg.name
			}

			for {
				fmt.Println("Post Message to ", httpend )
				resp, err := http.Post(httpend, "application/json", msg.msgPtr)
				//Handle Error
				if err != nil {
					fmt.Printf("An Error Occured %v", err)
					time.Sleep(1 * time.Second)
					continue
				}
				defer resp.Body.Close()
				//Read the response body
				_, err = ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					time.Sleep(1 * time.Second)
					continue
				}
				fmt.Printf("Message Post %v Success\n", httpend)
				break
			}
		}
	}

}
