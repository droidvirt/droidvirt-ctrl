package config

var (
	AndroidHookSidecarAnnotation = map[string]string{
		"hooks.kubevirt.io/hookSidecars":  "[{\"image\": \"registry.cn-shanghai.aliyuncs.com/droidvirt/hook-sidecar:base\"}]",
		"vnc.droidvirt.io/port":           "5900",
		"websocket.vnc.droidvirt.io/port": "5901",
	}
)
