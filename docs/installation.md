---
meta:
  - name: description
    content: "Fairwinds Nova | Install Documentation"
---
## Installation

### From GitHub Releases
Visit the [releases page](https://github.com/FairwindsOps/nova/releases) to find the release
that's right for your environment. For example, on Linux:
```
curl -L "https://github.com/FairwindsOps/nova/releases/download/3.2.0/nova_3.2.0_linux_amd64.tar.gz" > nova.tar.gz
tar -xvf nova.tar.gz
sudo mv nova /usr/local/bin/
```

### asdf-vm
```
asdf plugin-add nova
asdf install nova latest
asdf local nova latest
```

### Homebrew
```
brew tap fairwindsops/tap
brew install fairwindsops/tap/nova
```

### From source
```
go install github.com/fairwindsops/nova@latest
```
