// Copy Kibana object from space to another spaces
// API documentation: https://www.elastic.co/guide/en/kibana/master/spaces-api-copy-saved-objects.html
// Supported version:
//  - v7

package kb

import (
	"context"
	"fmt"

	kibana "github.com/disaster37/go-kibana-rest/v8"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

// Resource specification to handle kibana save object
func resourceKibanaCopyObject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKibanaCopyObjectCreate,
		ReadContext:   resourceKibanaCopyObjectRead,
		UpdateContext: resourceKibanaCopyObjectUpdate,
		DeleteContext: resourceKibanaCopyObjectDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_space": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"target_spaces": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"object": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"include_reference": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"create_new_copies": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"force_update": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

// Copy objects in Kibana
func resourceKibanaCopyObjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	err := copyObject(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)

	log.Infof("Copy objects %s successfully", name)
	fmt.Printf("[INFO] Copy objects %s successfully", name)

	return resourceKibanaCopyObjectRead(ctx, d, meta)
}

// Read object on kibana
func resourceKibanaCopyObjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	id := d.Id()
	sourceSpace := d.Get("source_space").(string)
	targetSpaces := convertArrayInterfaceToArrayString(d.Get("target_spaces").(*schema.Set).List())
	objects := buildCopyObjects(d.Get("object").(*schema.Set).List())
	includeReference := d.Get("include_reference").(bool)
	overwrite := d.Get("overwrite").(bool)
	createNewCopies := d.Get("create_new_copies").(bool)
	forceUpdate := d.Get("force_update").(bool)

	log.Debugf("Resource id:  %s", id)
	log.Debugf("Source space: %s", sourceSpace)
	log.Debugf("Target spaces: %+v", targetSpaces)
	log.Debugf("Objects: %+v", objects)
	log.Debugf("Include reference: %t", includeReference)
	log.Debugf("Overwrite: %t", overwrite)
	log.Debugf("CreateNewCopies: %t", createNewCopies)
	log.Debugf("force_update: %t", forceUpdate)

	// @ TODO
	// A good when is to check if already exported object is the same that original space
	// To avoid this hard code, we juste use force_update and overwrite field
	// It make same result but in bad way on terraform spirit

	if err = d.Set("name", id); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("source_space", sourceSpace); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("target_spaces", targetSpaces); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("object", objects); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("include_reference", includeReference); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("overwrite", overwrite); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("create_new_copies", createNewCopies); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("force_update", false); err != nil {
		return diag.FromErr(err)
	}

	log.Infof("Read resource %s successfully", id)
	fmt.Printf("[INFO] Read resource %s successfully", id)

	return nil
}

// Update existing object in Kibana
func resourceKibanaCopyObjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	err := copyObject(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Infof("Updated resource %s successfully", id)
	fmt.Printf("[INFO] Updated resource %s successfully", id)

	return resourceKibanaCopyObjectRead(ctx, d, meta)
}

// Delete object in Kibana is not supported
// It just remove object from state
func resourceKibanaCopyObjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	d.SetId("")

	log.Infof("Delete object in not supported - just removing from state")
	fmt.Printf("[INFO] Delete object in not supported - just removing from state")
	return nil

}

// Build list of object to export
func buildCopyObjects(raws []interface{}) []map[string]string {

	results := make([]map[string]string, len(raws))

	for i, raw := range raws {
		m := raw.(map[string]interface{})
		object := map[string]string{}
		object["type"] = m["type"].(string)
		object["id"] = m["id"].(string)
		results[i] = object
	}

	return results
}

// Copy objects in Kibana
func copyObject(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	sourceSpace := d.Get("source_space").(string)
	targetSpaces := convertArrayInterfaceToArrayString(d.Get("target_spaces").(*schema.Set).List())
	objects := buildCopyObjects(d.Get("object").(*schema.Set).List())
	includeReference := d.Get("include_reference").(bool)
	overwrite := d.Get("overwrite").(bool)
	createNewCopies := d.Get("create_new_copies").(bool)

	log.Debugf("Source space: %s", sourceSpace)
	log.Debugf("Target spaces: %+v", targetSpaces)
	log.Debugf("Objects: %+v", objects)
	log.Debugf("Include reference: %t", includeReference)
	log.Debugf("Overwrite: %t", overwrite)
	log.Debugf("CreateNewCopies: %t", createNewCopies)

	client := meta.(*kibana.Client)

	objectsParameter := make([]kbapi.KibanaSpaceObjectParameter, 0, 1)
	for _, object := range objects {
		objectsParameter = append(objectsParameter, kbapi.KibanaSpaceObjectParameter{
			ID:   object["id"],
			Type: object["type"],
		})
	}

	parameter := &kbapi.KibanaSpaceCopySavedObjectParameter{
		Spaces:            targetSpaces,
		Objects:           objectsParameter,
		IncludeReferences: includeReference,
		Overwrite:         overwrite,
		CreateNewCopies:   createNewCopies,
	}

	err := client.API.KibanaSpaces.CopySavedObjects(parameter, sourceSpace)
	if err != nil {
		return err
	}

	log.Debugf("Copy object for resource successfully: %s", name)

	return nil
}
