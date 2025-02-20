// policy_config.go hanles creation of policy objects, either by parsing JSON
// or by programs building them programmatically.

// The New* constructors are intended to be a stable API. FIXME: after an independent review.

// Do not invoke the internals of the JSON marshaling/unmarshaling directly.

// We can't just blindly call json.Unmarshal because that would silently ignore
// typos, and that would just not do for security policy.

// FIXME? This is by no means an user-friendly parser: No location information in error messages, no other context.
// But at least it is not worse than blind json.Unmarshal()…

package signature

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/nholuongut/image/docker/reference"
	"github.com/nholuongut/image/transports"
	"github.com/nholuongut/image/types"
	"github.com/pkg/errors"
)

// systemDefaultPolicyPath is the policy path used for DefaultPolicy().
// You can override this at build time with
// -ldflags '-X github.com/nholuongut/image/signature.systemDefaultPolicyPath=$your_path'
var systemDefaultPolicyPath = builtinDefaultPolicyPath

// builtinDefaultPolicyPath is the policy path used for DefaultPolicy().
// DO NOT change this, instead see systemDefaultPolicyPath above.
const builtinDefaultPolicyPath = "/etc/nholuongut/policy.json"

// InvalidPolicyFormatError is returned when parsing an invalid policy configuration.
type InvalidPolicyFormatError string

func (err InvalidPolicyFormatError) Error() string {
	return string(err)
}

// DefaultPolicy returns the default policy of the system.
// Most applications should be using this method to get the policy configured
// by the system administrator.
// sys should usually be nil, can be set to override the default.
// NOTE: When this function returns an error, report it to the user and abort.
// DO NOT hard-code fallback policies in your application.
func DefaultPolicy(sys *types.SystemContext) (*Policy, error) {
	return NewPolicyFromFile(defaultPolicyPath(sys))
}

// defaultPolicyPath returns a path to the default policy of the system.
func defaultPolicyPath(sys *types.SystemContext) string {
	if sys != nil {
		if sys.SignaturePolicyPath != "" {
			return sys.SignaturePolicyPath
		}
		if sys.RootForImplicitAbsolutePaths != "" {
			return filepath.Join(sys.RootForImplicitAbsolutePaths, systemDefaultPolicyPath)
		}
	}
	return systemDefaultPolicyPath
}

// NewPolicyFromFile returns a policy configured in the specified file.
func NewPolicyFromFile(fileName string) (*Policy, error) {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	policy, err := NewPolicyFromBytes(contents)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid policy in %q", fileName)
	}
	return policy, nil
}

// NewPolicyFromBytes returns a policy parsed from the specified blob.
// Use this function instead of calling json.Unmarshal directly.
func NewPolicyFromBytes(data []byte) (*Policy, error) {
	p := Policy{}
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, InvalidPolicyFormatError(err.Error())
	}
	return &p, nil
}

// Compile-time check that Policy implements json.Unmarshaler.
var _ json.Unmarshaler = (*Policy)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *Policy) UnmarshalJSON(data []byte) error {
	*p = Policy{}
	transports := policyTransportsMap{}
	if err := paranoidUnmarshalJSONObject(data, func(key string) interface{} {
		switch key {
		case "default":
			return &p.Default
		case "transports":
			return &transports
		default:
			return nil
		}
	}); err != nil {
		return err
	}

	if p.Default == nil {
		return InvalidPolicyFormatError("Default policy is missing")
	}
	p.Transports = map[string]PolicyTransportScopes(transports)
	return nil
}

// policyTransportsMap is a specialization of this map type for the strict JSON parsing semantics appropriate for the Policy.Transports member.
type policyTransportsMap map[string]PolicyTransportScopes

// Compile-time check that policyTransportsMap implements json.Unmarshaler.
var _ json.Unmarshaler = (*policyTransportsMap)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *policyTransportsMap) UnmarshalJSON(data []byte) error {
	// We can't unmarshal directly into map values because it is not possible to take an address of a map value.
	// So, use a temporary map of pointers-to-slices and convert.
	tmpMap := map[string]*PolicyTransportScopes{}
	if err := paranoidUnmarshalJSONObject(data, func(key string) interface{} {
		// transport can be nil
		transport := transports.Get(key)
		// paranoidUnmarshalJSONObject detects key duplication for us, check just to be safe.
		if _, ok := tmpMap[key]; ok {
			return nil
		}
		ptsWithTransport := policyTransportScopesWithTransport{
			transport: transport,
			dest:      &PolicyTransportScopes{}, // This allocates a new instance on each call.
		}
		tmpMap[key] = ptsWithTransport.dest
		return &ptsWithTransport
	}); err != nil {
		return err
	}
	for key, ptr := range tmpMap {
		(*m)[key] = *ptr
	}
	return nil
}

