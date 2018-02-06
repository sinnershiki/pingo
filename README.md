pingo
--

勉強がてらpingツール実装

# 実行結果例

```
% cat Vagrantfile
Vagrant.configure("2") do |config|
  config.vm.box = "bento/centos-6.9"

  config.vm.define "node1" do |node|
    node.vm.network :private_network, ip: "192.168.10.11"
  end

  config.vm.define "node2" do |node|
    node.vm.network :private_network, ip: "192.168.10.12"
  end
end
% go run ping.go 127.0.0.1 192.168.10.11 192.168.10.12 192.168.10.13
[127.0.0.1 192.168.10.11 192.168.10.12 192.168.10.13]
got reflection from 127.0.0.1:0

got reflection from 192.168.10.11:0

got reflection from 192.168.10.12:0

timeout
```
