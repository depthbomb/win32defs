// Package security provides Windows security constants and SID authorities.
package security

// SIDIdentifierAuthority identifies the issuing authority of a Windows SID.
type SIDIdentifierAuthority [6]byte

// SID_IDENTIFIER_AUTHORITY is the exact-name counterpart of the native structure.
type SID_IDENTIFIER_AUTHORITY = SIDIdentifierAuthority
