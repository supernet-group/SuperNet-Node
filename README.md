# SuperNet-Node
Share your unused computing capacity to provide support for more AI creators in need and earn profits at the same time.

## Recommended configuration
### Software
- Operating System: Ubuntu 22.04
- Go: 1.21.0
- Git
### Hardware
- CPU: 4 cores
- Memory: 8GB
- Disk: 100GB
- GPU: Nvidia RTX 3090
- Network: 1Gbps

## Necessary software packages
Please ensure that the following software packages are installed before use.

- curl
- speedtest-cli
- docker
- nginx

***During the SuperNet-node program execution, all Docker containers will be stopped and the nginx configuration file will be reset. If necessary, please make a backup in advance.***

Additional software packages required for GPU devices.
- nvidia-docker2

## Instructions for use

***Please make sure to switch to the root account before starting.***

1. Clone the repository locally.

```
git clone https://github.com/supernet-group/SuperNet-Node.git
```

2. Compile binary executable file.

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/SuperNet main.go
```

3. Write configuration file.

Create config.yml file in the same directory as the executable file.
```
cd bin &&
cat << EOF > config.yml
base:
  rpc:
  # Mnemonics used to complete transactions
  privateKey:
  # The level of privacy protection provided
  securityLevel: 0
console:
  # Directory provided for training models. default: /data/super
  workDirectory:
  # default: https://supernet-ipfs.distri.ai
  ipfsNodeUrl:
  # If using a proxy, please specify the public IP address. If not using a proxy, you can ignore this.
  publicIP:
  # The port on which the SuperNet listens. default: 13010
  # Make sure the public network is accessible
  superPort:
  # The port on which the workspace or deploy listens. default: 13011
  workPort:
  # The port on which the Local server listens. default: 13012
  serverPort:
  # Optional, ease of use for users to deploy tools such as visualization or data tracking.
  # Make sure the public network is accessible
  publicPortExpand1:
  publicPortExpand2:
  publicPortExpand3:
EOF
```

4. Run executable file.

```
./SuperNet node start
```

- If you have the following error，please check your account for sufficient SOL and DIST.

![amount](https://github.com/supernet-group/SuperNet-Node/assets/122685398/fbc25da5-486b-4c4f-87b6-b555057ee5e7)

- The following information indicates success.

![success](https://github.com/supernet-group/SuperNet-Node/assets/122685398/0c87c803-cf49-42b0-962d-fde82219116b)
