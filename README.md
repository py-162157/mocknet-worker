# MockNet
一个分布式高性能网络仿真平台，在云端服务器进行一键式高性能仿真网络部署。该平台具有高并发、集群化、云规模等特性，利用云平台和容器化技术，可以在云端进行快速部署和迁移。 实验证明，仅用一台100G高性能交换机和三台40核服务器便可以仿真出bi-section带宽为70G的胖树结构网络。该项目为IEEE INFOCOM论文在投。

## 使用方法
1. 搭建好K8s集群，并在master节点上配置flannel-etcd.yaml。
2. 在本地编译安装带有高性能网卡支持的VPP版本。
3. 在master节点上运行MockNet-Master，在其余的服务器节点上运行MockNet-Worker。
4. 在用户端运行MockNet-CMD，与MockNet-Master连接后进行网络的创建工作。

## Mocknet-Worker
Mocknet-Worker是MockNet仿真系统的工作程序，运行在服务器集群的Worker机器上（即除了K8s的master节点外的其他节点）。
1. binapi里集成了各版本VPP的调用接口。
2. mocknet-agent为主程序。
3. etcd插件与etcd数据库进行交互，进而与集群中的各MockNet-Worker以及MockNet-Master进行通信。
4. controller插件是控制中心，根据从master接收到的网络拓扑信息，调用各子插件进行仿真网络的配置。
5. vpp插件根据controller插件的调用，对VPP的用户态网络空间进行详细的配置工作。
6. linux插件根据controller插件的调用，对仿真节点的linux网络空间进行详细的配置工作。
