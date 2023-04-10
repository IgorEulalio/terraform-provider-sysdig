package sysdig

import (
	"context"
	v2 "github.com/draios/terraform-provider-sysdig/sysdig/internal/client/v2"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	TeamSchemaThemeKey                      = "theme"
	TeamSchemaNameKey                       = "name"
	TeamSchemaDescriptionKey                = "description"
	TeamSchemaScopeByKey                    = "scope_by"
	TeamSchemaFilterKey                     = "filter"
	TeamSchemaEnableIBMPlatformMetricsKey   = "enable_ibm_platform_metrics"
	TeamSchemaIBMPlatformMetricsKey         = "ibm_platform_metrics"
	TeamSchemaCanUseSysdigCaptureKey        = "can_use_sysdig_capture"
	TeamSchemaCanUseInfrastructureEventsKey = "can_see_infrastructure_events"
	TeamSchemaCanUseAWSDataKey              = "can_use_aws_data"
	TeamSchemaUserRolesKey                  = "user_roles"
	TeamSchemaUserRolesEmailKey             = "email"
	TeamSchemaUserRolesRoleKey              = "role"
	TeamSchemaEntrypointKey                 = "entrypoint"
	TeamSchemaEntrypointTypeKey             = "type"
	TeamSchemaEntrypointSelectionKey        = "selection"
	TeamSchemaDefaultTeamKey                = "default_team"
	TeamSchemaVersionKey                    = "version"
)

func createBaseMonitorTeamSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		TeamSchemaThemeKey: {
			Type: schema.TypeString,
		},
		TeamSchemaNameKey: {
			Type: schema.TypeString,
		},
		TeamSchemaDescriptionKey: {
			Type: schema.TypeString,
		},
		TeamSchemaScopeByKey: {
			Type: schema.TypeString,
		},
		TeamSchemaFilterKey: {
			Type: schema.TypeString,
		},
		TeamSchemaEnableIBMPlatformMetricsKey: {
			Type: schema.TypeBool,
		},
		TeamSchemaIBMPlatformMetricsKey: {
			Type: schema.TypeString,
		},
		TeamSchemaCanUseSysdigCaptureKey: {
			Type: schema.TypeBool,
		},
		TeamSchemaCanUseInfrastructureEventsKey: {
			Type: schema.TypeBool,
		},
		TeamSchemaCanUseAWSDataKey: {
			Type: schema.TypeBool,
		},
		TeamSchemaUserRolesKey: {
			Type: schema.TypeSet,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					TeamSchemaUserRolesEmailKey: {
						Type: schema.TypeString,
					},
					TeamSchemaUserRolesRoleKey: {
						Type: schema.TypeString,
					},
				},
			},
		},
		TeamSchemaEntrypointKey: {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					TeamSchemaEntrypointTypeKey: {
						Type: schema.TypeString,
					},
					TeamSchemaEntrypointSelectionKey: {
						Type: schema.TypeString,
					},
				},
			},
		},
		TeamSchemaDefaultTeamKey: {
			Type: schema.TypeBool,
		},
		TeamSchemaVersionKey: {
			Type: schema.TypeInt,
		},
	}

	return s
}

func createMonitorTeamSchema() map[string]*schema.Schema {
	s := createBaseMonitorTeamSchema()
	applyOnSchema(s, func(s *schema.Schema) {
		s.Optional = true
	})

	s[TeamSchemaThemeKey].Default = "#05C391"
	s[TeamSchemaNameKey].Required = true
	s[TeamSchemaNameKey].Optional = false
	s[TeamSchemaScopeByKey].Default = "host"
	s[TeamSchemaCanUseSysdigCaptureKey].Default = false
	s[TeamSchemaCanUseInfrastructureEventsKey].Default = false
	s[TeamSchemaCanUseAWSDataKey].Default = false
	userRolesSchema := s[TeamSchemaUserRolesKey].Elem.(*schema.Resource).Schema
	userRolesSchema[TeamSchemaUserRolesEmailKey].Required = true
	userRolesSchema[TeamSchemaUserRolesEmailKey].Optional = false
	userRolesSchema[TeamSchemaUserRolesRoleKey].Default = "ROLE_TEAM_STANDARD"
	userRolesSchema[TeamSchemaUserRolesRoleKey].ValidateFunc = validation.StringInSlice([]string{
		"ROLE_TEAM_STANDARD", "ROLE_TEAM_EDIT", "ROLE_TEAM_READ", "ROLE_TEAM_MANAGER",
	}, false)
	s[TeamSchemaEntrypointKey].Required = true
	s[TeamSchemaEntrypointKey].Optional = false
	entrypointSchema := s[TeamSchemaEntrypointKey].Elem.(*schema.Resource).Schema
	entrypointSchema[TeamSchemaEntrypointTypeKey].Required = true
	entrypointSchema[TeamSchemaEntrypointTypeKey].Optional = false
	entrypointSchema[TeamSchemaEntrypointTypeKey].ValidateFunc = validation.StringInSlice([]string{
		"Explore", "Dashboards", "Events", "Alerts", "Settings",
	}, false)
	s[TeamSchemaDefaultTeamKey].Default = false
	s[TeamSchemaVersionKey].Computed = true
	s[TeamSchemaVersionKey].Optional = false
	return s
}

