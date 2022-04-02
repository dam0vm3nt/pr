/*
 * Bitbucket API
 *
 * Code against the Bitbucket API to automate simple tasks, embed Bitbucket data into your own site, build mobile or desktop apps, or even add custom UI add-ons into Bitbucket itself using the Connect framework.
 *
 * API version: 2.0
 * Contact: support@bitbucket.org
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package bitbucket

type PipelineRefTarget struct {
	Type_ string `json:"type"`
	// The type of reference (branch/tag).
	RefType string `json:"ref_type,omitempty"`
	// The name of the reference.
	RefName  string            `json:"ref_name,omitempty"`
	Commit   *Commit           `json:"commit,omitempty"`
	Selector *PipelineSelector `json:"selector,omitempty"`
}
