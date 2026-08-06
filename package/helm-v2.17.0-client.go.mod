module k8s.io/helm

go 1.26.0

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/goutils v1.1.0
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Masterminds/vcs v1.13.3
	github.com/PuerkitoBio/purell v1.1.1
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/cpuguy83/go-md2man v1.0.10
	github.com/cyphar/filepath-securejoin v0.2.2
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96
	github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/evanphx/json-patch v4.2.0+incompatible
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d
	github.com/fatih/color v1.15.0
	github.com/ghodss/yaml v1.0.1-0.20180820084758-c7ce16629ff4
	github.com/go-openapi/jsonpointer v0.19.2
	github.com/go-openapi/jsonreference v0.18.0
	github.com/go-openapi/spec v0.19.2
	github.com/go-openapi/swag v0.19.2
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/flock v0.7.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.4
	github.com/google/btree v1.1.3
	github.com/google/go-cmp v0.7.0
	github.com/google/gofuzz v1.0.0
	github.com/google/uuid v1.6.0
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d
	github.com/gosuri/uitable v0.0.3
	github.com/gregjones/httpcache v0.0.0-20170728041850-787624de3eb7
	github.com/hashicorp/golang-lru v0.5.1
	github.com/huandu/xstrings v1.2.0
	github.com/imdario/mergo v0.3.5
	github.com/jmoiron/sqlx v1.2.0
	github.com/json-iterator/go v1.1.12
	github.com/lib/pq v1.2.1-0.20191011153232-f91d3411e481
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/mailru/easyjson v0.0.0-20190312143242-1de009706dbe
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.19
	github.com/mattn/go-runewidth v0.0.5
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/go-wordwrap v1.0.0
	github.com/mitchellh/reflectwalk v1.0.1
	github.com/moby/term v0.5.2
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.2
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rubenv/sql-migrate v0.0.0-20191025130928-9355dd04f4b3
	github.com/russross/blackfriday v1.5.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245
	golang.org/x/crypto v0.54.0
	golang.org/x/net v0.57.0
	golang.org/x/oauth2 v0.36.0
	golang.org/x/sync v0.22.0
	golang.org/x/sys v0.47.0
	golang.org/x/term v0.45.0
	golang.org/x/text v0.40.0
	golang.org/x/time v0.11.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478
	google.golang.org/grpc v1.82.1
	google.golang.org/protobuf v1.36.11
	gopkg.in/gorp.v1 v1.7.2
	gopkg.in/inf.v0 v0.9.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/klog v0.4.0
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1
	sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/BurntSushi/toml => github.com/BurntSushi/toml v0.3.1
	github.com/MakeNowJust/heredoc => github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/goutils => github.com/Masterminds/goutils v1.1.0
	github.com/Masterminds/semver => github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig => github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Masterminds/vcs => github.com/Masterminds/vcs v1.13.3
	github.com/PuerkitoBio/purell => github.com/PuerkitoBio/purell v1.0.0
	github.com/PuerkitoBio/urlesc => github.com/PuerkitoBio/urlesc v0.0.0-20160726150825-5bd2802263f2
	github.com/asaskevich/govalidator => github.com/asaskevich/govalidator v0.0.0-20160518190739-766470278477
	github.com/cpuguy83/go-md2man => github.com/cpuguy83/go-md2man v1.0.10
	github.com/cyphar/filepath-securejoin => github.com/cyphar/filepath-securejoin v0.2.2
	github.com/davecgh/go-spew => github.com/davecgh/go-spew v1.1.1
	github.com/docker/docker => ./pkg/pasturestack/shims/docker-term
	github.com/docker/spdystream => github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96
	github.com/emicklei/go-restful => github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/evanphx/json-patch => github.com/evanphx/json-patch v4.2.0+incompatible
	github.com/exponent-io/jsonpath => github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d
	github.com/fatih/color => github.com/fatih/color v1.7.1-0.20181010231311-3f9d52f7176a
	github.com/ghodss/yaml => github.com/ghodss/yaml v1.0.1-0.20180820084758-c7ce16629ff4
	github.com/go-openapi/jsonpointer => github.com/go-openapi/jsonpointer v0.0.0-20160704185906-46af16f9f7b1
	github.com/go-openapi/jsonreference => github.com/go-openapi/jsonreference v0.0.0-20160704190145-13c6e3589ad9
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.0.0-20160808142527-6aced65f8501
	github.com/go-openapi/swag => github.com/go-openapi/swag v0.0.0-20160704191624-1d0bd113de87
	github.com/gobwas/glob => github.com/gobwas/glob v0.2.3
	github.com/gofrs/flock => github.com/gofrs/flock v0.7.1
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.4
	github.com/google/btree => github.com/google/btree v1.0.0
	github.com/google/go-cmp => github.com/google/go-cmp v0.3.0
	github.com/google/gofuzz => github.com/google/gofuzz v1.0.0
	github.com/google/uuid => github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d
	github.com/gosuri/uitable => github.com/gosuri/uitable v0.0.3
	github.com/gregjones/httpcache => github.com/gregjones/httpcache v0.0.0-20170728041850-787624de3eb7
	github.com/hashicorp/golang-lru => github.com/hashicorp/golang-lru v0.5.1
	github.com/huandu/xstrings => github.com/huandu/xstrings v1.2.0
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.5
	github.com/jmoiron/sqlx => github.com/jmoiron/sqlx v1.2.0
	github.com/json-iterator/go => github.com/json-iterator/go v1.1.7
	github.com/lib/pq => github.com/lib/pq v1.2.1-0.20191011153232-f91d3411e481
	github.com/liggitt/tabwriter => github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/mailru/easyjson => github.com/mailru/easyjson v0.0.0-20160728113105-d5b7844b561a
	github.com/mattn/go-colorable => github.com/mattn/go-colorable v0.1.4
	github.com/mattn/go-isatty => github.com/mattn/go-isatty v0.0.11-0.20191009155615-0e9ddb7c0c0a
	github.com/mattn/go-runewidth => github.com/mattn/go-runewidth v0.0.5
	github.com/mitchellh/copystructure => github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/go-wordwrap => github.com/mitchellh/go-wordwrap v1.0.0
	github.com/mitchellh/reflectwalk => github.com/mitchellh/reflectwalk v1.0.1
	github.com/moby/term => github.com/moby/term v0.5.2
	github.com/modern-go/concurrent => github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 => github.com/modern-go/reflect2 v1.0.1
	github.com/peterbourgon/diskv => github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors => github.com/pkg/errors v0.8.2-0.20190227000051-27936f6d90f9
	github.com/rubenv/sql-migrate => github.com/rubenv/sql-migrate v0.0.0-20191025130928-9355dd04f4b3
	github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2
	github.com/spf13/cobra => github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag => github.com/spf13/pflag v1.0.5
	github.com/technosophos/moniker => github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245
	golang.org/x/crypto => golang.org/x/crypto v0.54.0
	golang.org/x/net => golang.org/x/net v0.57.0
	golang.org/x/oauth2 => golang.org/x/oauth2 v0.36.0
	golang.org/x/sync => golang.org/x/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190124100055-b90733256f2e
	golang.org/x/term => golang.org/x/term v0.45.0
	golang.org/x/text => golang.org/x/text v0.40.0
	golang.org/x/time => golang.org/x/time v0.0.0-20161028155119-f51c12702a4d
	google.golang.org/genproto/googleapis/rpc => google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478
	google.golang.org/grpc => google.golang.org/grpc v1.82.1
	google.golang.org/protobuf => google.golang.org/protobuf v1.36.11
	gopkg.in/gorp.v1 => gopkg.in/gorp.v1 v1.7.2
	gopkg.in/inf.v0 => gopkg.in/inf.v0 v0.9.0
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.4
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/klog => k8s.io/klog v0.4.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	k8s.io/utils => k8s.io/utils v0.0.0-20190801114015-581e00157fb1
	sigs.k8s.io/kustomize => sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.1.0
)