func resourceSysdigMonitorTeam() *schema.Resource {
	timeout := 5 * time.Minute

	return &schema.Resource{
		CreateContext: resourceSysdigMonitorTeamCreate,
		UpdateContext: resourceSysdigMonitorTeamUpdate,
		ReadContext:   resourceSysdigMonitorTeamRead,
		DeleteContext: resourceSysdigMonitorTeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(timeout),
			Update: schema.DefaultTimeout(timeout),
			Read:   schema.DefaultTimeout(timeout),
			Delete: schema.DefaultTimeout(5 * time.Minute), // Removing the team is for some reason slower.
		},

		Schema: createMonitorTeamSchema(),
	}
}

func getMonitorTeamClient(c SysdigClients) (v2.TeamInterface, error) {
	var client v2.TeamInterface
	var err error
	switch c.GetClientType() {
	case IBMMonitor:
		client, err = c.ibmMonitorClient()
		if err != nil {
			return nil, err
		}
	default:
		client, err = c.sysdigMonitorClientV2()
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func resourceSysdigMonitorTeamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clients := meta.(SysdigClients)
	client, err := getMonitorTeamClient(clients)
	if err != nil {
		return diag.FromErr(err)
	}

	team := teamFromResourceData(d, clients.GetClientType())
	team.Products = []string{"SDC"}

	team, err = client.CreateTeam(ctx, team)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(team.ID))
	resourceSysdigMonitorTeamRead(ctx, d, meta)

	return nil
}

// Retrieves the information of a resource form the file and loads it in Terraform
func resourceSysdigMonitorTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clients := meta.(SysdigClients)
	client, err := getMonitorTeamClient(clients)
	if err != nil {
		return diag.FromErr(err)
	}

	id, _ := strconv.Atoi(d.Id())
	t, err := client.GetTeamById(ctx, id)

	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	err = teamMonitorToResourceData(d, clients, t)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	return nil
}

