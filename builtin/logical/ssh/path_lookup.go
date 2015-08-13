package ssh

import (
	"fmt"
	"net"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathLookup(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "lookup",
		Fields: map[string]*framework.FieldSchema{
			"ip": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "IP address of target",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.WriteOperation: b.pathLookupWrite,
		},
		HelpSynopsis:    pathLookupSyn,
		HelpDescription: pathLookupDesc,
	}
}

func (b *backend) pathLookupWrite(req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	ipAddr := d.Get("ip").(string)
	if ipAddr == "" {
		return logical.ErrorResponse("Missing ip"), nil
	}
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return logical.ErrorResponse(fmt.Sprintf("Invalid IP '%s'", ip.String())), nil
	}

	keys, err := req.Storage.List("roles/")
	if err != nil {
		return nil, err
	}

	var matchingRoles []string
	for _, role := range keys {
		if contains, _ := roleContainsIP(req.Storage, role, ip.String()); contains {
			matchingRoles = append(matchingRoles, role)
		}
	}
	return &logical.Response{
		Data: map[string]interface{}{
			"roles": matchingRoles,
		},
	}, nil
}

const pathLookupSyn = `
Lists 'roles' that can be used to create a dynamic key.
`

const pathLookupDesc = `
The IP address for which the key is requested, is searched in the
CIDR blocks registered with vault using the 'roles' endpoint. Keys
can be generated only by specifying the 'role' name. The roles that
can be used to generate the key for a particular IP, are listed via
this endpoint. For example, if this backend is mounted at "ssh", then
"ssh/lookup" lists the roles associated with keys can be generated
for a target IP, if the CIDR block encompassing the IP, is registered
with vault.
`
