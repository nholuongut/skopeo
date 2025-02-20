package archive

import (
	"context"
	"io"

	ocilayout "github.com/nholuongut/image/oci/layout"
	"github.com/nholuongut/image/types"
	digest "github.com/opennholuongut/go-digest"
	imgspecv1 "github.com/opennholuongut/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type ociArchiveImageSource struct {
	ref         ociArchiveReference
	unpackedSrc types.ImageSource
	tempDirRef  tempDirOCIRef
}

// newImageSource returns an ImageSource for reading from an existing directory.
// newImageSource untars the file and saves it in a temp directory
func newImageSource(ctx context.Context, sys *types.SystemContext, ref ociArchiveReference) (types.ImageSource, error) {
	tempDirRef, err := createUntarTempDir(ref)
	if err != nil {
		return nil, errors.Wrap(err, "error creating temp directory")
	}

	unpackedSrc, err := tempDirRef.ociRefExtracted.NewImageSource(ctx, sys)
	if err != nil {
		if err := tempDirRef.deleteTempDir(); err != nil {
			return nil, errors.Wrapf(err, "error deleting temp directory %q", tempDirRef.tempDirectory)
		}
		return nil, err
	}
	return &ociArchiveImageSource{ref: ref,
		unpackedSrc: unpackedSrc,
		tempDirRef:  tempDirRef}, nil
}

// LoadManifestDescriptor loads the manifest
func LoadManifestDescriptor(imgRef types.ImageReference) (imgspecv1.Descriptor, error) {
	ociArchRef, ok := imgRef.(ociArchiveReference)
	if !ok {
		return imgspecv1.Descriptor{}, errors.Errorf("error typecasting, need type ociArchiveReference")
	}
	tempDirRef, err := createUntarTempDir(ociArchRef)
	if err != nil {
		return imgspecv1.Descriptor{}, errors.Wrap(err, "error creating temp directory")
	}
	defer tempDirRef.deleteTempDir()

	descriptor, err := ocilayout.LoadManifestDescriptor(tempDirRef.ociRefExtracted)
	if err != nil {
		return imgspecv1.Descriptor{}, errors.Wrap(err, "error loading index")
	}
	return descriptor, nil
}

// Reference returns the reference used to set up this source.
func (s *ociArchiveImageSource) Reference() types.ImageReference {
	return s.ref
}

// Close removes resources associated with an initialized ImageSource, if any.
// Close deletes the temporary directory at dst
func (s *ociArchiveImageSource) Close() error {
	defer s.tempDirRef.deleteTempDir()
	return s.unpackedSrc.Close()
}

// GetManifest returns the image's manifest along with its MIME type (which may be empty when it can't be determined but the manifest is available).
// It may use a remote (= slow) service.
// If instanceDigest is not nil, it contains a digest of the specific manifest instance to retrieve (when the primary manifest is a manifest list);
// this never happens if the primary manifest is not a manifest list (e.g. if the source never returns manifest lists).
func (s *ociArchiveImageSource) GetManifest(ctx context.Context, instanceDigest *digest.Digest) ([]byte, string, error) {
	return s.unpackedSrc.GetManifest(ctx, instanceDigest)
}

// HasThreadSafeGetBlob indicates whether GetBlob can be executed concurrently.
func (s *ociArchiveImageSource) HasThreadSafeGetBlob() bool {
	return false
}

// GetBlob returns a stream for the specified blob, and the blob’s size (or -1 if unknown).
// The Digest field in BlobInfo is guaranteed to be provided, Size may be -1 and MediaType may be optionally provided.
// May update BlobInfoCache, preferably after it knows for certain that a blob truly exists at a specific location.
func (s *ociArchiveImageSource) GetBlob(ctx context.Context, info types.BlobInfo, cache types.BlobInfoCache) (io.ReadCloser, int64, error) {
	return s.unpackedSrc.GetBlob(ctx, info, cache)
}

// GetSignatures returns the image's signatures.  It may use a remote (= slow) service.
// If instanceDigest is not nil, it contains a digest of the specific manifest instance to retrieve signatures for
// (when the primary manifest is a manifest list); this never happens if the primary manifest is not a manifest list
// (e.g. if the source never returns manifest lists).
func (s *ociArchiveImageSource) GetSignatures(ctx context.Context, instanceDigest *digest.Digest) ([][]byte, error) {
	return s.unpackedSrc.GetSignatures(ctx, instanceDigest)
}

// LayerInfosForCopy() returns updated layer info that should be used when reading, in preference to values in the manifest, if specified.
func (s *ociArchiveImageSource) LayerInfosForCopy(ctx context.Context) ([]types.BlobInfo, error) {
	return nil, nil
}