// Compile-time check that PolicyTransportScopes "implements"" json.Unmarshaler.
// we want to only use policyTransportScopesWithTransport
var _ json.Unmarshaler = (*PolicyTransportScopes)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *PolicyTransportScopes) UnmarshalJSON(data []byte) error {
	return errors.New("Do not try to unmarshal PolicyTransportScopes directly")
}

// policyTransportScopesWithTransport is a way to unmarshal a PolicyTransportScopes
// while validating using a specific ImageTransport if not nil.
type policyTransportScopesWithTransport struct {
	transport types.ImageTransport
	dest      *PolicyTransportScopes
}

// Compile-time check that policyTransportScopesWithTransport implements json.Unmarshaler.
var _ json.Unmarshaler = (*policyTransportScopesWithTransport)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *policyTransportScopesWithTransport) UnmarshalJSON(data []byte) error {
	// We can't unmarshal directly into map values because it is not possible to take an address of a map value.
	// So, use a temporary map of pointers-to-slices and convert.
	tmpMap := map[string]*PolicyRequirements{}
	if err := paranoidUnmarshalJSONObject(data, func(key string) interface{} {
		// paranoidUnmarshalJSONObject detects key duplication for us, check just to be safe.
		if _, ok := tmpMap[key]; ok {
			return nil
		}
		if key != "" && m.transport != nil {
			if err := m.transport.ValidatePolicyConfigurationScope(key); err != nil {
				return nil
			}
		}
		ptr := &PolicyRequirements{} // This allocates a new instance on each call.
		tmpMap[key] = ptr
		return ptr
	}); err != nil {
		return err
	}
	for key, ptr := range tmpMap {
		(*m.dest)[key] = *ptr
	}
	return nil
}

// Compile-time check that PolicyRequirements implements json.Unmarshaler.
var _ json.Unmarshaler = (*PolicyRequirements)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *PolicyRequirements) UnmarshalJSON(data []byte) error {
	reqJSONs := []json.RawMessage{}
	if err := json.Unmarshal(data, &reqJSONs); err != nil {
		return err
	}
	if len(reqJSONs) == 0 {
		return InvalidPolicyFormatError("List of verification policy requirements must not be empty")
	}
	res := make([]PolicyRequirement, len(reqJSONs))
	for i, reqJSON := range reqJSONs {
		req, err := newPolicyRequirementFromJSON(reqJSON)
		if err != nil {
			return err
		}
		res[i] = req
	}
	*m = res
	return nil
}

