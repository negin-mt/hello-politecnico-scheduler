package hello

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const Name = "HelloPolitecnico"

// HelloPolitecnico is a minimal Filter plugin for teaching demos.
type HelloPolitecnico struct{}

var _ framework.FilterPlugin = &HelloPolitecnico{}

// New is the plugin factory registered with the scheduler.
func New(_ context.Context, _ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &HelloPolitecnico{}, nil
}

func (pl *HelloPolitecnico) Name() string {
	return Name
}

// Filter logs a hello message and always accepts the node.
func (pl *HelloPolitecnico) Filter(
	_ context.Context,
	_ *framework.CycleState,
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo,
) *framework.Status {
	nodeName := "unknown"
	if nodeInfo != nil && nodeInfo.Node() != nil {
		nodeName = nodeInfo.Node().Name
	}

	klog.Infof("Hello Politecnico: evaluating node [%s] for pod [%s]", nodeName, pod.Name)
	return framework.NewStatus(framework.Success)
}
