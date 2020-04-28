package impl

import (
	"encoding/json"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/websocket"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/access"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/db"
	"github.com/open-kingfisher/king-utils/interrupt"
	"github.com/open-kingfisher/king-utils/kit"
	"net/http"
	"os"
)

type terminalSize struct {
	conn     *websocket.Conn
	sizeChan chan *remotecommand.TerminalSize
}

func (t terminalSize) Read(p []byte) (int, error) {
	var reply string
	var msg map[string]uint16
	if err := websocket.Message.Receive(t.conn, &reply); err != nil {
		return 0, err
	}
	if err := json.Unmarshal([]byte(reply), &msg); err != nil {
		return copy(p, reply), nil
	} else {
		t.sizeChan <- &remotecommand.TerminalSize{
			Width:  msg["cols"],
			Height: msg["rows"],
		}
		return 0, nil
	}
}

func (t *terminalSize) Next() *remotecommand.TerminalSize {
	size := <-t.sizeChan
	log.Info("terminal size to width: %s height: %s", size.Width, size.Height)
	return size
}

func Terminal(ws *websocket.Conn) {
	defer func() {
		ws.Close()
		if err := recover(); err != nil {
			log.Errorf("Terminal panic: %s", err)
		}
	}()

	c := ws.Request()
	ws.PayloadType = websocket.BinaryFrame // TextFrame 文本模式改为 BinaryFrame 二进制模式，否则乱码无法正常显示
	containerName := c.FormValue("containerName")
	namespace := c.FormValue("namespace")
	podName := c.FormValue("podName")
	cluster := c.FormValue("cluster")
	clientSet, err := access.Access(cluster)
	if err != nil {
		log.Errorf("Client set error: %v", err)
	}
	config, err := getConfig(cluster)
	if err != nil {
		log.Errorf("Get config error: %v", err)
	}
	if err := Handler(ws, namespace, podName, containerName, "/bin/sh", config, clientSet); err != nil {
		err := Handler(ws, namespace, podName, containerName, "/bin/bash", config, clientSet)
		if err != nil {
			log.Error("Handler /bin/bash: ", err)
		}
	}
	//err = Handler(ws, namespace, podName, containerName, "exit", config, clientSet)
}

func getConfig(clusterId string) (*restclient.Config, error) {
	cluster := common.ClusterDB{}
	if err := db.GetById(common.Cluster, clusterId, &cluster); err != nil {
		return nil, err
	}
	kubeconfig := common.KubeConfigPath + cluster.Id
	// 判断配置文件是否存在，存在不在创建
	if kit.IsExist(kubeconfig) == false {
		// 从数据库中读取kubeconfig内容创建kubeconfig配置文件
		if err := kit.CreateConfig(cluster, kubeconfig); err != nil {
			return nil, err
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	} else {
		return config, nil
	}
}

func Handler(ws *websocket.Conn, namespace, podname, container, cmd string, config *restclient.Config, clientSet *kubernetes.Clientset) error {
	fn := func() error {
		req := clientSet.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(podname).
			Namespace(namespace).
			SubResource("exec")
		//Param("container", container).
		//Param("stdin", "true").
		//Param("stdout", "true").
		//Param("stderr", "true").
		//Param("command", cmd).Param("tty", "true")
		c := make(chan *remotecommand.TerminalSize)

		t := &terminalSize{ws, c}
		req.VersionedParams(
			&v1.PodExecOptions{
				Container: container,
				Command:   []string{cmd},
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
			},
			scheme.ParameterCodec,
		)
		executor, err := remotecommand.NewSPDYExecutor(
			config, http.MethodPost, req.URL(),
		)
		if err != nil {
			return err
		}

		return executor.Stream(remotecommand.StreamOptions{
			Stdin:             t,
			Stdout:            ws,
			Stderr:            ws,
			Tty:               true,
			TerminalSizeQueue: t,
		})
	}
	inFd, isTerminal := term.GetFdInfo(ws)
	if !isTerminal {
		if f, err := os.Open("/dev/tty"); err == nil {
			defer f.Close()
			inFd = f.Fd()
			isTerminal = term.IsTerminal(inFd)
		}
		log.Info("isTerminal: ", isTerminal)
	}
	state, err := term.SaveState(inFd)
	if err != nil {
		log.Error("Terminal save state error: ", err)
	}
	return interrupt.Chain(nil, func() {
		if err := term.RestoreTerminal(inFd, state); err != nil {
			log.Error("RestoreTerminal error: ", err)
		}
	}).Run(fn)
}
