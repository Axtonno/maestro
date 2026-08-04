package runtime

type Capability string

const (
	CapabilityConfigure Capability = "configure"
	CapabilityInitialize Capability = "initialize"
	CapabilityStart Capability = "start"
	CapabilityStop Capability = "stop"
	CapabilityReload Capability = "reload"
	CapabilityHealth Capability = "health"
)