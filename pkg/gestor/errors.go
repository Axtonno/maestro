package gestor

import "errors"

var (
	ErrInvalidCapabilityID     = errors.New("invalid capability ID")
	ErrInvalidSourceID         = errors.New("invalid source ID")
	ErrInvalidTarget           = errors.New("invalid capability target")
	ErrInvalidAvailability     = errors.New("invalid capability availability")
	ErrInvalidDescriptor       = errors.New("invalid capability descriptor")
	ErrInvalidQuery            = errors.New("invalid capability query")
	ErrInvalidSnapshot         = errors.New("invalid capability snapshot")
	ErrInvalidResolution       = errors.New("invalid capability resolution")
	ErrInvalidSource           = errors.New("invalid capability source")
	ErrSourceAlreadyRegistered = errors.New("capability source already registered")
	ErrSourceFailure           = errors.New("capability source failure")
	ErrNotFound                = errors.New("capability not found")
	ErrUnavailable             = errors.New("capability unavailable")
	ErrAmbiguous               = errors.New("ambiguous capability resolution")
	ErrStaleSnapshot           = errors.New("stale capability snapshot")
)
