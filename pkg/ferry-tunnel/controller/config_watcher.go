package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"sync"

	"github.com/ferry-proxy/ferry/pkg/utils/objref"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type ConfigWatcher struct {
	clientset     *kubernetes.Clientset
	namespace     string
	labelSelector string
	logger        logr.Logger
	mut           sync.Mutex
	cache         map[string][]json.RawMessage
	reloadFunc    func(d []json.RawMessage)
}

type ConfigWatcherConfig struct {
	Clientset     *kubernetes.Clientset
	Logger        logr.Logger
	Namespace     string
	LabelSelector string
	ReloadFunc    func(d []json.RawMessage)
}

func NewConfigWatcher(conf *ConfigWatcherConfig) *ConfigWatcher {
	n := &ConfigWatcher{
		clientset:     conf.Clientset,
		namespace:     conf.Namespace,
		labelSelector: conf.LabelSelector,
		logger:        conf.Logger,
		cache:         map[string][]json.RawMessage{},
		reloadFunc:    conf.ReloadFunc,
	}
	return n
}

func (c *ConfigWatcher) Run(ctx context.Context) error {
	infor := informers.NewSharedInformerFactoryWithOptions(c.clientset, 0,
		informers.WithNamespace(c.namespace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = c.labelSelector
		}),
	).Core().V1().ConfigMaps().Informer()
	infor.AddEventHandler(c)
	infor.Run(ctx.Done())
	return nil
}

func (c *ConfigWatcher) Reload() {
	c.mut.Lock()
	defer c.mut.Unlock()

	sum := []json.RawMessage{}
	uniq := map[string]struct{}{}
	for _, data := range c.cache {
		for _, d := range data {
			if _, ok := uniq[string(d)]; !ok {
				uniq[string(d)] = struct{}{}
				sum = append(sum, d)
			}
		}
	}
	sort.SliceStable(sum, func(i, j int) bool {
		return string(sum[i]) < string(sum[j])
	})

	c.reloadFunc(sum)
}

func (c *ConfigWatcher) update(cm *corev1.ConfigMap) {
	data := make([]json.RawMessage, 0, len(cm.Data)+len(cm.BinaryData))
	for key, content := range cm.Data {
		tmp := []json.RawMessage{}
		err := json.Unmarshal([]byte(content), &tmp)
		if err != nil {
			c.logger.Error(err, "unmarshal context failed",
				"configmap", objref.KObj(cm),
				"key", key,
				"context", content,
			)
			continue
		}
		for _, item := range tmp {
			v, err := shrinkJSON(item)
			if err != nil {
				c.logger.Error(err, "shrink json failed",
					"configmap", objref.KObj(cm),
					"key", key,
					"item", item,
				)
				continue
			}
			data = append(data, v)
		}
	}
	for key, content := range cm.BinaryData {
		tmp := []json.RawMessage{}
		err := json.Unmarshal(content, &tmp)
		if err != nil {
			c.logger.Error(err, "unmarshal context failed",
				"configmap", objref.KObj(cm),
				"key", key,
				"context", string(content),
			)
			continue
		}
		for _, item := range tmp {
			v, err := shrinkJSON(item)
			if err != nil {
				c.logger.Error(err, "shrink json failed",
					"configmap", objref.KObj(cm),
					"key", key,
					"item", item,
				)
				continue
			}
			data = append(data, v)
		}
	}

	defer c.Reload()
	c.mut.Lock()
	defer c.mut.Unlock()

	c.cache[cm.Name] = data
}

func (c *ConfigWatcher) delete(cm *corev1.ConfigMap) {
	defer c.Reload()
	c.mut.Lock()
	defer c.mut.Unlock()
	delete(c.cache, cm.Name)
}

func (c *ConfigWatcher) OnAdd(obj interface{}) {
	cm := obj.(*corev1.ConfigMap)
	c.logger.Info("add configmap", "configmap", objref.KObj(cm))
	c.update(cm)
}

func (c *ConfigWatcher) OnUpdate(oldObj, newObj interface{}) {
	cm := newObj.(*corev1.ConfigMap)
	c.logger.Info("update configmap", "configmap", objref.KObj(cm))
	c.update(cm)
}

func (c *ConfigWatcher) OnDelete(obj interface{}) {
	cm := obj.(*corev1.ConfigMap)
	c.logger.Info("delete configmap", "configmap", objref.KObj(cm))
	c.delete(cm)
}

func shrinkJSON(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := json.Indent(&buf, src, "", "")
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
