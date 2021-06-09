package types

type Assistant struct {
	UID   string `json:"uid"`
	Phone string `json:"phone"`
	Wxid  string `json:"wxid,omitempty"`
}

type DroidVirt struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Wxid       string `json:"wxid,omitempty"`
	GatewayVNC string `json:"gatewayVNC,omitempty"`
	InternalIP string `json:"internalIP,omitempty"`
}
