package provider

import (
	"testing"
)

func TestParseIP4NetworkProperties_BasicProperties(t *testing.T) {
	input := "CIDR=10.0.0.0/24|gateway=10.0.0.1|name=TestNetwork|allowDuplicateHost=enable|pingBeforeAssign=disable"

	props, diags := parseIP4NetworkProperties(input)
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	if got := props.cidr.ValueString(); got != "10.0.0.0/24" {
		t.Errorf("cidr = %q, want %q", got, "10.0.0.0/24")
	}
	if got := props.gateway.ValueString(); got != "10.0.0.1" {
		t.Errorf("gateway = %q, want %q", got, "10.0.0.1")
	}
	if got := props.name.ValueString(); got != "TestNetwork" {
		t.Errorf("name = %q, want %q", got, "TestNetwork")
	}
	if got := props.allowDuplicateHost.ValueString(); got != "enable" {
		t.Errorf("allowDuplicateHost = %q, want %q", got, "enable")
	}
	if got := props.pingBeforeAssign.ValueString(); got != "disable" {
		t.Errorf("pingBeforeAssign = %q, want %q", got, "disable")
	}
}

func TestParseIP4NetworkProperties_ValuesContainingEquals(t *testing.T) {
	input := "CIDR=10.0.0.0/24|name=foo=bar=baz"

	props, diags := parseIP4NetworkProperties(input)
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	if got := props.name.ValueString(); got != "foo=bar=baz" {
		t.Errorf("name = %q, want %q", got, "foo=bar=baz")
	}
}

func TestParseIP4NetworkProperties_MalformedSegments(t *testing.T) {
	input := "CIDR=10.0.0.0/24|malformed_no_equals|gateway=10.0.0.1"

	props, diags := parseIP4NetworkProperties(input)
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	if got := props.cidr.ValueString(); got != "10.0.0.0/24" {
		t.Errorf("cidr = %q, want %q", got, "10.0.0.0/24")
	}
	if got := props.gateway.ValueString(); got != "10.0.0.1" {
		t.Errorf("gateway = %q, want %q", got, "10.0.0.1")
	}
}

func TestParseIP4NetworkProperties_EmptySegments(t *testing.T) {
	input := "|CIDR=10.0.0.0/24||gateway=10.0.0.1|"

	props, diags := parseIP4NetworkProperties(input)
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	if got := props.cidr.ValueString(); got != "10.0.0.0/24" {
		t.Errorf("cidr = %q, want %q", got, "10.0.0.0/24")
	}
	if got := props.gateway.ValueString(); got != "10.0.0.1" {
		t.Errorf("gateway = %q, want %q", got, "10.0.0.1")
	}
}

func TestParseIP4NetworkProperties_BooleanFields(t *testing.T) {
	input := "inheritAllowDuplicateHost=true|inheritPingBeforeAssign=false|inheritDNSRestrictions=true|inheritDefaultDomains=false|inheritDefaultView=true|locationInherited=false"

	props, diags := parseIP4NetworkProperties(input)
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{"inheritAllowDuplicateHost", props.inheritAllowDuplicateHost.ValueBool(), true},
		{"inheritPingBeforeAssign", props.inheritPingBeforeAssign.ValueBool(), false},
		{"inheritDNSRestrictions", props.inheritDNSRestrictions.ValueBool(), true},
		{"inheritDefaultDomains", props.inheritDefaultDomains.ValueBool(), false},
		{"inheritDefaultView", props.inheritDefaultView.ValueBool(), true},
		{"locationInherited", props.locationInherited.ValueBool(), false},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

func TestParseIP4NetworkProperties_InvalidBooleanReturnsError(t *testing.T) {
	input := "inheritAllowDuplicateHost=notabool"

	_, diags := parseIP4NetworkProperties(input)
	if !diags.HasError() {
		t.Fatal("expected error for invalid boolean, got none")
	}
}

func TestParseIP4NetworkProperties_NumericFields(t *testing.T) {
	input := "template=12345|defaultView=67890"

	props, diags := parseIP4NetworkProperties(input)
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	if got := props.template.ValueInt64(); got != 12345 {
		t.Errorf("template = %d, want %d", got, 12345)
	}
	if got := props.defaultView.ValueInt64(); got != 67890 {
		t.Errorf("defaultView = %d, want %d", got, 67890)
	}
}

func TestParseIP4NetworkProperties_InvalidNumericReturnsError(t *testing.T) {
	input := "template=notanumber"

	_, diags := parseIP4NetworkProperties(input)
	if !diags.HasError() {
		t.Fatal("expected error for invalid number, got none")
	}
}

func TestParseIP4NetworkProperties_CustomProperties(t *testing.T) {
	input := "CIDR=10.0.0.0/24|myCustomField=customValue|anotherField=anotherValue"

	props, diags := parseIP4NetworkProperties(input)
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	cpMap := props.customProperties.Elements()
	if len(cpMap) != 2 {
		t.Fatalf("customProperties has %d entries, want 2", len(cpMap))
	}
}

func TestParseIP4NetworkProperties_DefaultDomainsAndDNSRestrictions(t *testing.T) {
	input := "defaultDomains=100,200,300|dnsRestrictions=400,500"

	props, diags := parseIP4NetworkProperties(input)
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	if got := len(props.defaultDomains.Elements()); got != 3 {
		t.Errorf("defaultDomains has %d elements, want 3", got)
	}

	if got := len(props.dnsRestrictions.Elements()); got != 2 {
		t.Errorf("dnsRestrictions has %d elements, want 2", got)
	}
}

func TestParseIP4NetworkProperties_EmptyString(t *testing.T) {
	props, diags := parseIP4NetworkProperties("")
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	if !props.cidr.IsNull() {
		t.Errorf("cidr should be null for empty input, got %q", props.cidr.ValueString())
	}
}