// newPolicyRequirementFromJSON parses JSON data into a PolicyRequirement implementation.
func newPolicyRequirementFromJSON(data []byte) (PolicyRequirement, error) {
	var typeField prCommon
	if err := json.Unmarshal(data, &typeField); err != nil {
		return nil, err
	}
	var res PolicyRequirement
	switch typeField.Type {
	case prTypeInsecureAcceptAnything:
		res = &prInsecureAcceptAnything{}
	case prTypeReject:
		res = &prReject{}
	case prTypeSignedBy:
		res = &prSignedBy{}
	case prTypeSignedBaseLayer:
		res = &prSignedBaseLayer{}
	default:
		return nil, InvalidPolicyFormatError(fmt.Sprintf("Unknown policy requirement type \"%s\"", typeField.Type))
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// newPRInsecureAcceptAnything is NewPRInsecureAcceptAnything, except it returns the private type.
func newPRInsecureAcceptAnything() *prInsecureAcceptAnything {
	return &prInsecureAcceptAnything{prCommon{Type: prTypeInsecureAcceptAnything}}
}

// NewPRInsecureAcceptAnything returns a new "insecureAcceptAnything" PolicyRequirement.
func NewPRInsecureAcceptAnything() PolicyRequirement {
	return newPRInsecureAcceptAnything()
}

// Compile-time check that prInsecureAcceptAnything implements json.Unmarshaler.
var _ json.Unmarshaler = (*prInsecureAcceptAnything)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (pr *prInsecureAcceptAnything) UnmarshalJSON(data []byte) error {
	*pr = prInsecureAcceptAnything{}
	var tmp prInsecureAcceptAnything
	if err := paranoidUnmarshalJSONObjectExactFields(data, map[string]interface{}{
		"type": &tmp.Type,
	}); err != nil {
		return err
	}

	if tmp.Type != prTypeInsecureAcceptAnything {
		return InvalidPolicyFormatError(fmt.Sprintf("Unexpected policy requirement type \"%s\"", tmp.Type))
	}
	*pr = *newPRInsecureAcceptAnything()
	return nil
}

// newPRReject is NewPRReject, except it returns the private type.
func newPRReject() *prReject {
	return &prReject{prCommon{Type: prTypeReject}}
}

// NewPRReject returns a new "reject" PolicyRequirement.
func NewPRReject() PolicyRequirement {
	return newPRReject()
}

// Compile-time check that prReject implements json.Unmarshaler.
var _ json.Unmarshaler = (*prReject)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (pr *prReject) UnmarshalJSON(data []byte) error {
	*pr = prReject{}
	var tmp prReject
	if err := paranoidUnmarshalJSONObjectExactFields(data, map[string]interface{}{
		"type": &tmp.Type,
	}); err != nil {
		return err
	}

	if tmp.Type != prTypeReject {
		return InvalidPolicyFormatError(fmt.Sprintf("Unexpected policy requirement type \"%s\"", tmp.Type))
	}
	*pr = *newPRReject()
	return nil
}

// newPRSignedBy returns a new prSignedBy if parameters are valid.
func newPRSignedBy(keyType sbKeyType, keyPath string, keyData []byte, signedIdentity PolicyReferenceMatch) (*prSignedBy, error) {
	if !keyType.IsValid() {
		return nil, InvalidPolicyFormatError(fmt.Sprintf("invalid keyType \"%s\"", keyType))
	}
	if len(keyPath) > 0 && len(keyData) > 0 {
		return nil, InvalidPolicyFormatError("keyType and keyData cannot be used simultaneously")
	}
	if signedIdentity == nil {
		return nil, InvalidPolicyFormatError("signedIdentity not specified")
	}
	return &prSignedBy{
		prCommon:       prCommon{Type: prTypeSignedBy},
		KeyType:        keyType,
		KeyPath:        keyPath,
		KeyData:        keyData,
		SignedIdentity: signedIdentity,
	}, nil
}

// newPRSignedByKeyPath is NewPRSignedByKeyPath, except it returns the private type.
func newPRSignedByKeyPath(keyType sbKeyType, keyPath string, signedIdentity PolicyReferenceMatch) (*prSignedBy, error) {
	return newPRSignedBy(keyType, keyPath, nil, signedIdentity)
}

// NewPRSignedByKeyPath returns a new "signedBy" PolicyRequirement using a KeyPath
func NewPRSignedByKeyPath(keyType sbKeyType, keyPath string, signedIdentity PolicyReferenceMatch) (PolicyRequirement, error) {
	return newPRSignedByKeyPath(keyType, keyPath, signedIdentity)
}

// newPRSignedByKeyData is NewPRSignedByKeyData, except it returns the private type.
func newPRSignedByKeyData(keyType sbKeyType, keyData []byte, signedIdentity PolicyReferenceMatch) (*prSignedBy, error) {
	return newPRSignedBy(keyType, "", keyData, signedIdentity)
}

// NewPRSignedByKeyData returns a new "signedBy" PolicyRequirement using a KeyData
func NewPRSignedByKeyData(keyType sbKeyType, keyData []byte, signedIdentity PolicyReferenceMatch) (PolicyRequirement, error) {
	return newPRSignedByKeyData(keyType, keyData, signedIdentity)
}

// Compile-time check that prSignedBy implements json.Unmarshaler.
var _ json.Unmarshaler = (*prSignedBy)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (pr *prSignedBy) UnmarshalJSON(data []byte) error {
	*pr = prSignedBy{}
	var tmp prSignedBy
	var gotKeyPath, gotKeyData = false, false
	var signedIdentity json.RawMessage
	if err := paranoidUnmarshalJSONObject(data, func(key string) interface{} {
		switch key {
		case "type":
			return &tmp.Type
		case "keyType":
			return &tmp.KeyType
		case "keyPath":
			gotKeyPath = true
			return &tmp.KeyPath
		case "keyData":
			gotKeyData = true
			return &tmp.KeyData
		case "signedIdentity":
			return &signedIdentity
		default:
			return nil
		}
	}); err != nil {
		return err
	}

	if tmp.Type != prTypeSignedBy {
		return InvalidPolicyFormatError(fmt.Sprintf("Unexpected policy requirement type \"%s\"", tmp.Type))
	}
	if signedIdentity == nil {
		tmp.SignedIdentity = NewPRMMatchRepoDigestOrExact()
	} else {
		si, err := newPolicyReferenceMatchFromJSON(signedIdentity)
		if err != nil {
			return err
		}
		tmp.SignedIdentity = si
	}

	var res *prSignedBy
	var err error
	switch {
	case gotKeyPath && gotKeyData:
		return InvalidPolicyFormatError("keyPath and keyData cannot be used simultaneously")
	case gotKeyPath && !gotKeyData:
		res, err = newPRSignedByKeyPath(tmp.KeyType, tmp.KeyPath, tmp.SignedIdentity)
	case !gotKeyPath && gotKeyData:
		res, err = newPRSignedByKeyData(tmp.KeyType, tmp.KeyData, tmp.SignedIdentity)
	case !gotKeyPath && !gotKeyData:
		return InvalidPolicyFormatError("At least one of keyPath and keyData mus be specified")
	default: // Coverage: This should never happen
		return errors.Errorf("Impossible keyPath/keyData presence combination!?")
	}
	if err != nil {
		return err
	}
	*pr = *res

	return nil
}

// IsValid returns true iff kt is a recognized value
func (kt sbKeyType) IsValid() bool {
	switch kt {
	case SBKeyTypeGPGKeys, SBKeyTypeSignedByGPGKeys,
		SBKeyTypeX509Certificates, SBKeyTypeSignedByX509CAs:
		return true
	default:
		return false
	}
}

// Compile-time check that sbKeyType implements json.Unmarshaler.
var _ json.Unmarshaler = (*sbKeyType)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (kt *sbKeyType) UnmarshalJSON(data []byte) error {
	*kt = sbKeyType("")
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if !sbKeyType(s).IsValid() {
		return InvalidPolicyFormatError(fmt.Sprintf("Unrecognized keyType value \"%s\"", s))
	}
	*kt = sbKeyType(s)
	return nil
}

// newPRSignedBaseLayer is NewPRSignedBaseLayer, except it returns the private type.
func newPRSignedBaseLayer(baseLayerIdentity PolicyReferenceMatch) (*prSignedBaseLayer, error) {
	if baseLayerIdentity == nil {
		return nil, InvalidPolicyFormatError("baseLayerIdentity not specified")
	}
	return &prSignedBaseLayer{
		prCommon:          prCommon{Type: prTypeSignedBaseLayer},
		BaseLayerIdentity: baseLayerIdentity,
	}, nil
}

// NewPRSignedBaseLayer returns a new "signedBaseLayer" PolicyRequirement.
func NewPRSignedBaseLayer(baseLayerIdentity PolicyReferenceMatch) (PolicyRequirement, error) {
	return newPRSignedBaseLayer(baseLayerIdentity)
}

// Compile-time check that prSignedBaseLayer implements json.Unmarshaler.
var _ json.Unmarshaler = (*prSignedBaseLayer)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (pr *prSignedBaseLayer) UnmarshalJSON(data []byte) error {
	*pr = prSignedBaseLayer{}
	var tmp prSignedBaseLayer
	var baseLayerIdentity json.RawMessage
	if err := paranoidUnmarshalJSONObjectExactFields(data, map[string]interface{}{
		"type":              &tmp.Type,
		"baseLayerIdentity": &baseLayerIdentity,
	}); err != nil {
		return err
	}

	if tmp.Type != prTypeSignedBaseLayer {
		return InvalidPolicyFormatError(fmt.Sprintf("Unexpected policy requirement type \"%s\"", tmp.Type))
	}
	bli, err := newPolicyReferenceMatchFromJSON(baseLayerIdentity)
	if err != nil {
		return err
	}
	res, err := newPRSignedBaseLayer(bli)
	if err != nil {
		// Coverage: This should never happen, newPolicyReferenceMatchFromJSON has ensured bli is valid.
		return err
	}
	*pr = *res
	return nil
}

// newPolicyReferenceMatchFromJSON parses JSON data into a PolicyReferenceMatch implementation.
func newPolicyReferenceMatchFromJSON(data []byte) (PolicyReferenceMatch, error) {
	var typeField prmCommon
	if err := json.Unmarshal(data, &typeField); err != nil {
		return nil, err
	}
	var res PolicyReferenceMatch
	switch typeField.Type {
	case prmTypeMatchExact:
		res = &prmMatchExact{}
	case prmTypeMatchRepoDigestOrExact:
		res = &prmMatchRepoDigestOrExact{}
	case prmTypeMatchRepository:
		res = &prmMatchRepository{}
	case prmTypeExactReference:
		res = &prmExactReference{}
	case prmTypeExactRepository:
		res = &prmExactRepository{}
	default:
		return nil, InvalidPolicyFormatError(fmt.Sprintf("Unknown policy reference match type \"%s\"", typeField.Type))
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// newPRMMatchExact is NewPRMMatchExact, except it resturns the private type.
func newPRMMatchExact() *prmMatchExact {
	return &prmMatchExact{prmCommon{Type: prmTypeMatchExact}}
}

// NewPRMMatchExact returns a new "matchExact" PolicyReferenceMatch.
func NewPRMMatchExact() PolicyReferenceMatch {
	return newPRMMatchExact()
}

// Compile-time check that prmMatchExact implements json.Unmarshaler.
var _ json.Unmarshaler = (*prmMatchExact)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (prm *prmMatchExact) UnmarshalJSON(data []byte) error {
	*prm = prmMatchExact{}
	var tmp prmMatchExact
	if err := paranoidUnmarshalJSONObjectExactFields(data, map[string]interface{}{
		"type": &tmp.Type,
	}); err != nil {
		return err
	}

	if tmp.Type != prmTypeMatchExact {
		return InvalidPolicyFormatError(fmt.Sprintf("Unexpected policy requirement type \"%s\"", tmp.Type))
	}
	*prm = *newPRMMatchExact()
	return nil
}

// newPRMMatchRepoDigestOrExact is NewPRMMatchRepoDigestOrExact, except it resturns the private type.
func newPRMMatchRepoDigestOrExact() *prmMatchRepoDigestOrExact {
	return &prmMatchRepoDigestOrExact{prmCommon{Type: prmTypeMatchRepoDigestOrExact}}
}

// NewPRMMatchRepoDigestOrExact returns a new "matchRepoDigestOrExact" PolicyReferenceMatch.
func NewPRMMatchRepoDigestOrExact() PolicyReferenceMatch {
	return newPRMMatchRepoDigestOrExact()
}

// Compile-time check that prmMatchRepoDigestOrExact implements json.Unmarshaler.
var _ json.Unmarshaler = (*prmMatchRepoDigestOrExact)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (prm *prmMatchRepoDigestOrExact) UnmarshalJSON(data []byte) error {
	*prm = prmMatchRepoDigestOrExact{}
	var tmp prmMatchRepoDigestOrExact
	if err := paranoidUnmarshalJSONObjectExactFields(data, map[string]interface{}{
		"type": &tmp.Type,
	}); err != nil {
		return err
	}

	if tmp.Type != prmTypeMatchRepoDigestOrExact {
		return InvalidPolicyFormatError(fmt.Sprintf("Unexpected policy requirement type \"%s\"", tmp.Type))
	}
	*prm = *newPRMMatchRepoDigestOrExact()
	return nil
}

// newPRMMatchRepository is NewPRMMatchRepository, except it resturns the private type.
func newPRMMatchRepository() *prmMatchRepository {
	return &prmMatchRepository{prmCommon{Type: prmTypeMatchRepository}}
}

// NewPRMMatchRepository returns a new "matchRepository" PolicyReferenceMatch.
func NewPRMMatchRepository() PolicyReferenceMatch {
	return newPRMMatchRepository()
}

// Compile-time check that prmMatchRepository implements json.Unmarshaler.
var _ json.Unmarshaler = (*prmMatchRepository)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (prm *prmMatchRepository) UnmarshalJSON(data []byte) error {
	*prm = prmMatchRepository{}
	var tmp prmMatchRepository
	if err := paranoidUnmarshalJSONObjectExactFields(data, map[string]interface{}{
		"type": &tmp.Type,
	}); err != nil {
		return err
	}

	if tmp.Type != prmTypeMatchRepository {
		return InvalidPolicyFormatError(fmt.Sprintf("Unexpected policy requirement type \"%s\"", tmp.Type))
	}
	*prm = *newPRMMatchRepository()
	return nil
}

// newPRMExactReference is NewPRMExactReference, except it resturns the private type.
func newPRMExactReference(dockerReference string) (*prmExactReference, error) {
	ref, err := reference.ParseNormalizedNamed(dockerReference)
	if err != nil {
		return nil, InvalidPolicyFormatError(fmt.Sprintf("Invalid format of dockerReference %s: %s", dockerReference, err.Error()))
	}
	if reference.IsNameOnly(ref) {
		return nil, InvalidPolicyFormatError(fmt.Sprintf("dockerReference %s contains neither a tag nor digest", dockerReference))
	}
	return &prmExactReference{
		prmCommon:       prmCommon{Type: prmTypeExactReference},
		DockerReference: dockerReference,
	}, nil
}

// NewPRMExactReference returns a new "exactReference" PolicyReferenceMatch.
func NewPRMExactReference(dockerReference string) (PolicyReferenceMatch, error) {
	return newPRMExactReference(dockerReference)
}

// Compile-time check that prmExactReference implements json.Unmarshaler.
var _ json.Unmarshaler = (*prmExactReference)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (prm *prmExactReference) UnmarshalJSON(data []byte) error {
	*prm = prmExactReference{}
	var tmp prmExactReference
	if err := paranoidUnmarshalJSONObjectExactFields(data, map[string]interface{}{
		"type":            &tmp.Type,
		"dockerReference": &tmp.DockerReference,
	}); err != nil {
		return err
	}

	if tmp.Type != prmTypeExactReference {
		return InvalidPolicyFormatError(fmt.Sprintf("Unexpected policy requirement type \"%s\"", tmp.Type))
	}

	res, err := newPRMExactReference(tmp.DockerReference)
	if err != nil {
		return err
	}
	*prm = *res
	return nil
}

// newPRMExactRepository is NewPRMExactRepository, except it resturns the private type.
func newPRMExactRepository(dockerRepository string) (*prmExactRepository, error) {
	if _, err := reference.ParseNormalizedNamed(dockerRepository); err != nil {
		return nil, InvalidPolicyFormatError(fmt.Sprintf("Invalid format of dockerRepository %s: %s", dockerRepository, err.Error()))
	}
	return &prmExactRepository{
		prmCommon:        prmCommon{Type: prmTypeExactRepository},
		DockerRepository: dockerRepository,
	}, nil
}

// NewPRMExactRepository returns a new "exactRepository" PolicyRepositoryMatch.
func NewPRMExactRepository(dockerRepository string) (PolicyReferenceMatch, error) {
	return newPRMExactRepository(dockerRepository)
}

// Compile-time check that prmExactRepository implements json.Unmarshaler.
var _ json.Unmarshaler = (*prmExactRepository)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (prm *prmExactRepository) UnmarshalJSON(data []byte) error {
	*prm = prmExactRepository{}
	var tmp prmExactRepository
	if err := paranoidUnmarshalJSONObjectExactFields(data, map[string]interface{}{
		"type":             &tmp.Type,
		"dockerRepository": &tmp.DockerRepository,
	}); err != nil {
		return err
	}

	if tmp.Type != prmTypeExactRepository {
		return InvalidPolicyFormatError(fmt.Sprintf("Unexpected policy requirement type \"%s\"", tmp.Type))
	}

	res, err := newPRMExactRepository(tmp.DockerRepository)
	if err != nil {
		return err
	}
	*prm = *res
	return nil
}
