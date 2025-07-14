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
	DynamicUpdate             types.Bool

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
				case "dynamicUpdate":
					b, err := strconv.ParseBool(val)
					if err != nil {
						d.AddError("error parsing dynamicUpdate to bool", err.Error())
						break
					}
					i.DynamicUpdate = types.BoolValue(b)
				// these are user defined fields that are not built-in
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

// IP4BlockModel describes the data model the built-in properties for an IP4Block object.
type IP4BlockModel struct {
	// These are exposed via the entity properties field for objects of type IP4Block
	CIDR                      types.String
	DefaultDomains            types.Set
	Start                     types.String
	End                       types.String
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

	// these are user defined fields that are not built-in
	UserDefinedFields types.Map
}

func flattenIP4BlockProperties(e *gobam.APIEntity) (*IP4BlockModel, diag.Diagnostics) {
	var d diag.Diagnostics

	if e == nil {
		d.AddError("invalid input to flattenIP4Block", "entity passed was nil")
		return nil, d
	}
	if e.Type == nil {
		d.AddError("invalid input to flattenIP4Block", "type of entity passed was nil")
		return nil, d
	} else if *e.Type != "IP4Block" {
		d.AddError("invalid input to flattenIP4Block", fmt.Sprintf("type of entity passed was %s", *e.Type))
		return nil, d
	}

	i := &IP4BlockModel{}
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
				case "start":
					i.Start = types.StringValue(val)
				case "end":
					i.End = types.StringValue(val)
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

// IP4AddressModel describes the data model the built-in properties for an IP4Address object.
type IP4AddressModel struct {
	// These are exposed via the entity properties field for objects of type IP4Network
	Address               types.String
	State                 types.String
	MACAddress            types.String
	RouterPortInfo        types.String
	SwitchPortInfo        types.String
	VLANInfo              types.String
	LeaseTime             types.String
	ExpiryTime            types.String
	ParameterRequestList  types.String
	VendorClassIdentifier types.String
	LocationCode          types.String
	LocationInherited     types.Bool

	// these are user defined fields that are not built-in
	UserDefinedFields types.Map
}

func flattenIP4AddressProperties(e *gobam.APIEntity) (*IP4AddressModel, diag.Diagnostics) {
	var d diag.Diagnostics

	if e == nil {
		d.AddError("invalid input to flattenIP4Network", "entity passed was nil")
		return nil, d
	}
	if e.Type == nil {
		d.AddError("invalid input to flattenIP4Network", "type of entity passed was nil")
		return nil, d
	} else if *e.Type != "IP4Address" {
		d.AddError("invalid input to flattenIP4Address", fmt.Sprintf("type of entity passed was %s", *e.Type))
		return nil, d
	}

	i := &IP4AddressModel{}
	udfMap := make(map[string]attr.Value)

	if e.Properties != nil {
		props := strings.Split(*e.Properties, "|")
		for x := range props {
			if len(props[x]) > 0 {
				prop := strings.Split(props[x], "=")[0]
				val := strings.Split(props[x], "=")[1]

				switch prop {
				case "address":
					i.Address = types.StringValue(val)
				case "state":
					i.State = types.StringValue(val)
				case "macAddress":
					i.MACAddress = types.StringValue(val)
				case "routerPortInfo":
					i.RouterPortInfo = types.StringValue(val)
				case "switchPortInfo":
					i.SwitchPortInfo = types.StringValue(val)
				case "vlanInfo":
					i.VLANInfo = types.StringValue(val)
				case "leaseTime":
					i.LeaseTime = types.StringValue(val)
				case "expiryTime":
					i.ExpiryTime = types.StringValue(val)
				case "parameterRequestList":
					i.ParameterRequestList = types.StringValue(val)
				case "vendorClassIdentifier":
					i.VendorClassIdentifier = types.StringValue(val)
				case "locationCode":
					i.LocationCode = types.StringValue(val)
				case "locationInherited":
					b, err := strconv.ParseBool(val)
					if err != nil {
						d.AddError("error parsing locationInherited to bool", err.Error())
						break
					}
					i.LocationInherited = types.BoolValue(b)
				default:
					udfMap[prop] = types.StringValue(val)
				}
			}
		}
	}

	var userDefinedFields basetypes.MapValue
	userDefinedFields, udfDiag := basetypes.NewMapValue(types.StringType, udfMap)
	if udfDiag.HasError() {
		d.Append(udfDiag...)
	}
	i.UserDefinedFields = userDefinedFields
	return i, d
}

// HostRecordModel describes the data model the built-in properties for a Host Record object.
type HostRecordModel struct {
	// These are exposed via the entity properties field for objects of type IP4Network
	TTL           types.Int64
	AbsoluteName  types.String
	Addresses     types.Set
	ReverseRecord types.Bool

	// these are user defined fields that are not built-in
	UserDefinedFields types.Map

	// these are returned by the API but do not appear in the documentation
	AddressIDs types.Set

	// these are returned by the API with a hint based search but do not appear in the documentation
	ParentID   types.Int64
	ParentType types.String
}

func flattenHostRecordProperties(e *gobam.APIEntity) (*HostRecordModel, diag.Diagnostics) {
	var d diag.Diagnostics

	if e == nil {
		d.AddError("invalid input to flattenHostRecordProperties", "entity passed was nil")
		return nil, d
	}
	if e.Type == nil {
		d.AddError("invalid input to flattenHostRecordProperties", "type of entity passed was nil")
		return nil, d
	} else if *e.Type != "HostRecord" {
		d.AddError("invalid input to flattenHostRecordProperties", fmt.Sprintf("type of entity passed was %s", *e.Type))
		return nil, d
	}

	h := &HostRecordModel{}
	udfMap := make(map[string]attr.Value)

	addressesFound := false
	addressIDsFound := false
	var ttl int64 = -1
	var addressesSet basetypes.SetValue
	var addressIDsSet basetypes.SetValue

	if e.Properties != nil {
		props := strings.Split(*e.Properties, "|")
		for x := range props {
			if len(props[x]) > 0 {
				prop := strings.Split(props[x], "=")[0]
				val := strings.Split(props[x], "=")[1]

				switch prop {
				case "ttl":
					t, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						d.AddError("error parsing ttl to int64", err.Error())
						break
					}
					ttl = t
				case "absoluteName":
					h.AbsoluteName = types.StringValue(val)
				case "addresses":
					addressesFound = true
					var aDiag diag.Diagnostics
					addresses := strings.Split(val, ",")
					addressesList := []attr.Value{}
					for x := range addresses {
						addressesList = append(addressesList, types.StringValue(addresses[x]))
					}

					addressesSet, aDiag = basetypes.NewSetValue(types.StringType, addressesList)
					if aDiag.HasError() {
						d.Append(aDiag...)
						break
					}
				case "addressIds":
					addressIDsFound = true
					var aDiag diag.Diagnostics
					addressIDs := strings.Split(val, ",")
					addressIDsList := []attr.Value{}
					for x := range addressIDs {
						addressID, err := strconv.ParseInt(addressIDs[x], 10, 64)
						if err != nil {
							d.AddError("error parsing addressIds to int64", err.Error())
							break
						}
						addressIDsList = append(addressIDsList, types.Int64Value(addressID))
					}
					addressIDsSet, aDiag = basetypes.NewSetValue(types.Int64Type, addressIDsList)
					if aDiag.HasError() {
						d.Append(aDiag...)
						break
					}
				case "parentId":
					pid, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						d.AddError("error parsing parentId to int64", err.Error())
						break
					}
					h.ParentID = types.Int64Value(pid)
				case "parentType":
					h.ParentType = types.StringValue(val)
				case "reverseRecord":
					b, err := strconv.ParseBool(val)
					if err != nil {
						d.AddError("error parsing reverseRecord to bool", err.Error())
						break
					}
					h.ReverseRecord = types.BoolValue(b)
				default:
					udfMap[prop] = types.StringValue(val)
				}
			}
		}
	}

	if !addressesFound {
		addressesSet = basetypes.NewSetNull(types.StringType)
	}
	h.Addresses = addressesSet

	if !addressIDsFound {
		addressIDsSet = basetypes.NewSetNull(types.Int64Type)
	}
	h.AddressIDs = addressIDsSet

	h.TTL = types.Int64Value(ttl)

	var userDefinedFields basetypes.MapValue
	var udfDiag diag.Diagnostics
	userDefinedFields, udfDiag = basetypes.NewMapValue(types.StringType, udfMap)
	if udfDiag.HasError() {
		d.Append(udfDiag...)
	}
	h.UserDefinedFields = userDefinedFields

	return h, d
}
