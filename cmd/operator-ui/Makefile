TAG=$(shell cat build.properties | cut -d'=' -f 2)
ifdef IMAGE_TAG
TAG="$(IMAGE_TAG)"
endif

buffalo:
	GO111MODULE=on SYNOPSYS_OPERATOR_VERSION=$(TAG) CONFIG_FILE_PATH=./config.json buffalo dev