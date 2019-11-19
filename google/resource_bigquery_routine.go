//
package google

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"google.golang.org/api/bigquery/v2"
)

var standardSQLDataType = map[string]*schema.Schema{
	"type_kind": {
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.StringInSlice([]string{
			"TYPE_KIND_UNSPECIFIED",
			"INT64",
			"BOOL",
			"FLOAT64",
			"STRING",
			"BYTES",
			"TIMESTAMP",
			"DATE",
			"TIME",
			"DATETIME",
			"GEOGRAPHY",
			"NUMERIC",
			"ARRAY",
			"STRUCT",
		}, false),
	},
}

func init() {
	standardSQLDataType["array_element_type"] = &schema.Schema{
		Type: schema.TypeSet,
		Elem: &schema.Resource{
			Schema: standardSQLDataType,
		},
		Optional: true,
		MaxItems: 1,
	}

	standardSQLDataType["struct_type"] = &schema.Schema{
		Type: schema.TypeSet,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"fields": {
					Type: schema.TypeSet,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"type": {
								Type: schema.TypeSet,
								Elem: &schema.Resource{
									Schema: standardSQLDataType,
								},
								Optional: true,
								MaxItems: 1,
							},
						},
					},
					Required: true,
				},
			},
		},
		Optional: true,
		MaxItems: 1,
	}
}

func resourceBigQueryRoutine() *schema.Resource {
	return &schema.Resource{
		Create: resourceBigQueryRoutineCreate,
		Read:   resourceBigQueryRoutineRead,
		Delete: resourceBigQueryRoutineDelete,
		Update: resourceBigQueryRoutineUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			// RoutineId: [Required] The ID of the routine. The ID must contain only
			// letters (a-z, A-Z), numbers (0-9), or underscores (_). The maximum
			// length is 256 characters.
			"routine_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// DatasetId: [Required] The ID of the dataset containing this routine.
			"dataset_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// ProjectId: [Required] The ID of the project containing this routine.
			"project": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// RoutineType: [Required] The type of routine.
			// ROUTINE_TYPE_UNSPECIFIED, SCALAR_FUNCTION or PROCEDURE
			"routine_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ROUTINE_TYPE_UNSPECIFIED", "SCALAR_FUNCTION", "PROCEDURE"}, false),
			},

			// Description: [Optional] A user-friendly description of this routine.
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// CreationTime: [Output-only] The time when this routine was created, in
			// milliseconds since the epoch.
			"creation_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			// Etag: [Output-only] A hash of this resource.
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Language: [Optional] The language of the routine.
			// LANGUAGE_UNSPECIFIED, SQL or JAVASCRIPT
			"language": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"LANGUAGE_UNSPECIFIED", "SQL", "JAVASCRIPT"}, false),
			},

			// LastModifiedTime: [Output-only] The time when this routine was last
			// modified, in milliseconds since the epoch.
			"last_modified_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			// Arguments: [Optional] Input/output argument of a function or a stored procedure.
			"arguments": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"argument_kind": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"ARGUMENT_KIND_UNSPECIFIED", "FIXED_TYPE", "ANY_TYPE"}, false),
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"MODE_UNSPECIFIED", "IN", "OUT", "INOUT"}, false),
						},
						"data_type": {
							Type:     schema.TypeSet,
							Elem:     &schema.Resource{Schema: standardSQLDataType},
							Optional: true,
							MaxItems: 1,
						},
					},
				},
				Optional: true,
			},

			"return_type": {
				Type:     schema.TypeSet,
				Elem:     &schema.Resource{Schema: standardSQLDataType},
				MaxItems: 1,
				Optional: true,
			},

			"imported_libraries": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			"definition_body": {
				Type:     schema.TypeString,
				Required: true,
			},

			// SelfLink: [Output-only] A URL that can be used to access this
			// resource again.
			"self_link": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRoutine(d *schema.ResourceData, meta interface{}) (*bigquery.Routine, error) {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return nil, err
	}

	routine := &bigquery.Routine{
		RoutineReference: &bigquery.RoutineReference{
			DatasetId: d.Get("dataset_id").(string),
			RoutineId: d.Get("routine_id").(string),
			ProjectId: project,
		},
	}

	if v, ok := d.GetOk("routine_type"); ok {
		routine.RoutineType = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		routine.Description = v.(string)
	}

	if v, ok := d.GetOk("language"); ok {
		routine.Language = v.(string)
	}

	if v, ok := d.GetOk("arguments"); ok {
		arguments := v.(*schema.Set).List()
		routine.Arguments = make([]*bigquery.Argument, 0, len(arguments))
		for idx, v := range arguments {
			rd := v.(*schema.ResourceData)
			routine.Arguments[idx] = new(bigquery.Argument)

			if v, ok := rd.GetOk("name"); ok {
				routine.Arguments[idx].Name = v.(string)
			}

			if v, ok := rd.GetOk("argument_kind"); ok {
				routine.Arguments[idx].ArgumentKind = v.(string)
			}

			if v, ok := rd.GetOk("mode"); ok {
				routine.Arguments[idx].Mode = v.(string)
			}

			if v, ok := rd.GetOk("data_type"); ok {
				routine.Arguments[idx].DataType = new(bigquery.StandardSqlDataType)
				if err := setArgumentDataType(
					routine.Arguments[idx].DataType,
					v.(*schema.Set).List()[0].(*schema.ResourceData),
				); err != nil {
					return nil, err
				}
			}
		}
	}

	return routine, nil
}

func resourceBigQueryRoutineCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	routine, err := resourceRoutine(d, meta)
	if err != nil {
		return err
	}

	datasetID := d.Get("dataset_id").(string)

	log.Printf("[INFO] Creating BigQuery routine: %s", routine.RoutineReference.RoutineId)

	res, err := config.clientBigQuery.Routines.Insert(project, datasetID, routine).Do()
	if err != nil {
		return err
	}

	log.Printf("[INFO] BigQuery routine %s has been created", res.RoutineReference.RoutineId)

	d.SetId(fmt.Sprintf("projects/%s/datasets/%s/routines/%s", res.RoutineReference.ProjectId, res.RoutineReference.DatasetId, res.RoutineReference.RoutineId))

	return resourceBigQueryRoutineRead(d, meta)
}

func resourceBigQueryRoutineRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[INFO] Reading BigQuery routine: %s", d.Id())

	id, err := parseBigQueryRoutineId(d.Id())
	if err != nil {
		return err
	}

	res, err := config.clientBigQuery.Routines.Get(id.Project, id.DatasetId, id.RoutineId).Do()
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("BigQuery routine %q", id.RoutineId))
	}

	d.Set("project", id.Project)
	d.Set("description", res.Description)
	d.Set("creation_time", res.CreationTime)
	d.Set("etag", res.Etag)
	d.Set("last_modified_time", res.LastModifiedTime)
	d.Set("routine_id", res.RoutineReference.RoutineId)
	d.Set("dataset_id", res.RoutineReference.DatasetId)

	return nil
}

func resourceBigQueryRoutineUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	routine, err := resourceRoutine(d, meta)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Updating BigQuery routine: %s", d.Id())

	id, err := parseBigQueryRoutineId(d.Id())
	if err != nil {
		return err
	}

	if _, err = config.clientBigQuery.Routines.Update(id.Project, id.DatasetId, id.RoutineId, routine).Do(); err != nil {
		return err
	}

	return resourceBigQueryRoutineRead(d, meta)
}

func resourceBigQueryRoutineDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[INFO] Deleting BigQuery routine: %s", d.Id())

	id, err := parseBigQueryRoutineId(d.Id())
	if err != nil {
		return err
	}

	if err := config.clientBigQuery.Routines.Delete(id.Project, id.DatasetId, id.RoutineId).Do(); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

type bigQueryRoutineId struct {
	Project, DatasetId, RoutineId string
}

func parseBigQueryRoutineId(id string) (*bigQueryRoutineId, error) {
	// Expected format is "projects/{{project}}/datasets/{{dataset}}/routines/{{routine}}"
	matchRegex := regexp.MustCompile("^projects/(.+)/datasets/(.+)/routines/(.+)$")
	subMatches := matchRegex.FindStringSubmatch(id)
	if subMatches == nil {
		return nil, fmt.Errorf("Invalid BigQuery routine specifier. Expecting projects/{{project}}/datasets/{{dataset}}/routines/{{routine}}, got %s", id)
	}
	return &bigQueryRoutineId{
		Project:   subMatches[1],
		DatasetId: subMatches[2],
		RoutineId: subMatches[3],
	}, nil
}

func setArgumentDataType(sqldt *bigquery.StandardSqlDataType, rd *schema.ResourceData) error {
	if v, ok := rd.GetOk("type_kind"); ok {
		sqldt.TypeKind = v.(string)
	}

	if v, ok := rd.GetOk("array_element_type"); ok {
		sqldt.ArrayElementType = new(bigquery.StandardSqlDataType)
		if err := setArgumentDataType(sqldt.ArrayElementType, v.(*schema.ResourceData)); err != nil {
			return err
		}
	}

	if v, ok := rd.GetOk("struct_type"); ok {
		v.(*schema.Set).List()[0].()
	}

	return nil
}
