package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/umich-vci/gobam"
)

// IP4NetworkModel describes the data model the built-in properties for an IP4Network object.
type IP4NetworkModel struct {
	// These are exposed via the entity properties field for objects of type IP4Network
	CIDR                      types.String
	Template                  types.Int64
	Gateway                   types.String
	DefaultDomains            types.Set
	DefaultView               types.Int64
	DNSRestrictions           types.Set
	AllowDuplicateHost        types.Bool
	PingBeforeAssign          types.Bool
	InheritAllowDuplicateHost types.Bool
	InheritPingBeforeAssign   types.Bool
	InheritDNSRestrictions    types.Bool
	InheritDefaultDomains     types.Bool
	InheritDefaultView        types.Bool
	LocationCode              types.String
	LocationInherited         types.Bool
	SharedNetwork             types.String

	// these are user defined fields that are not built-in
	UserDefinedFields types.Map
}

func flattenIP4NetworkProperties(e *gobam.APIEntity) (*IP4NetworkModel, diag.Diagnostics) {
	var d diag.Diagnostics

	if e == nil {
		d.AddError("invalid input to flattenIP4Network", "entity passed was nil")
		return nil, d
	}
	if e.Type == nil {
		d.AddError("invalid input to flattenIP4Network", "type of entity passed was nil")
		return nil, d
	} else if *e.Type != "IP4Network" {
		d.AddError("invalid input to flattenIP4Network", fmt.Sprintf("type of entity passed was %s", *e.Type))
		return nil, d
	}

	i := &IP4NetworkModel{}
	udfMap := make(map[string]attr.Value)

	var defaultDomainsSet basetypes.SetValue
	var dnsRestrictionsSet basetypes.SetValue
	defaultDomainsFound := false
	dnsRestrictionsFound := false

	if e.Properties != nil {
		props := strings.Split(*e.Properties, "|")
		for x := range props {
			if len(props[x]) > 0 {
				prop := strings.Split(props[x], "=")[0]
				val := strings.Split(props[x], "=")[1]

				switch prop {
				case "name":
					// we ignore the name because it is already a top level parameter
				case "CIDR":
					i.CIDR = types.StringValue(val)
				case "template":
					t, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						d.AddError("error parsing template to int64", err.Error())
						break
					}
					i.Template = types.Int64Value(t)
				case "gateway":
					i.Gateway = types.StringValue(val)
				case "defaultDomains":
					defaultDomainsFound = true
					var ddDiag diag.Diagnostics
					defaultDomains := strings.Split(val, ",")
					defaultDomainsList := []attr.Value{}
					for x := range defaultDomains {
						dID, err := strconv.ParseInt(defaultDomains[x], 10, 64)
						if err != nil {
							d.AddError("error parsing defaultDomains to int64", err.Error())
							break
						}
						defaultDomainsList = append(defaultDomainsList, types.Int64Value(dID))
					}

					defaultDomainsSet, ddDiag = basetypes.NewSetValue(types.Int64Type, defaultDomainsList)
					if ddDiag.HasError() {
						d.Append(ddDiag...)
						break
					}
				case "defaultView":
					dv, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						d.AddError("error parsing defaultView to int64", err.Error())
						break
					}
					i.DefaultView = types.Int64Value(dv)
				case "dnsRestrictions":
					dnsRestrictionsFound = true
					var drDiag diag.Diagnostics
					dnsRestrictions := strings.Split(val, ",")
					didList := []attr.Value{}
					for x := range dnsRestrictions {
						dID, err := strconv.ParseInt(dnsRestrictions[x], 10, 64)
						if err != nil {
							d.AddError("error parsing dnsRestrictions to int64", err.Error())
							break
						}
						didList = append(didList, types.Int64Value(dID))
					}
					dnsRestrictionsSet, drDiag = basetypes.NewSetValue(types.Int64Type, didList)
					if drDiag.HasError() {
						d.Append(drDiag...)
					}
				case "allowDuplicateHost":
					i.AllowDuplicateHost = types.BoolPointerValue(enableDisableToBool(val))
				case "pingBeforeAssign":
					i.PingBeforeAssign = types.BoolPointerValue(enableDisableToBool(val))
				case "inheritAllowDuplicateHost":
					b, err := strconv.ParseBool(val)
					if err != nil {
						d.AddError("error parsing inheritAllowDuplicateHost to bool", err.Error())
						break
					}
					i.InheritAllowDuplicateHost = types.BoolValue(b)
				case "inheritPingBeforeAssign":
					b, err := strconv.ParseBool(val)
					if err != nil {
						d.AddError("error parsing inheritPingBeforeAssign to bool", err.Error())
						break
					}
					i.InheritPingBeforeAssign = types.BoolValue(b)
				case "inheritDNSRestrictions":
					b, err := strconv.ParseBool(val)
					if err != nil {
						d.AddError("error parsing inheritDNSRestrictions to bool", err.Error())
						break
					}
					i.InheritDNSRestrictions = types.BoolValue(b)
				case "inheritDefaultDomains":
					b, err := strconv.ParseBool(val)
					if err != nil {
						d.AddError("error parsing inheritDefaultDomains to bool", err.Error())
						break
					}
					i.InheritDefaultDomains = types.BoolValue(b)
				case "inheritDefaultView":
					b, err := strconv.ParseBool(val)
					if err != nil {
						d.AddError("error parsing inheritDefaultView to bool", err.Error())
						break
					}
					i.InheritDefaultView = types.BoolValue(b)
				case "locationCode":
					i.LocationCode = types.StringValue(val)
				case "locationInherited":
					b, err := strconv.ParseBool(val)
					if err != nil {
						d.AddError("error parsing locationInherited to bool", err.Error())
						break
					}
					i.LocationInherited = types.BoolValue(b)
				case "sharedNetwork":
					i.SharedNetwork = types.StringValue(val)
				default:
					udfMap[prop] = types.StringValue(val)
				}
			}
		}
	}

	if !dnsRestrictionsFound {
		dnsRestrictionsSet = basetypes.NewSetNull(types.Int64Type)
	}
	i.DNSRestrictions = dnsRestrictionsSet

	if !defaultDomainsFound {
		defaultDomainsSet = basetypes.NewSetNull(types.Int64Type)
	}
	i.DefaultDomains = defaultDomainsSet

	var userDefinedFields basetypes.MapValue
	userDefinedFields, udfDiag := basetypes.NewMapValue(types.StringType, udfMap)
	if udfDiag.HasError() {
		d.Append(udfDiag...)
	}
	i.UserDefinedFields = userDefinedFields

	return i, d
}

func enableDisableToBool(s string) *bool {
	var val *bool

	switch s {
	case "enable":
		val = new(bool)
		*val = true
	case "disable":
		val = new(bool)
		*val = false
	default:
		val = nil
	}
	return val
}

func boolToEnableDisable(b *bool) string {
	var s string

	if b == nil {
		s = ""
	} else if *b {
		s = "enable"
	} else {
		s = "disable"
	}
	return s
}
