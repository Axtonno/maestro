package tool

import (
	"context"
	"fmt"
	"reflect"
)

// Tool is the trusted in-process extension SPI. Agent code receives Runtime,
// not Tool instances; Runtime.Invoke is the public authorized execution path.
type Tool interface {
	Descriptor() Descriptor
	Prepare(context.Context, Invocation) (PreparedInvocation, error)
	Execute(context.Context, PreparedInvocation) (Result, error)
}

func ValidateTool(candidate Tool) error {
	if candidate == nil || nilInterface(candidate) {
		return ErrInvalidTool
	}
	if err := candidate.Descriptor().Validate(); err != nil {
		return fmt.Errorf("tool descriptor: %w: %w", err, ErrInvalidTool)
	}
	return nil
}

type ExecutionRequest struct {
	invocation Invocation
	policy     PolicyID
	approver   Approver
	limits     ExecutionLimits
}

func NewExecutionRequest(
	invocation Invocation,
	policy PolicyID,
	approver Approver,
	limits ExecutionLimits,
) (ExecutionRequest, error) {
	if err := invocation.Validate(); err != nil {
		return ExecutionRequest{}, fmt.Errorf("execution invocation: %w: %w", err, ErrInvalidExecutionRequest)
	}
	if err := policy.Validate(); err != nil {
		return ExecutionRequest{}, fmt.Errorf("execution policy: %w: %w", err, ErrInvalidExecutionRequest)
	}
	if approver != nil && nilInterface(approver) {
		return ExecutionRequest{}, fmt.Errorf("execution approver is typed nil: %w: %w", ErrInvalidApprover, ErrInvalidExecutionRequest)
	}
	if err := limits.Validate(); err != nil {
		return ExecutionRequest{}, fmt.Errorf("execution limits: %w: %w", err, ErrInvalidExecutionRequest)
	}
	return ExecutionRequest{invocation: invocation, policy: policy, approver: approver, limits: limits}, nil
}

func (request ExecutionRequest) Invocation() Invocation  { return request.invocation }
func (request ExecutionRequest) Policy() PolicyID        { return request.policy }
func (request ExecutionRequest) Approver() Approver      { return request.approver }
func (request ExecutionRequest) Limits() ExecutionLimits { return request.limits }

func (request ExecutionRequest) Validate() error {
	_, err := NewExecutionRequest(request.invocation, request.policy, request.approver, request.limits)
	return err
}

type Catalog interface {
	Register(Tool) error
	Descriptors() []Descriptor
}

// Runtime deliberately exposes no method that accepts a Decision. Invoke
// performs policy resolution, authorization, permit issuance/consumption, and
// execution as one operation.
type Runtime interface {
	Catalog
	RegisterPolicy(Policy) error
	Policies() []PolicyID
	Invoke(context.Context, ExecutionRequest) (Result, error)
	AuthorizeModel(context.Context, PermissionRequest, Approver) (Decision, error)
}

func ValidatePolicy(policy Policy) error {
	if policy == nil || nilInterface(policy) {
		return ErrInvalidPolicy
	}
	if err := policy.ID().Validate(); err != nil {
		return fmt.Errorf("policy identity: %w: %w", err, ErrInvalidPolicy)
	}
	return nil
}

func nilInterface(value any) bool {
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
