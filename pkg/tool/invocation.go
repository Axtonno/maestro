package tool

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
)

type Invocation struct {
	tool      ID
	call      CallID
	run       RunID
	arguments json.RawMessage
}

func NewInvocation(tool ID, call CallID, run RunID, arguments json.RawMessage) (Invocation, error) {
	if err := tool.Validate(); err != nil {
		return Invocation{}, fmt.Errorf("invocation tool: %w: %w", err, ErrInvalidInvocation)
	}
	if err := call.Validate(); err != nil {
		return Invocation{}, fmt.Errorf("invocation call: %w: %w", err, ErrInvalidInvocation)
	}
	if err := run.Validate(); err != nil {
		return Invocation{}, fmt.Errorf("invocation run: %w: %w", err, ErrInvalidInvocation)
	}
	normalized, err := normalizeJSONObject(arguments)
	if err != nil {
		return Invocation{}, fmt.Errorf("invocation arguments must be one JSON object: %w", ErrInvalidInvocation)
	}
	return Invocation{tool: tool, call: call, run: run, arguments: normalized}, nil
}

func (invocation Invocation) Tool() ID                   { return invocation.tool }
func (invocation Invocation) Call() CallID               { return invocation.call }
func (invocation Invocation) Run() RunID                 { return invocation.run }
func (invocation Invocation) Arguments() json.RawMessage { return bytes.Clone(invocation.arguments) }

func (invocation Invocation) Validate() error {
	_, err := NewInvocation(invocation.tool, invocation.call, invocation.run, invocation.arguments)
	return err
}

type PreparedInvocation struct {
	invocation  Invocation
	version     Version
	arguments   json.RawMessage
	actions     []Action
	fingerprint Fingerprint
}

func NewPreparedInvocation(
	invocation Invocation,
	version Version,
	normalizedArguments json.RawMessage,
	actions []Action,
) (PreparedInvocation, error) {
	if err := invocation.Validate(); err != nil {
		return PreparedInvocation{}, fmt.Errorf("prepared invocation source: %w: %w", err, ErrInvalidPreparedInvocation)
	}
	if err := version.Validate(); err != nil {
		return PreparedInvocation{}, fmt.Errorf("prepared invocation version: %w: %w", err, ErrInvalidPreparedInvocation)
	}
	normalized, err := normalizeJSONObject(normalizedArguments)
	if err != nil {
		return PreparedInvocation{}, fmt.Errorf("prepared arguments must be one JSON object: %w", ErrInvalidPreparedInvocation)
	}
	clonedActions := slices.Clone(actions)
	if len(clonedActions) == 0 {
		return PreparedInvocation{}, fmt.Errorf("prepared invocation requires at least one action: %w", ErrInvalidPreparedInvocation)
	}
	for index, action := range clonedActions {
		if err := action.Validate(); err != nil {
			return PreparedInvocation{}, fmt.Errorf("prepared action %d: %w: %w", index, err, ErrInvalidPreparedInvocation)
		}
		if !action.effect.ValidForTool() {
			return PreparedInvocation{}, fmt.Errorf("prepared action %d effect %q is reserved for model permissions: %w", index, action.effect, ErrInvalidPreparedInvocation)
		}
	}
	fingerprint := fingerprintPrepared(invocation, version, normalized, clonedActions)
	return PreparedInvocation{
		invocation: invocation, version: version, arguments: normalized,
		actions: clonedActions, fingerprint: fingerprint,
	}, nil
}

func (prepared PreparedInvocation) Invocation() Invocation { return prepared.invocation }
func (prepared PreparedInvocation) Version() Version       { return prepared.version }
func (prepared PreparedInvocation) Arguments() json.RawMessage {
	return bytes.Clone(prepared.arguments)
}
func (prepared PreparedInvocation) Actions() []Action        { return slices.Clone(prepared.actions) }
func (prepared PreparedInvocation) Fingerprint() Fingerprint { return prepared.fingerprint }

func (prepared PreparedInvocation) Validate() error {
	rebuilt, err := NewPreparedInvocation(
		prepared.invocation,
		prepared.version,
		prepared.arguments,
		prepared.actions,
	)
	if err != nil {
		return err
	}
	if prepared.fingerprint != rebuilt.fingerprint {
		return fmt.Errorf("prepared invocation fingerprint does not match its contents: %w", ErrInvalidPreparedInvocation)
	}
	return nil
}

func fingerprintPrepared(invocation Invocation, version Version, arguments json.RawMessage, actions []Action) Fingerprint {
	hash := sha256.New()
	writeFingerprintPart(hash, string(invocation.tool))
	writeFingerprintPart(hash, string(version))
	writeFingerprintPart(hash, string(invocation.call))
	writeFingerprintPart(hash, string(invocation.run))
	writeFingerprintPart(hash, string(arguments))
	for _, action := range actions {
		writeFingerprintPart(hash, string(action.effect))
		writeFingerprintPart(hash, action.resource)
		writeFingerprintPart(hash, string(action.workspace))
	}
	return Fingerprint(hex.EncodeToString(hash.Sum(nil)))
}

func fingerprintToolPermission(policy PolicyID, prepared PreparedInvocation) Fingerprint {
	hash := sha256.New()
	writeFingerprintPart(hash, string(policy))
	writeFingerprintPart(hash, string(PermissionSubjectTool))
	writeFingerprintPart(hash, string(prepared.fingerprint))
	return Fingerprint(hex.EncodeToString(hash.Sum(nil)))
}

type fingerprintWriter interface {
	Write([]byte) (int, error)
}

func writeFingerprintPart(writer fingerprintWriter, value string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = writer.Write(size[:])
	_, _ = writer.Write([]byte(value))
}
