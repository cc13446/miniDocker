package network

import (
	"cc.com/miniDocker/container"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

var (
	defaultNetworkPath = "/var/run/miniDocker/network/network"
	drivers            = map[string]Driver{}
	networks           = map[string]*Network{}
)

// Endpoint 端点
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	Network     *Network
	PortMapping []string
}

// Network 网络
type Network struct {
	Name    string
	IPRange *net.IPNet
	Driver  string
}

// dump 保存网络
func (nw *Network) dump(dumpPath string) error {
	// 检查保存的目录是否存在，不存在则创建
	if _, err := os.Stat(dumpPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(dumpPath, 0644); err != nil {
			return fmt.Errorf("create network dump path %s error: %v", dumpPath, err)
		}
	}

	// 保存的文件名是网络的名字
	nwPath := path.Join(dumpPath, nw.Name)

	// 打开保存的文件用于写入,后面打开的模式参数分别是存在内容则清空、只写入、不存在则创建
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open file %s failed %v", dumpPath, err)
	}
	defer func(nwFile *os.File) {
		if err := nwFile.Close(); err != nil {
			log.Errorf("Close file %s error %v", nwPath, err)
		}
	}(nwFile)

	nwJson, err := json.Marshal(nw)
	if err != nil {
		return fmt.Errorf("marshal %s error %v", nw, err)
	}

	if _, err = nwFile.Write(nwJson); err != nil {
		return fmt.Errorf("write %s failed %v", nwJson, err)
	}
	return nil
}

// remove 移除网络
func (nw *Network) remove(dumpPath string) error {
	// 检查网络对应的配置文件状态，如果文件不存在就直接返回
	fullPath := path.Join(dumpPath, nw.Name)
	if _, err := os.Stat(fullPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	// 否则删除这个网络对应的配置文件
	return os.Remove(fullPath)
}

// load 加载网络
func (nw *Network) load(dumpPath string) error {
	// 打开配置文件
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer func(nwConfigFile *os.File) {
		if err := nwConfigFile.Close(); err != nil {
			log.Errorf("Close file %s error %v", dumpPath, err)
		}
	}(nwConfigFile)

	// 从配置文件中读取网络 配置 json 符串
	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(nwJson[:n], nw); err != nil {
		return fmt.Errorf("unmarshal %s failed %v", nwJson[:n], err)
	}
	return nil
}

// Driver 网络驱动接口
type Driver interface {
	// Name 网络驱动名字
	Name() string
	// Create 创建网络
	Create(subnet string, name string) (*Network, error)
	// Delete 删除网络
	Delete(network Network) error
	// Connect 将网络和端点连接起来
	Connect(network *Network, endpoint *Endpoint) error
	// Disconnect 断开网络和端点的连接
	Disconnect(network Network, endpoint *Endpoint) error
}

// Init 初始化
func Init() error {

	// 加载网络驱动
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(defaultNetworkPath, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err := filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		// 加载文件名作为网络名
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}

		// 加载网络的配置信息
		if err = nw.load(nwPath); err != nil {
			log.Errorf("error load network: %s", err)
		}
		// 将网络的配置信息加入到 networks 字典中
		networks[nwName] = nw
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// CreateNetwork 创建网络
func CreateNetwork(driver, subnet, name string) error {
	// 将网段的字符串转换成net. IPNet的对象
	_, cidr, _ := net.ParseCIDR(subnet)

	// 通过IPAM分配网关IP，获取到网段中第一个IP作为网关的IP
	ip, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip
	// 调用指定的网络驱动创建网络
	// Create 方法创建网络，后面会以 Bridge 驱动为例介绍它的实现
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	// 保存网络信息，将网络的信息保存在文件系统中，以便查询和在网络上连接网络端点
	return nw.dump(defaultNetworkPath)
}

// ListNetwork 打印网络信息
func ListNetwork() {
	// 通过 tab writer 库把信息打印到屏幕上
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	if _, err := fmt.Fprint(w, "NAME\tIpRange\tDriver\n"); err != nil {
		log.Errorf("Print error %v", err)
	}
	for _, nw := range networks {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IPRange.String(),
			nw.Driver,
		); err != nil {
			log.Errorf("Print error %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
		return
	}
}

// DeleteNetwork 删除网络
func DeleteNetwork(networkName string) error {
	// 网络不存在直接返回一个error
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no Such Network: %s", networkName)
	}
	// 调用IPAM的实例ipAllocator释放网络网关的IP
	if err := ipAllocator.Release(nw.IPRange, &nw.IPRange.IP); err != nil {
		return fmt.Errorf("remove Network gateway ip failed %v", err)
	}
	// 调用网络驱动删除网络创建的设备与配置
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("remove Network DriverError failed %v", err)
	}
	// 最后从网络的配直目录中删除该网络对应的配置文件
	return nw.remove(defaultNetworkPath)
}

// Connect 连接容器到之前创建的网络
func Connect(networkName string, info *container.Info) error {
	// 从networks字典中取到容器连接的网络的信息，networks字典中保存了当前己经创建的网络
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no Such Network: %s", networkName)
	}

	// 分配容器IP地址
	ip, err := ipAllocator.Allocate(network.IPRange)
	if err != nil {
		return fmt.Errorf("allocate ip error %v", err)
	}

	// 创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", info.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: info.PortMapping,
	}
	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	// 到容器的namespace配置容器网络设备IP地址
	if err = configEndpointIpAddressAndRoute(ep, info); err != nil {
		return err
	}
	// 配置端口映射信息
	return configPortMapping(ep)
}

