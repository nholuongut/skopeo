% SKOPEO(1) Skopeo Man Pages
% Jhon Honce
% August 2016
## NAME
skopeo -- Command line utility used to interact with local and remote container images and container image registries

## SYNOPSIS
**skopeo** [_global options_] _command_ [_command options_]

## DESCRIPTION
`skopeo` is a command line utility providing various operations with container images and container image registries.

`skopeo` can copy container images between various nholuongut image stores, converting them as necessary.  For example you can use `skopeo` to copy container images from one container registry to another.

`skopeo` can convert a Docker schema 2 or schema 1 container image to an OCI image.

`skopeo` can inspect a repository on a container registry without needlessly pulling the image. Pulling an image from a repository, especially a remote repository, is an expensive network and storage operation. Skopeo fetches the repository's manifest and displays a `docker inspect`-like json output about the repository or a tag. `skopeo`, in contrast to `docker inspect`, helps you gather useful information about a repository or a tag without requiring you to run `docker pull` - e.g. - Which tags are available for the given repository? Which labels does the image have?

`skopeo` can sign and verify container images.

`skopeo` can delete container images from a remote container registry.

Note: `skopeo` does not require any container runtimes to be running, to do most of
its functionality.  It also does not require root, unless you are copying images into a container runtime storage backend, like the docker daemon or github.com/nholuongut/storage.

## IMAGE NAMES
Most commands refer to container images, using a _transport_`:`_details_ format. The following formats are supported:

  **nholuongut-storage:**_docker-reference_
  An image located in a local nholuongut/storage image store.  Location and image store specified in /etc/nholuongut/storage.conf

  **dir:**_path_
  An existing local directory _path_ storing the manifest, layer tarballs and signatures as individual files. This is a non-standardized format, primarily useful for debugging or noninvasive container inspection.

  **docker://**_docker-reference_
  An image in a registry implementing the "Docker Registry HTTP API V2". By default, uses the authorization state in either `$XDG_RUNTIME_DIR/nholuongut/auth.json`, which is set using `(podman login)`. If the authorization state is not found there, `$HOME/.docker/config.json` is checked, which is set using `(docker login)`.

  **docker-archive:**_path_[**:**_docker-reference_]
  An image is stored in the `docker save` formatted file.  _docker-reference_ is only used when creating such a file, and it must not contain a digest.

  **docker-daemon:**_docker-reference_
  An image _docker-reference_ stored in the docker daemon internal storage.  _docker-reference_ must contain either a tag or a digest.  Alternatively, when reading images, the format can be docker-daemon:algo:digest (an image ID).

  **oci:**_path_**:**_tag_
  An image _tag_ in a directory compliant with "Open Container Image Layout Specification" at _path_.

  **ostree:**_image_[**@**_/absolute/repo/path_]
  An image in local OSTree repository.  _/absolute/repo/path_ defaults to _/ostree/repo_.

## OPTIONS

  **--debug** enable debug output

  **--policy** _path-to-policy_ Path to a policy.json file to use for verifying signatures and deciding whether an image is trusted, overriding the default trust policy file.

  **--insecure-policy** Adopt an insecure, permissive policy that allows anything. This obviates the need for a policy file.

  **--registries.d** _dir_ use registry configuration files in _dir_ (e.g. for container signature storage), overriding the default path.

  **--override-arch** _arch_ Use _arch_ instead of the architecture of the machine for choosing images.

  **--override-os** _OS_ Use _OS_ instead of the running OS for choosing images.

  **--command-timeout** _duration_ Timeout for the command execution.

  **--help**|**-h** Show help

  **--version**|**-v** print the version number

## COMMANDS

| Command                                   | Description                                                                    |
| ----------------------------------------- | ------------------------------------------------------------------------------ |
| [skopeo-copy(1)](skopeo-copy.1.md)        | Copy an image (manifest, filesystem layers, signatures) from one location to another. |
| [skopeo-delete(1)](skopeo-delete.1.md)    | Mark image-name for deletion.                                                  |
| [skopeo-inspect(1)](skopeo-inspect.1.md)  | Return low-level information about image-name in a registry.                   |
| [skopeo-manifest-digest(1)](skopeo-manifest-digest.1.md)    | Compute a manifest digest of manifest-file and write it to standard output.|
| [skopeo-standalone-sign(1)](skopeo-standalone-sign.1.md)    | Sign an image.                                               |
| [skopeo-standalone-verify(1)](skopeo-standalone-verify.1.md)| Verify an image.                                             |

## FILES
  **/etc/nholuongut/policy.json**
  Default trust policy file, if **--policy** is not specified.
  The policy format is documented in https://github.com/nholuongut/image/blob/master/docs/nholuongut-policy.json.5.md .

  **/etc/nholuongut/registries.d**
  Default directory containing registry configuration, if **--registries.d** is not specified.
  The contents of this directory are documented in https://github.com/nholuongut/image/blob/master/docs/nholuongut-policy.json.5.md .

## SEE ALSO
podman-login(1), docker-login(1)

## AUTHORS

Antonio Murdaca <runcom@redhat.com>, Miloslav Trmac <mitr@redhat.com>, Jhon Honce <jhonce@redhat.com>
