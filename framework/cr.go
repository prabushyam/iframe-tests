package framework

import (
	"log"

	"github.com/ardielle/ardielle-go/rdl"
	"github.com/yahoo/athenz/clients/go/zms"
	athenz_domain "github.com/yahoo/k8s-athenz-syncer/pkg/apis/athenz/v1"
	athenzClientset "github.com/yahoo/k8s-athenz-syncer/pkg/client/clientset/versioned"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func CreateCrd(config *rest.Config) error {
	crd := &v1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "athenzdomains.athenz.io",
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group: "athenz.io",
			Scope: v1beta1.ClusterScoped,
			Versions: []v1beta1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural:     "athenzdomains",
				Singular:   "athenzdomain",
				Kind:       "AthenzDomain",
				ShortNames: []string{"domain"},
				ListKind:   "AthenzDomainList",
			},
		},
	}

	rClientset, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return err
	}
	created, err := rClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		return err
	}
	log.Println("created:", created)

	got, err := rClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get("athenzdomains.athenz.io", metav1.GetOptions{})
	if err != nil {
		return err
	}
	log.Println("got:", got)
	return nil
}

func CreateCr(config *rest.Config) {
	domain := "home.foo"
	fakeDomain := getFakeDomain()
	newCR := &athenz_domain.AthenzDomain{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AthenzDomain",
			APIVersion: "athenz.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: domain,
		},
		Spec: athenz_domain.AthenzDomainSpec{
			SignedDomain: fakeDomain,
		},
		Status: athenz_domain.AthenzDomainStatus{
			Message: "",
		},
	}

	versiondClient, err := athenzClientset.NewForConfig(config)
	if err != nil {
		log.Println(err)
		return
	}

	list, err := versiondClient.AthenzV1().AthenzDomains().List(metav1.ListOptions{})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("list:", list)

	created, err := versiondClient.AthenzV1().AthenzDomains().Create(newCR)
	if err != nil {
		log.Println("error creating athenz domain:", err)
		return
	}
	log.Println("created cr:", created)

	got, err := versiondClient.AthenzV1().AthenzDomains().Get(domain, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("got cr:", got)
}

func getFakeDomain() zms.SignedDomain {
	allow := zms.ALLOW
	timestamp, err := rdl.TimestampParse("2019-06-21T19:28:09.305Z")
	if err != nil {
		panic(err)
	}

	domainName := "home.foo"
	username := "user.foo"
	return zms.SignedDomain{
		Domain: &zms.DomainData{
			Modified: timestamp,
			Name:     zms.DomainName(domainName),
			Policies: &zms.SignedPolicies{
				Contents: &zms.DomainPolicies{
					Domain: zms.DomainName(domainName),
					Policies: []*zms.Policy{
						{
							Assertions: []*zms.Assertion{
								{
									Role:     domainName + ":role.admin",
									Resource: domainName + ".test:*",
									Action:   "*",
									Effect:   &allow,
								},
							},
							Modified: &timestamp,
							Name:     zms.ResourceName(domainName + ":policy.admin"),
						},
					},
				},
				KeyId:     "col-env-1.1",
				Signature: "signature-policy",
			},
			Roles: []*zms.Role{
				{
					Members:  []zms.MemberName{zms.MemberName(username)},
					Modified: &timestamp,
					Name:     zms.ResourceName(domainName + ":role.admin"),
					RoleMembers: []*zms.RoleMember{
						{
							MemberName: zms.MemberName(username),
						},
					},
				},
				{
					Trust:    "parent.domain",
					Modified: &timestamp,
					Name:     zms.ResourceName(domainName + ":role.trust"),
				},
			},
			Services: []*zms.ServiceIdentity{},
			Entities: []*zms.Entity{},
		},
		KeyId:     "colo-env-1.1",
		Signature: "signature",
	}
}
