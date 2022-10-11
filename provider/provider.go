package provider

import (
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/site24x7/terraform-provider-site24x7/backoff"
	"github.com/site24x7/terraform-provider-site24x7/site24x7"
	"github.com/site24x7/terraform-provider-site24x7/site24x7/common"
	"github.com/site24x7/terraform-provider-site24x7/site24x7/integration"
	"github.com/site24x7/terraform-provider-site24x7/site24x7/monitors"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"oauth2_client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SITE24X7_OAUTH2_CLIENT_ID", nil),
				Description: "OAuth2 Client ID",
			},
			"oauth2_client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SITE24X7_OAUTH2_CLIENT_SECRET", nil),
				Description: "OAuth2 Client Secret",
			},
			"oauth2_refresh_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SITE24X7_OAUTH2_REFRESH_TOKEN", nil),
				Description: "OAuth2 Refresh Token",
			},
			"oauth2_access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SITE24X7_OAUTH2_ACCESS_TOKEN", nil),
				Description: "OAuth2 Access Token",
			},
			"access_token_expiry": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Access token expiry in seconds",
			},
			"data_center": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Site24x7 data center.",
			},
			"zaaid": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "MSP customer zaaid",
			},
			"retry_min_wait": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Minimum wait time in seconds before retrying failed API requests.",
			},
			"retry_max_wait": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
				Description: "Maximum wait time in seconds before retrying failed API requests (exponential backoff).",
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     4,
				Description: "Maximum number of retries for Site24x7 API errors until giving up",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"site24x7_website_monitor":        monitors.ResourceSite24x7WebsiteMonitor(),
			"site24x7_web_page_speed_monitor": monitors.ResourceSite24x7WebPageSpeedMonitor(),
			"site24x7_ssl_monitor":            monitors.ResourceSite24x7SSLMonitor(),
			"site24x7_rest_api_monitor":       monitors.ResourceSite24x7RestApiMonitor(),
			"site24x7_server_monitor":         monitors.ResourceSite24x7ServerMonitor(),
			"site24x7_monitor_group":          site24x7.ResourceSite24x7MonitorGroup(),
			"site24x7_subgroup":               site24x7.ResourceSite24x7Subgroup(),
			"site24x7_url_action":             site24x7.ResourceSite24x7URLAction(),
			"site24x7_threshold_profile":      site24x7.ResourceSite24x7ThresholdProfile(),
			"site24x7_location_profile":       site24x7.ResourceSite24x7LocationProfile(),
			"site24x7_notification_profile":   site24x7.ResourceSite24x7NotificationProfile(),
			"site24x7_user_group":             site24x7.ResourceSite24x7UserGroup(),
			"site24x7_tag":                    site24x7.ResourceSite24x7Tag(),
			"site24x7_schedule_maintenance":   common.ResourceSite24x7ScheduleMaintenance(),
			"site24x7_opsgenie_integration":   integration.ResourceSite24x7OpsgenieIntegration(),
			"site24x7_slack_integration":      integration.ResourceSite24x7SlackIntegration(),
			"site24x7_webhook_integration":    integration.ResourceSite24x7WebhookIntegration(),
			"site24x7_pagerduty_integration":  integration.ResourceSite24x7PagerDutyIntegration(),
			"site24x7_servicenow_integration": integration.ResourceSite24x7ServiceNowIntegration(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"site24x7_monitor":              monitors.DataSourceSite24x7Monitor(),
			"site24x7_monitors":             monitors.DataSourceSite24x7Monitors(),
			"site24x7_location_profile":     site24x7.DataSourceSite24x7LocationProfile(),
			"site24x7_threshold_profile":    site24x7.DataSourceSite24x7ThresholdProfile(),
			"site24x7_notification_profile": site24x7.DataSourceSite24x7NotificationProfile(),
			"site24x7_monitor_group":        site24x7.DataSourceSite24x7MonitorGroup(),
			"site24x7_user_group":           site24x7.DataSourceSite24x7UserGroup(),
			"site24x7_tag":                  site24x7.DataSourceSite24x7Tag(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	tfLog := os.Getenv("TF_LOG")
	if tfLog == "DEBUG" || tfLog == "TRACE" {
		log.SetLevel(log.DebugLevel)
	}
	dataCenter := site24x7.GetDataCenter(d.Get("data_center").(string))
	log.Println("GetAPIBaseURL : ", dataCenter.GetAPIBaseURL())
	log.Println("GetTokenURL : ", dataCenter.GetTokenURL())
	config := site24x7.Config{
		ClientID:     d.Get("oauth2_client_id").(string),
		ClientSecret: d.Get("oauth2_client_secret").(string),
		RefreshToken: d.Get("oauth2_refresh_token").(string),
		AccessToken:  d.Get("oauth2_access_token").(string),
		Expiry:       d.Get("access_token_expiry").(string),
		ZAAID:        d.Get("zaaid").(string),
		APIBaseURL:   dataCenter.GetAPIBaseURL(),
		TokenURL:     dataCenter.GetTokenURL(),
		RetryConfig: &backoff.RetryConfig{
			MinWait:    time.Duration(d.Get("retry_min_wait").(int)) * time.Second,
			MaxWait:    time.Duration(d.Get("retry_max_wait").(int)) * time.Second,
			MaxRetries: d.Get("max_retries").(int),
		},
	}

	return site24x7.New(config), nil
}
