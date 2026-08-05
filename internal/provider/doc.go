// Package provider contains Maestro's in-process provider registry and router.
// It owns provider registration and default selection, and never holds its
// internal lock while invoking provider code.
package provider
