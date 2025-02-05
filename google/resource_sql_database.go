// ----------------------------------------------------------------------------
//
//     ***     AUTO GENERATED CODE    ***    AUTO GENERATED CODE     ***
//
// ----------------------------------------------------------------------------
//
//     This file is automatically generated by Magic Modules and manual
//     changes will be clobbered when the file is regenerated.
//
//     Please read more about how to change this file in
//     .github/CONTRIBUTING.md.
//
// ----------------------------------------------------------------------------

package google

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

func resourceSQLDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceSQLDatabaseCreate,
		Read:   resourceSQLDatabaseRead,
		Update: resourceSQLDatabaseUpdate,
		Delete: resourceSQLDatabaseDelete,

		Importer: &schema.ResourceImporter{
			State: resourceSQLDatabaseImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: `The name of the Cloud SQL instance. This does not include the project
ID.`,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: `The name of the database in the Cloud SQL instance.
This does not include the project ID or instance name.`,
			},
			"charset": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				Description: `The charset value. See MySQL's
[Supported Character Sets and Collations](https://dev.mysql.com/doc/refman/5.7/en/charset-charsets.html)
and Postgres' [Character Set Support](https://www.postgresql.org/docs/9.6/static/multibyte.html)
for more details and supported values. Postgres databases only support
a value of 'UTF8' at creation time.`,
			},
			"collation": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				Description: `The collation value. See MySQL's
[Supported Character Sets and Collations](https://dev.mysql.com/doc/refman/5.7/en/charset-charsets.html)
and Postgres' [Collation Support](https://www.postgresql.org/docs/9.6/static/collation.html)
for more details and supported values. Postgres databases only support
a value of 'en_US.UTF8' at creation time.`,
			},
			"project": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"self_link": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSQLDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	obj := make(map[string]interface{})
	charsetProp, err := expandSQLDatabaseCharset(d.Get("charset"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("charset"); !isEmptyValue(reflect.ValueOf(charsetProp)) && (ok || !reflect.DeepEqual(v, charsetProp)) {
		obj["charset"] = charsetProp
	}
	collationProp, err := expandSQLDatabaseCollation(d.Get("collation"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("collation"); !isEmptyValue(reflect.ValueOf(collationProp)) && (ok || !reflect.DeepEqual(v, collationProp)) {
		obj["collation"] = collationProp
	}
	nameProp, err := expandSQLDatabaseName(d.Get("name"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("name"); !isEmptyValue(reflect.ValueOf(nameProp)) && (ok || !reflect.DeepEqual(v, nameProp)) {
		obj["name"] = nameProp
	}
	instanceProp, err := expandSQLDatabaseInstance(d.Get("instance"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("instance"); !isEmptyValue(reflect.ValueOf(instanceProp)) && (ok || !reflect.DeepEqual(v, instanceProp)) {
		obj["instance"] = instanceProp
	}

	lockName, err := replaceVars(d, config, "google-sql-database-instance-{{project}}-{{instance}}")
	if err != nil {
		return err
	}
	mutexKV.Lock(lockName)
	defer mutexKV.Unlock(lockName)

	url, err := replaceVars(d, config, "{{SQLBasePath}}projects/{{project}}/instances/{{instance}}/databases")
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Creating new Database: %#v", obj)
	project, err := getProject(d, config)
	if err != nil {
		return err
	}
	res, err := sendRequestWithTimeout(config, "POST", project, url, obj, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("Error creating Database: %s", err)
	}

	// Store the ID now
	id, err := replaceVars(d, config, "projects/{{project}}/instances/{{instance}}/databases/{{name}}")
	if err != nil {
		return fmt.Errorf("Error constructing id: %s", err)
	}
	d.SetId(id)

	op := &sqladmin.Operation{}
	err = Convert(res, op)
	if err != nil {
		return err
	}

	waitErr := sqlAdminOperationWaitTime(
		config.clientSqlAdmin, op, project, "Creating Database",
		int(d.Timeout(schema.TimeoutCreate).Minutes()))

	if waitErr != nil {
		// The resource didn't actually create
		d.SetId("")
		return fmt.Errorf("Error waiting to create Database: %s", waitErr)
	}

	log.Printf("[DEBUG] Finished creating Database %q: %#v", d.Id(), res)

	return resourceSQLDatabaseRead(d, meta)
}

func resourceSQLDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	url, err := replaceVars(d, config, "{{SQLBasePath}}projects/{{project}}/instances/{{instance}}/databases/{{name}}")
	if err != nil {
		return err
	}

	project, err := getProject(d, config)
	if err != nil {
		return err
	}
	res, err := sendRequest(config, "GET", project, url, nil)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("SQLDatabase %q", d.Id()))
	}

	if err := d.Set("project", project); err != nil {
		return fmt.Errorf("Error reading Database: %s", err)
	}

	if err := d.Set("charset", flattenSQLDatabaseCharset(res["charset"], d)); err != nil {
		return fmt.Errorf("Error reading Database: %s", err)
	}
	if err := d.Set("collation", flattenSQLDatabaseCollation(res["collation"], d)); err != nil {
		return fmt.Errorf("Error reading Database: %s", err)
	}
	if err := d.Set("name", flattenSQLDatabaseName(res["name"], d)); err != nil {
		return fmt.Errorf("Error reading Database: %s", err)
	}
	if err := d.Set("instance", flattenSQLDatabaseInstance(res["instance"], d)); err != nil {
		return fmt.Errorf("Error reading Database: %s", err)
	}
	if err := d.Set("self_link", ConvertSelfLinkToV1(res["selfLink"].(string))); err != nil {
		return fmt.Errorf("Error reading Database: %s", err)
	}

	return nil
}

func resourceSQLDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	obj := make(map[string]interface{})
	charsetProp, err := expandSQLDatabaseCharset(d.Get("charset"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("charset"); !isEmptyValue(reflect.ValueOf(v)) && (ok || !reflect.DeepEqual(v, charsetProp)) {
		obj["charset"] = charsetProp
	}
	collationProp, err := expandSQLDatabaseCollation(d.Get("collation"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("collation"); !isEmptyValue(reflect.ValueOf(v)) && (ok || !reflect.DeepEqual(v, collationProp)) {
		obj["collation"] = collationProp
	}
	nameProp, err := expandSQLDatabaseName(d.Get("name"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("name"); !isEmptyValue(reflect.ValueOf(v)) && (ok || !reflect.DeepEqual(v, nameProp)) {
		obj["name"] = nameProp
	}
	instanceProp, err := expandSQLDatabaseInstance(d.Get("instance"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("instance"); !isEmptyValue(reflect.ValueOf(v)) && (ok || !reflect.DeepEqual(v, instanceProp)) {
		obj["instance"] = instanceProp
	}

	lockName, err := replaceVars(d, config, "google-sql-database-instance-{{project}}-{{instance}}")
	if err != nil {
		return err
	}
	mutexKV.Lock(lockName)
	defer mutexKV.Unlock(lockName)

	url, err := replaceVars(d, config, "{{SQLBasePath}}projects/{{project}}/instances/{{instance}}/databases/{{name}}")
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Updating Database %q: %#v", d.Id(), obj)
	res, err := sendRequestWithTimeout(config, "PUT", project, url, obj, d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return fmt.Errorf("Error updating Database %q: %s", d.Id(), err)
	}

	op := &sqladmin.Operation{}
	err = Convert(res, op)
	if err != nil {
		return err
	}

	err = sqlAdminOperationWaitTime(
		config.clientSqlAdmin, op, project, "Updating Database",
		int(d.Timeout(schema.TimeoutUpdate).Minutes()))

	if err != nil {
		return err
	}

	return resourceSQLDatabaseRead(d, meta)
}

func resourceSQLDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	lockName, err := replaceVars(d, config, "google-sql-database-instance-{{project}}-{{instance}}")
	if err != nil {
		return err
	}
	mutexKV.Lock(lockName)
	defer mutexKV.Unlock(lockName)

	url, err := replaceVars(d, config, "{{SQLBasePath}}projects/{{project}}/instances/{{instance}}/databases/{{name}}")
	if err != nil {
		return err
	}

	var obj map[string]interface{}
	log.Printf("[DEBUG] Deleting Database %q", d.Id())

	res, err := sendRequestWithTimeout(config, "DELETE", project, url, obj, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return handleNotFoundError(err, d, "Database")
	}

	op := &sqladmin.Operation{}
	err = Convert(res, op)
	if err != nil {
		return err
	}

	err = sqlAdminOperationWaitTime(
		config.clientSqlAdmin, op, project, "Deleting Database",
		int(d.Timeout(schema.TimeoutDelete).Minutes()))

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Database %q: %#v", d.Id(), res)
	return nil
}

func resourceSQLDatabaseImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if err := parseImportId([]string{
		"projects/(?P<project>[^/]+)/instances/(?P<instance>[^/]+)/databases/(?P<name>[^/]+)",
		"instances/(?P<instance>[^/]+)/databases/(?P<name>[^/]+)",
		"(?P<project>[^/]+)/(?P<instance>[^/]+)/(?P<name>[^/]+)",
		"(?P<instance>[^/]+)/(?P<name>[^/]+)",
		"(?P<name>[^/]+)",
	}, d, config); err != nil {
		return nil, err
	}

	// Replace import id for the resource id
	id, err := replaceVars(d, config, "projects/{{project}}/instances/{{instance}}/databases/{{name}}")
	if err != nil {
		return nil, fmt.Errorf("Error constructing id: %s", err)
	}
	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}

func flattenSQLDatabaseCharset(v interface{}, d *schema.ResourceData) interface{} {
	return v
}

func flattenSQLDatabaseCollation(v interface{}, d *schema.ResourceData) interface{} {
	return v
}

func flattenSQLDatabaseName(v interface{}, d *schema.ResourceData) interface{} {
	return v
}

func flattenSQLDatabaseInstance(v interface{}, d *schema.ResourceData) interface{} {
	return v
}

func expandSQLDatabaseCharset(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandSQLDatabaseCollation(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandSQLDatabaseName(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandSQLDatabaseInstance(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}
