module github.com/fairwindsops/nova

go 1.15

require (
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20200103221440-774c71fcf114 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.1.2
	k8s.io/apiextensions-apiserver v0.18.1 // indirect
	k8s.io/apimachinery v0.18.1
	k8s.io/client-go v1.5.1
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.2
)

// FIXME: we shouldn't need a replace directive....
replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)
