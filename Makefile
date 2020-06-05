TAG = $(shell date -u +"%Y%m%d%H%M")

.PHONY: operator

operator:
	docker build --rm -f ./build/Dockerfile -t registry.cn-shanghai.aliyuncs.com/droidvirt/droidvirt-ctrl:$(TAG) .
	docker push registry.cn-shanghai.aliyuncs.com/droidvirt/droidvirt-ctrl:$(TAG)