// configEndpointIpAddressAndRoute 配置容器网络端点的地址和路由
func configEndpointIpAddressAndRoute(ep *Endpoint, info *container.Info) error {
	// 根据名字找到对应Veth设备
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}
	// 将容器的网络端点加入到容器的网络空间中
	// 并使这个函数下面的操作都在这个网络空间中进行
	// 执行完函数后，恢复为默认的网络空间
	defer enterContainerNetNS(&peerLink, info)()

	// 获取到容器的IP地址及网段，用于配置容器内部接口地址
	// 比如容器IP是192.168.1.2， 而网络的网段是192.168.1.0/24
	// 那么这里产出的IP字符串就是192.168.1.2/24，用于容器内Veth端点配置
	interfaceIP := *ep.Network.IPRange
	interfaceIP.IP = ep.IPAddress

	// 设置容器内Veth端点的IP
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return err
	}
	// 启动容器内的Veth端点
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}
	// Net Namespace 中默认本地地址 127 的网卡是关闭状态的
	// 启动它以保证容器访问自己的请求
	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	// 设置容器内的外部请求都通过容器内的Veth端点访问
	// 0.0.0.0/0的网段，表示所有的IP地址段
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	// 构建要添加的路由数据，包括网络设备、网关IP及目的网段
	// 相当于route add -net 0.0.0.0/0 gw dev
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IPRange.IP,
		Dst:       cidr,
	}

	// 调用netlink的RouteAdd,添加路由到容器的网络空间
	// 相当于route add 命令
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

// configPortMapping 配置端口映射
func configPortMapping(ep *Endpoint) error {
	var err error
	// 遍历容器端口映射列表
	for _, pm := range ep.PortMapping {
		// 分割成宿主机的端口和容器的端口
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.Errorf("Port mapping format error, %v", pm)
			continue
		}
		// 由于iptables没有Go语言版本的实现，所以采用exec.Command的方式直接调用命令配置
		// 在iptables的PREROUTING中添加DNAT规则
		// 将宿主机的端口请求转发到容器的地址和端口上
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		// 执行iptables命令,添加端口映射转发规则
		if output, err := cmd.Output(); err != nil {
			log.Errorf("iptables Output, %v", output)
			continue
		}
	}
	return err
}

// enterContainerNetNS 将容器的网络端点加入到容器的网络空间中
// 并锁定当前程序所执行的线程，使当前线程进入到容器的网络空间
// 返回值是一个函数指针，执行这个返回函数才会退出容器的网络空间，回归到宿主机的网络空间
func enterContainerNetNS(enLink *netlink.Link, info *container.Info) func() {
	// 找到容器的Net Namespace
	// /proc/[pid]/ns/net 打开这个文件的文件描述符就可以来操作Net Namespace
	// 而ContainerInfo中的PID,即容器在宿主机上映射的进程ID
	// 它对应的/proc/[pid]/ns/net就是容器内部的Net Namespace
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", info.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("Error get container net namespace, %v", err)
	}
	nsFD := f.Fd()

	// 锁定当前程序所执行的线程，如果不锁定操作系统线程的话
	// Go语言的goroutine可能会被调度到别的线程上去
	// 就不能保证一直在所需要的网络空间中了
	// 所以先调用runtime.LockOSThread()锁定当前程序执行的线程
	runtime.LockOSThread()

	// 修改网络端点Veth的另外一端，将其移动到容器的Net Namespace 中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		log.Errorf("Error set link netns , %v", err)
	}

	// 获取当前的网络namespace
	originNetns, err := netns.Get()
	if err != nil {
		log.Errorf("Error get current netns, %v", err)
	}

	// 调用 netns.Set方法，将当前进程加入容器的Net Namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		log.Errorf("Error set netns, %v", err)
	}
	// 返回之前Net Namespace的函数
	// 在容器的网络空间中执行完容器配置之后调用此函数就可以将程序恢复到原生的Net Namespace
	return func() {
		// 恢复到上面获取到的之前的 Net Namespace
		if err := netns.Set(originNetns); err != nil {
			log.Errorf("Error set netns, %v", err)
		}
		if err := originNetns.Close(); err != nil {
			log.Errorf("Error close origin net ns, %v", err)
		}
		// 取消对当附程序的线程锁定
		runtime.UnlockOSThread()
		if err := f.Close(); err != nil {
			log.Errorf("Error close net ns file, %v", err)
		}
	}
}

func Disconnect(networkName string, info *container.Info) error {
	return nil
}
