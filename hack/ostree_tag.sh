#!/bin/bash

if test $(${GO:-go} env GOOS) != "linux" ; then
	exit 0
fi

if pkg-config ostree-1 &> /dev/null ; then
	# ostree: used by nholuongut/storage
	# nholuongut_image_ostree: used by nholuongut/image
	echo "ostree nholuongut_image_ostree"
fi
