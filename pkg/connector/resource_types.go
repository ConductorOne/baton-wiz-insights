package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
)

// issueResourceType represents Wiz security issues synced as security insights.
var issueResourceType = &v2.ResourceType{
	Id:          "security-insight",
	DisplayName: "Security Insight",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_SECURITY_INSIGHT},
	Annotations: annotations.New(
		&v2.CapabilityPermissions{
			Permissions: []*v2.CapabilityPermission{
				{Permission: "read:issues"},
			},
		},
		&v2.SkipEntitlementsAndGrants{},
	),
}
