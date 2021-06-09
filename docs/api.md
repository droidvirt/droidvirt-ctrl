## 使用流程
* 创建/删除助手（对平台来说，意味这初始化助手对应的数据卷，数据卷中包含着手机设备信息及wx数据）
- 使用相同的手机号创建助手，对于wx来说并不意味着同一个手机号，wx还会根据设备IMEI，mac地址等信息综合判断。因此，删除并重新创建助手后，wx号需要再次登录
* 启动/停止助手（对平台来说，意味着把数据卷挂载并运行在一个虚拟机下，因此，助手的启动与停止不影响助手对应wx的数据，只会影响外部访问的VNC链接）

## 接口

#### 1.查询平台管理下的助手信息
- GET /assistant
- query: { uid: string, phone: string, wxid: string }

#### 2.查询正在运行的助手信息
- GET /droidvirt
- query: { uid: string }