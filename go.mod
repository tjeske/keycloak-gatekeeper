module github.com/tjeske/keycloak-gatekeeper

go 1.14

require (
	github.com/PuerkitoBio/purell v1.1.0
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/armon/go-proxyproto v0.0.0-20180202201750-5b7edb60ff5f
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/coreos/go-oidc v0.0.0-20171020180921-e860bd55bfa7
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/elazarl/goproxy v0.0.0-20200220113713-29f9e0ba54ea
	github.com/etcd-io/bbolt v1.3.3
	github.com/fsnotify/fsnotify v1.4.7
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/go-chi/chi v4.0.0+incompatible
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348 // indirect
	github.com/rs/cors v1.6.0
	github.com/satori/go.uuid v1.2.0
	github.com/stretchr/testify v1.4.0
	github.com/unrolled/secure v0.0.0-20181221173256-0d6b5bb13069
	github.com/urfave/cli v1.22.2
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	gopkg.in/bsm/ratelimit.v1 v1.0.0-20160220154919-db14e161995a // indirect
	gopkg.in/redis.v4 v4.2.4
	gopkg.in/resty.v1 v1.10.3
	gopkg.in/yaml.v2 v2.2.7
)

require (
	github.com/Nerzal/gocloak/v4 v4.8.0
	github.com/docker/cli v0.0.0-20200210162036-a4bedce16568
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/docker/go-units v0.4.0
	github.com/go-chi/docgen v1.0.5
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.1
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.1.2
	github.com/tjeske/containerflight v0.3.0
	go.mongodb.org/mongo-driver v1.3.1
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc
)

replace github.com/spf13/pflag => github.com/thaJeztah/pflag v1.0.3-0.20180821151913-4cb166e4f25a

replace github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20191014053712-acdcf13d5eaf

replace github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c => github.com/docker/docker v1.4.2-0.20200212114129-58c261520896

replace github.com/tonistiigi/fsutil v0.0.0-20190819224149-3d2716dd0a4d => github.com/tonistiigi/fsutil v0.0.0-20191018213012-0f039a052ca1

replace github.com/xeipuuv/gojsonschema => github.com/xeipuuv/gojsonschema v0.0.0-20170528113821-0c8571ac0ce1

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
