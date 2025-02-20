package daemon

import (
	"context"

	"github.com/nholuongut/image/docker/tarfile"
	"github.com/nholuongut/image/types"
	"github.com/pkg/errors"
)

type daemonImageSource struct {
	ref             daemonReference
	*tarfile.Source // Implements most of types.ImageSource
}

type layerInfo struct {
	path string
	size int64
}

// newImageSource returns a types.ImageSource for the specified image reference.
// The caller must call .Close() on the returned ImageSource.
//
// It would be great if we were able to stream the input tar as it is being
// sent; but Docker sends the top-level manifest, which determines which paths
// to look for, at the end, so in we will need to seek back and re-read, several times.
// (We could, perhaps, expect an exact sequence, assume that the first plaintext file
// is the config, and that the following len(RootFS) files are the layers, but that feels
// way too brittle.)
func newImageSource(ctx context.Context, sys *types.SystemContext, ref daemonReference) (types.ImageSource, error) {
	c, err := newDockerClient(sys)
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing docker engine client")
	}
	// Per NewReference(), ref.StringWithinTransport() is either an image ID (config digest), or a !reference.NameOnly() reference.
	// Either way ImageSave should create a tarball with exactly one image.
	inputStream, err := c.ImageSave(ctx, []string{ref.StringWithinTransport()})
	if err != nil {
		return nil, errors.Wrap(err, "Error loading image from docker engine")
	}
	defer inputStream.Close()

	src, err := tarfile.NewSourceFromStream(inputStream)
	if err != nil {
		return nil, err
	}
	return &daemonImageSource{
		ref:    ref,
		Source: src,
	}, nil
}

// Reference returns the reference used to set up this source, _as specified by the user_
// (not as the image itself, or its underlying storage, claims).  This can be used e.g. to determine which public keys are trusted for this image.
func (s *daemonImageSource) Reference() types.ImageReference {
	return s.ref
}

// LayerInfosForCopy() returns updated layer info that should be used when reading, in preference to values in the manifest, if specified.
func (s *daemonImageSource) LayerInfosForCopy(ctx context.Context) ([]types.BlobInfo, error) {
	return nil, nil
}
