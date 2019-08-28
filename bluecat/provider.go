package bluecat

import (
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/tiaguinho/gosoap"
)

// Provider returns a terraform resource provider
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"static_ipv4_address": resourceStaticIPv4Address(),
		},
	}
}

var mutex = &sync.Mutex{}

func logoutClientIfError(client *gosoap.Client, err error, msg string) error {
	if err != nil {
		var result error
		result = multierror.Append(err)
		params := gosoap.Params{}

		if _, lerr := client.Call("logout", params); err != nil {
			result = multierror.Append(lerr)
		}
		return fmt.Errorf(msg, result)
	}
	return nil
}

// ObjectTypes contains all valid object types in the BlueCat API
var ObjectTypes = []string{
	"Entity",
	"Configuration",
	"View",
	"Zone",
	"InternalRootZone",
	"ZoneTemplate",
	"EnumZone",
	"EnumNumber",
	"HostRecord",
	"AliasRecord",
	"MXRecord",
	"TXTRecord",
	"SRVRecord",
	"GenericRecord",
	"HINFORecord",
	"NAPTRRecord",
	"RecordWithLink",
	"ExternalHostRecord",
	"StartOfAuthority",
	"IP4Block",
	"IP4Network",
	"IP6Block",
	"IP6Network",
	"IP6Address",
	"IP4NetworkTemplate",
	"DHCP4Range",
	"DHCP6Range",
	"IP4Address",
	"MACPool",
	"DenyMACPool",
	"MACAddress",
	"TagGroup",
	"Tag",
	"User",
	"UserGroup",
	"Server",
	"ServerGroup",
	"NetworkServerInterface",
	"PublishedServerInterface",
	"NetworkInterface",
	"VirtualInterface",
	"LDAP",
	"Kerberos",
	"KerberosRealm",
	"Radius",
	"TFTPGroup",
	"TFTPFolder",
	"TFTPFile",
	"TFTPDeploymentRole",
	"DNSDeploymentRole",
	"DHCPDeploymentRole",
	"DNSOption",
	"DHCPV4ClientOption",
	"DHCPServiceOption",
	"DHCPRawOption",
	"DNSRawOption",
	"DHCPV6ClientOption",
	"DHCPV6ServiceOption",
	"DHCPV6RawOption",
	"VendorProfile",
	"VendorOptionDef",
	"VendorClientOption",
	"CustomOptionDef",
	"DHCPMatchClass",
	"DHCPSubClass",
	"Device",
	"DeviceType",
	"DeviceSubtype",
	"DeploymentScheduler",
	"IP4ReconciliationPolicy",
	"DNSSECSigningPolicy",
	"IP4IPGroup",
	"ResponsePolicy",
	"TSIGKey",
	"RPZone",
	"Location",
	"InterfaceID",
}
