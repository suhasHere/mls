package mls

import (
	"fmt"
)

type CredentialType uint8

const (
	CredentialTypeBasic = 0
	CredentialTypeX509  = 1
)

// struct {
//     opaque identity<0..2^16-1>;
//     SignatureScheme algorithm;
//     SignaturePublicKey public_key;
// } BasicCredential;
type BasicCredential struct {
	Identity           []byte `tls:"head=2"`
	SignatureScheme    SignatureScheme
	SignaturePublicKey SignaturePublicKey
}

type X509Credential struct {
	CertData []byte `tls:"head=3"`
}

//	struct {
//		CredentialType credential_type;
//		select (Credential.credential_type) {
//			case basic:
//				BasicCredential;
//			case x509:
//				opaque cert_data<1..2^24-1>;
//		};
//} Credential;
type Credential struct {
	Basic *BasicCredential
	X509  *X509Credential
}

func (c Credential) Type() CredentialType {
	switch {
	case c.Basic != nil:
		return CredentialTypeBasic
	case c.X509 != nil:
		return CredentialTypeX509
	default:
		panic("Malformed credential")
	}
}

func (c Credential) MarshalTLS() ([]byte, error) {
	s := NewWriteStream()
	credentialType := c.Type()
	err := s.Write(credentialType)
	if err != nil {
		return nil, fmt.Errorf("mls.credential: Marshal failed for CredentialType")
	}
	switch credentialType {
	case CredentialTypeBasic:
		err = s.Write(c.Basic)
	case CredentialTypeX509:
		err = s.Write(c.X509)
	default:
		err = fmt.Errorf("mls.credential: CredentialType type not allowed")
	}

	if err != nil {
		return nil, fmt.Errorf("mls.credential: Marshal failed")
	}

	return s.Data(), nil
}

func (c *Credential) UnmarshalTLS(data []byte) (int, error) {
	s := NewReadStream(data)
	var credentialType CredentialType
	_, err := s.Read(&credentialType)
	if err != nil {
		return 0, fmt.Errorf("mls.credential: CredentialType Unmarshal failed %v", err)
	}

	switch credentialType {
	case CredentialTypeBasic:
		c.Basic = new(BasicCredential)
		_, err = s.Read(c.Basic)
	case CredentialTypeX509:
		c.X509 = new(X509Credential)
		_, err = s.Read(c.X509)
	default:
		err = fmt.Errorf("mls.credential: CredentialType type not allowed %v", err)
	}

	if err != nil {
		return 0, err
	}
	return s.Position(), nil
}