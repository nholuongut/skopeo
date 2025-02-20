// Policy evaluation for the various simple PolicyRequirement types.

package signature

import (
	"context"
	"fmt"

	"github.com/nholuongut/image/transports"
	"github.com/nholuongut/image/types"
)

func (pr *prInsecureAcceptAnything) isSignatureAuthorAccepted(ctx context.Context, image types.UnparsedImage, sig []byte) (signatureAcceptanceResult, *Signature, error) {
	// prInsecureAcceptAnything semantics: Every image is allowed to run,
	// but this does not consider the signature as verified.
	return sarUnknown, nil, nil
}

func (pr *prInsecureAcceptAnything) isRunningImageAllowed(ctx context.Context, image types.UnparsedImage) (bool, error) {
	return true, nil
}

func (pr *prReject) isSignatureAuthorAccepted(ctx context.Context, image types.UnparsedImage, sig []byte) (signatureAcceptanceResult, *Signature, error) {
	return sarRejected, nil, PolicyRequirementError(fmt.Sprintf("Any signatures for image %s are rejected by policy.", transports.ImageName(image.Reference())))
}

func (pr *prReject) isRunningImageAllowed(ctx context.Context, image types.UnparsedImage) (bool, error) {
	return false, PolicyRequirementError(fmt.Sprintf("Running image %s is rejected by policy.", transports.ImageName(image.Reference())))
}
