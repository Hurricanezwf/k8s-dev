package podlog

import (
	"io"
	"io/ioutil"
	"strconv"

	"k8s.io/client-go/kubernetes"
)

// PodLogs 对kubectl logs的简单封装，输出日志到给定的io.Writer
type PodLogs struct {
	// k8s 客户端实例
	k8sCli kubernetes.Interface

	// 命名空间
	namespace string

	// pod名字
	podName string

	// container
	container string

	// 是否启用follow模式
	follow bool

	// Label选择器
	labelSelectors map[string]string

	// 日志输出
	output io.Writer
}

// New 新建Pod日志输出实例
// 如果output为空，则日志将被丢弃
func New(k8sCli kubernetes.Interface, namespace, podName string, output io.Writer) *PodLogs {
	l := &PodLogs{
		k8sCli:    k8sCli,
		namespace: namespace,
		podName:   podName,
		output:    output,
	}
	return l
}

func (l *PodLogs) Container(v string) *PodLogs {
	l.container = v
	return l
}

func (l *PodLogs) Follow(v bool) *PodLogs {
	l.follow = v
	return l
}

func (l *PodLogs) LabelSelector(v map[string]string) *PodLogs {
	l.labelSelectors = v
	return l
}

// Collect 从远程Pod将日志拷贝至output, 如果output为空，则日志将被丢弃
// 接受日志的过程中，该调用将阻塞直到全部接受完成返回nil，或者接收失败返回error
func (l *PodLogs) Collect() error {
	req := l.k8sCli.CoreV1().RESTClient().Get().
		Resource("pods").
		Name(l.podName).
		SubResource("log").
		Param("follow", strconv.FormatBool(l.follow))

	if len(l.namespace) > 0 {
		req.Namespace(l.namespace)
	}
	if len(l.container) > 0 {
		req.Param("container", l.container)
	}

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	if l.output == nil {
		l.output = ioutil.Discard
	}
	_, err = io.Copy(l.output, readCloser)
	return err
}
