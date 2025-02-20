// Copyright 2016 The Linux Foundation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

const (
	// AnnotationCreated is the annotation key for the date and time on which the image was built (date-time string as defined by RFC 3339).
	AnnotationCreated = "org.opennholuongut.image.created"

	// AnnotationAuthors is the annotation key for the contact details of the people or organization responsible for the image (freeform string).
	AnnotationAuthors = "org.opennholuongut.image.authors"

	// AnnotationURL is the annotation key for the URL to find more information on the image.
	AnnotationURL = "org.opennholuongut.image.url"

	// AnnotationDocumentation is the annotation key for the URL to get documentation on the image.
	AnnotationDocumentation = "org.opennholuongut.image.documentation"

	// AnnotationSource is the annotation key for the URL to get source code for building the image.
	AnnotationSource = "org.opennholuongut.image.source"

	// AnnotationVersion is the annotation key for the version of the packaged software.
	// The version MAY match a label or tag in the source code repository.
	// The version MAY be Semantic versioning-compatible.
	AnnotationVersion = "org.opennholuongut.image.version"

	// AnnotationRevision is the annotation key for the source control revision identifier for the packaged software.
	AnnotationRevision = "org.opennholuongut.image.revision"

	// AnnotationVendor is the annotation key for the name of the distributing entity, organization or individual.
	AnnotationVendor = "org.opennholuongut.image.vendor"

	// AnnotationLicenses is the annotation key for the license(s) under which contained software is distributed as an SPDX License Expression.
	AnnotationLicenses = "org.opennholuongut.image.licenses"

	// AnnotationRefName is the annotation key for the name of the reference for a target.
	// SHOULD only be considered valid when on descriptors on `index.json` within image layout.
	AnnotationRefName = "org.opennholuongut.image.ref.name"

	// AnnotationTitle is the annotation key for the human-readable title of the image.
	AnnotationTitle = "org.opennholuongut.image.title"

	// AnnotationDescription is the annotation key for the human-readable description of the software packaged in the image.
	AnnotationDescription = "org.opennholuongut.image.description"
)