func teamMonitorToResourceData(d *schema.ResourceData, c SysdigClients, t v2.Team) error {
	d.SetId(strconv.Itoa(t.ID))

	err := d.Set("version", t.Version)
	if err != nil {
		return err
	}
	err = d.Set("theme", t.Theme)
	if err != nil {
		return err
	}
	err = d.Set("name", t.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", t.Description)
	if err != nil {
		return err
	}
	err = d.Set("scope_by", t.Show)
	if err != nil {
		return err
	}
	err = d.Set("filter", t.Filter)
	if err != nil {
		return err
	}
	err = d.Set("can_use_sysdig_capture", t.CanUseSysdigCapture)
	if err != nil {
		return err
	}
	err = d.Set("can_see_infrastructure_events", t.CanUseCustomEvents)
	if err != nil {
		return err
	}
	err = d.Set("can_use_aws_data", t.CanUseAwsMetrics)
	if err != nil {
		return err
	}
	err = d.Set("default_team", t.DefaultTeam)
	if err != nil {
		return err
	}
	err = d.Set("user_roles", userMonitorRolesToSet(t.UserRoles))
	if err != nil {
		return err
	}
	err = d.Set("entrypoint", entrypointToSet(t.EntryPoint))
	if err != nil {
		return err
	}

	if c.GetClientType() == IBMMonitor {
		err = resourceSysdigMonitorTeamReadIBM(d, &t)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceSysdigMonitorTeamReadIBM(d *schema.ResourceData, t *v2.Team) error {
	var ibmPlatformMetrics *string
	if t.NamespaceFilters != nil {
		ibmPlatformMetrics = t.NamespaceFilters.IBMPlatformMetrics
	}
	err := d.Set("enable_ibm_platform_metrics", t.CanUseBeaconMetrics)
	if err != nil {
		return err
	}
	return d.Set("ibm_platform_metrics", ibmPlatformMetrics)
}

func userMonitorRolesToSet(userRoles []v2.UserRoles) (res []map[string]interface{}) {
	for _, role := range userRoles {
		if role.Admin { // Admins are added by default, so skip them
			continue
		}

		roleMap := map[string]interface{}{
			"email": role.Email,
			"role":  role.Role,
		}
		res = append(res, roleMap)
	}
	return
}

func entrypointToSet(entrypoint *v2.EntryPoint) (res []map[string]interface{}) {
	if entrypoint == nil {
		return
	}

	entrypointMap := map[string]interface{}{
		"type":      entrypoint.Module,
		"selection": entrypoint.Selection,
	}
	return append(res, entrypointMap)
}

func resourceSysdigMonitorTeamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clients := meta.(SysdigClients)
	client, err := getMonitorTeamClient(clients)
	if err != nil {
		return diag.FromErr(err)
	}

	t := teamFromResourceData(d, clients.GetClientType())
	t.Products = []string{"SDC"}

	t.Version = d.Get("version").(int)
	t.ID, _ = strconv.Atoi(d.Id())

	_, err = client.UpdateTeam(ctx, t)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceSysdigMonitorTeamRead(ctx, d, meta)
	return nil
}

func resourceSysdigMonitorTeamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getMonitorTeamClient(meta.(SysdigClients))
	if err != nil {
		return diag.FromErr(err)
	}

	id, _ := strconv.Atoi(d.Id())

	err = client.DeleteTeam(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func updateNamespaceFilters(filters *v2.NamespaceFilters, update v2.NamespaceFilters) *v2.NamespaceFilters {
	if filters == nil {
		filters = &v2.NamespaceFilters{}
	}

	if update.IBMPlatformMetrics != nil {
		filters.IBMPlatformMetrics = update.IBMPlatformMetrics
	}

	return filters
}

func teamFromResourceData(d *schema.ResourceData, clientType ClientType) v2.Team {
	canUseSysdigCapture := d.Get("can_use_sysdig_capture").(bool)
	canUseCustomEvents := d.Get("can_see_infrastructure_events").(bool)
	canUseAwsMetrics := d.Get("can_use_aws_data").(bool)
	canUseBeaconMetrics := false
	t := v2.Team{
		Theme:               d.Get("theme").(string),
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		Show:                d.Get("scope_by").(string),
		Filter:              d.Get("filter").(string),
		CanUseSysdigCapture: &canUseSysdigCapture,
		CanUseCustomEvents:  &canUseCustomEvents,
		CanUseAwsMetrics:    &canUseAwsMetrics,
		CanUseBeaconMetrics: &canUseBeaconMetrics,
		DefaultTeam:         d.Get("default_team").(bool),
	}

	userRoles := make([]v2.UserRoles, 0)
	for _, userRole := range d.Get("user_roles").(*schema.Set).List() {
		ur := userRole.(map[string]interface{})
		userRoles = append(userRoles, v2.UserRoles{
			Email: ur["email"].(string),
			Role:  ur["role"].(string),
		})
	}
	t.UserRoles = userRoles

	t.EntryPoint = &v2.EntryPoint{}
	t.EntryPoint.Module = d.Get("entrypoint.0.type").(string)
	if val, ok := d.GetOk("entrypoint.0.selection"); ok {
		t.EntryPoint.Selection = val.(string)
	}

	if clientType == IBMMonitor {
		teamFromResourceDataIBM(d, &t)
	}

	return t
}

func teamFromResourceDataIBM(d *schema.ResourceData, t *v2.Team) {
	canUseBeaconMetrics := d.Get("enable_ibm_platform_metrics").(bool)
	t.CanUseBeaconMetrics = &canUseBeaconMetrics

	if v, ok := d.GetOk("ibm_platform_metrics"); ok {
		metrics := v.(string)
		t.NamespaceFilters = updateNamespaceFilters(t.NamespaceFilters, v2.NamespaceFilters{
			IBMPlatformMetrics: &metrics,
		})
	}
}
