package container

const (
	RUNNING string = "running"
	STOP    string = "stopped"
	Exit    string = "exited"
)
const (
	DefaultInfoLocation string = "/var/run/miniDocker/info/%s/"
	ConfigFileName      string = "config.json"
	LogFileName         string = "container.log"
)

type Info struct {
	Pid         string `json:"pid"`        // 容器的init进程在宿主机上的 PID
	Id          string `json:"id"`         // 容器Id
	Name        string `json:"name"`       // 容器名
	Command     string `json:"command"`    // 容器内init运行命令
	CreatedTime string `json:"createTime"` // 创建时间
	Status      string `json:"status"`     // 容器的状态
	Volume      string `json:"volume"`     // 容器目录挂载
}
