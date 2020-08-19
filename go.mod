module github.com/fairwindsops/nova

go 1.13

require (
	// required for dependency issue: https://github.com/kubernetes/client-go/issues/628
	github.com/Azure/go-autorest v12.2.0+incompatible

	github.com/DATA-DOG/go-sqlmock v1.4.1 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/lib/pq v1.5.2 // indirect
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/rubenv/sql-migrate v0.0.0-20200429072036-ae26b214fa43 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	gopkg.in/yaml.v2 v2.2.8
	helm.sh/helm v2.16.6+incompatible
	helm.sh/helm/v3 v3.1.2
	k8s.io/apiextensions-apiserver v0.18.1 // indirect
	k8s.io/client-go v1.5.1
	k8s.io/helm v2.16.5+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.2
)

// FIXME: we shouldn't need a replace directive....
replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)
