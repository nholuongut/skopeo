package directory

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/nholuongut/image/manifest"
	"github.com/nholuongut/image/types"
	"github.com/opennholuongut/go-digest"
	"github.com/pkg/errors"
)

type dirImageSource struct {
	ref dirReference
}

// newImageSource returns an ImageSource reading from an existing directory.
// The caller must call .Close() on the returned ImageSource.
func newImageSource(ref dirReference) types.ImageSource {
	return &dirImageSource{ref}
}

// Reference returns the reference used to set up this source, _as specified by the user_
// (not as the image itself, or its underlying storage, claims).  This can be used e.g. to determine which public keys are trusted for this image.
func (s *dirImageSource) Reference() types.ImageReference {
	return s.ref
}

// Close removes resources associated with an initialized ImageSource, if any.
func (s *dirImageSource) Close() error {
	return nil
}

// GetManifest returns the image's manifest along with its MIME type (which may be empty when it can't be determined but the manifest is available).
// It may use a remote (= slow) service.
// If instanceDigest is not nil, it contains a digest of the specific manifest instance to retrieve (when the primary manifest is a manifest list);
// this never happens if the primary manifest is not a manifest list (e.g. if the source never returns manifest lists).
func (s *dirImageSource) GetManifest(ctx context.Context, instanceDigest *digest.Digest) ([]byte, string, error) {
	if instanceDigest != nil {
		return nil, "", errors.Errorf(`Getting target manifest not supported by "dir:"`)
	}
	m, err := ioutil.ReadFile(s.ref.manifestPath())
	if err != nil {
		return nil, "", err
	}
	return m, manifest.GuessMIMEType(m), err
}

// HasThreadSafeGetBlob indicates whether GetBlob can be executed concurrently.
func (s *dirImageSource) HasThreadSafeGetBlob() bool {
	return false
}

// GetBlob returns a stream for the specified blob, and the blob’s size (or -1 if unknown).
// The Digest field in BlobInfo is guaranteed to be provided, Size may be -1 and MediaType may be optionally provided.
// May update BlobInfoCache, preferably after it knows for certain that a blob truly exists at a specific location.
func (s *dirImageSource) GetBlob(ctx context.Context, info types.BlobInfo, cache types.BlobInfoCache) (io.ReadCloser, int64, error) {
	r, err := os.Open(s.ref.layerPath(info.Digest))
	if err != nil {
		return nil, -1, err
	}
	fi, err := r.Stat()
	if err != nil {
		return nil, -1, err
	}
	return r, fi.Size(), nil
}

// GetSignatures returns the image's signatures.  It may use a remote (= slow) service.
// If instanceDigest is not nil, it contains a digest of the specific manifest instance to retrieve signatures for
// (when the primary manifest is a manifest list); this never happens if the primary manifest is not a manifest list
// (e.g. if the source never returns manifest lists).
func (s *dirImageSource) GetSignatures(ctx context.Context, instanceDigest *digest.Digest) ([][]byte, error) {
	if instanceDigest != nil {
		return nil, errors.Errorf(`Manifests lists are not supported by "dir:"`)
	}
	signatures := [][]byte{}
	for i := 0; ; i++ {
		signature, err := ioutil.ReadFile(s.ref.signaturePath(i))
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
			return nil, err
		}
		signatures = append(signatures, signature)
	}
	return signatures, nil
}

// LayerInfosForCopy() returns updated layer info that should be used when copying, in preference to values in the manifest, if specified.
func (s *dirImageSource) LayerInfosForCopy(ctx context.Context) ([]types.BlobInfo, error) {
	return nil, nil
}
