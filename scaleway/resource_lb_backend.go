package scaleway

import (
	"context"
	"math"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayLbBackend() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayLbBackendCreate,
		ReadContext:   resourceScalewayLbBackendRead,
		UpdateContext: resourceScalewayLbBackendUpdate,
		DeleteContext: resourceScalewayLbBackendDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultLbLbTimeout),
			Read:    schema.DefaultTimeout(defaultLbLbTimeout),
			Update:  schema.DefaultTimeout(defaultLbLbTimeout),
			Delete:  schema.DefaultTimeout(defaultLbLbTimeout),
			Default: schema.DefaultTimeout(defaultLbLbTimeout),
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{Version: 0, Type: lbUpgradeV1SchemaType(), Upgrade: lbUpgradeV1SchemaUpgradeFunc},
		},
		Schema: map[string]*schema.Schema{
			"lb_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The load-balancer ID",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the backend",
			},
			"forward_protocol": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					lbSDK.ProtocolTCP.String(),
					lbSDK.ProtocolHTTP.String(),
				}, false),
				Required:    true,
				Description: "Backend protocol",
			},
			"forward_port": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "User sessions will be forwarded to this port of backend servers",
			},
			"forward_port_algorithm": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					lbSDK.ForwardPortAlgorithmRoundrobin.String(),
					lbSDK.ForwardPortAlgorithmLeastconn.String(),
					lbSDK.ForwardPortAlgorithmFirst.String(),
				}, false),
				Default:     lbSDK.ForwardPortAlgorithmRoundrobin.String(),
				Optional:    true,
				Description: "Load balancing algorithm",
			},
			"sticky_sessions": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					lbSDK.StickySessionsTypeNone.String(),
					lbSDK.StickySessionsTypeCookie.String(),
					lbSDK.StickySessionsTypeTable.String(),
				}, false),
				Default:     lbSDK.StickySessionsTypeNone.String(),
				Optional:    true,
				Description: "The type of sticky sessions",
			},
			"sticky_sessions_cookie_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Cookie name for for sticky sessions",
			},
			"server_ips": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPAddress,
				},
				Optional:    true,
				Description: "Backend server IP addresses list (IPv4 or IPv6)",
			},
			"send_proxy_v2": {
				Type:        schema.TypeBool,
				Description: "Enables PROXY protocol version 2",
				Optional:    true,
				Computed:    true,
				Deprecated:  "Please use proxy_protocol instead",
			},
			"proxy_protocol": {
				Type:        schema.TypeString,
				Description: "Type of PROXY protocol to enable",
				Optional:    true,
				Default:     flattenLbProxyProtocol(lbSDK.ProxyProtocolProxyProtocolNone).(string),
				ValidateFunc: validation.StringInSlice([]string{
					flattenLbProxyProtocol(lbSDK.ProxyProtocolProxyProtocolNone).(string),
					flattenLbProxyProtocol(lbSDK.ProxyProtocolProxyProtocolV1).(string),
					flattenLbProxyProtocol(lbSDK.ProxyProtocolProxyProtocolV2).(string),
					flattenLbProxyProtocol(lbSDK.ProxyProtocolProxyProtocolV2Ssl).(string),
					flattenLbProxyProtocol(lbSDK.ProxyProtocolProxyProtocolV2SslCn).(string),
				}, false),
			},
			// Timeouts
			"timeout_server": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "5m",
				DiffSuppressFunc: diffSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Description:      "Maximum server connection inactivity time",
			},
			"timeout_connect": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "5s",
				DiffSuppressFunc: diffSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Description:      "Maximum initial server connection establishment time",
			},
			"timeout_tunnel": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "15m",
				DiffSuppressFunc: diffSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Description:      "Maximum tunnel inactivity time",
			},

			// Health Check
			"health_check_timeout": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Default:          "30s",
				Description:      "Timeout before we consider a HC request failed",
			},
			"health_check_delay": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Default:          "60s",
				Description:      "Interval between two HC requests",
			},
			"health_check_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Port the HC requests will be send to. Default to `forward_port`",
			},
			"health_check_max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     2,
				Description: "Number of allowed failed HC requests before the backend server is marked down",
			},
			"health_check_tcp": {
				Type:          schema.TypeList,
				MaxItems:      1,
				ConflictsWith: []string{"health_check_http", "health_check_https"},
				Optional:      true,
				Computed:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
			"health_check_http": {
				Type:          schema.TypeList,
				MaxItems:      1,
				ConflictsWith: []string{"health_check_tcp", "health_check_https"},
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The HTTP endpoint URL to call for HC requests",
						},
						"method": {
							Type:        schema.TypeString,
							Default:     "GET",
							Optional:    true,
							Description: "The HTTP method to use for HC requests",
						},
						"code": {
							Type:        schema.TypeInt,
							Default:     200,
							Optional:    true,
							Description: "The expected HTTP status code",
						},
						"host_header": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The HTTP host header to use for HC requests",
						},
					},
				},
			},
			"health_check_https": {
				Type:          schema.TypeList,
				MaxItems:      1,
				ConflictsWith: []string{"health_check_tcp", "health_check_http"},
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The HTTPS endpoint URL to call for HC requests",
						},
						"method": {
							Type:        schema.TypeString,
							Default:     "GET",
							Optional:    true,
							Description: "The HTTP method to use for HC requests",
						},
						"code": {
							Type:        schema.TypeInt,
							Default:     200,
							Optional:    true,
							Description: "The expected HTTP status code",
						},
						"host_header": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The HTTP host header to use for HC requests",
						},
						"sni": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The SNI to use for HC requests over SSL",
						},
					},
				},
			},
			"health_check_transient_delay": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "0.5s",
				ValidateFunc:     validateDuration(),
				DiffSuppressFunc: diffSuppressFuncDuration,
				Description:      "Time to wait between two consecutive health checks when a backend server is in a transient state (going UP or DOWN)",
			},
			"health_check_send_proxy": {
				Type:        schema.TypeBool,
				Description: "Defines whether proxy protocol should be activated for the health check",
				Optional:    true,
				Default:     false,
			},
			"on_marked_down_action": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"none",
					lbSDK.OnMarkedDownActionShutdownSessions.String(),
				}, false),
				Default:     "none",
				Optional:    true,
				Description: "Modify what occurs when a backend server is marked down",
			},
			"failover_host": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `Scaleway S3 bucket website to be served in case all backend servers are down

**NOTE** : Only the host part of the Scaleway S3 bucket website is expected.
E.g. 'failover-website.s3-website.fr-par.scw.cloud' if your bucket website URL is 'https://failover-website.s3-website.fr-par.scw.cloud/'.`,
			},
			"ssl_bridging": {
				Type:        schema.TypeBool,
				Description: "Enables SSL between load balancer and backend servers",
				Optional:    true,
				Default:     false,
			},
			"ignore_ssl_server_verify": {
				Type:        schema.TypeBool,
				Description: "Specifies whether the Load Balancer should check the backend server’s certificate before initiating a connection",
				Optional:    true,
				Default:     false,
			},
			"max_connections": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, math.MaxInt32),
				Description:  "Maximum number of connections allowed per backend server",
			},
			"timeout_queue": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "0s",
				ValidateFunc:     validateDuration(),
				DiffSuppressFunc: diffSuppressFuncDuration,
				Description:      "Maximum time (in seconds) for a request to be left pending in queue when `max_connections` is reached",
			},
			"redispatch_attempt_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, math.MaxInt32),
				Description:  "Whether to use another backend server on each attempt",
			},
			"max_retries": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntBetween(0, math.MaxInt32),
				Description:  "Number of retries when a backend server connection failed",
			},
		},
	}
}

func resourceScalewayLbBackendCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, _, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	// parse lb_id. It will be forced to a zoned lb
	zone, lbID, err := parseZonedID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	healthCheckPort := d.Get("health_check_port").(int)
	if healthCheckPort == 0 {
		healthCheckPort = d.Get("forward_port").(int)
	}

	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	healthCheckoutTimeout, err := expandDuration(d.Get("health_check_timeout"))
	if err != nil {
		return diag.FromErr(err)
	}
	healthCheckDelay, err := expandDuration(d.Get("health_check_delay"))
	if err != nil {
		return diag.FromErr(err)
	}
	timeoutServer, err := expandDuration(d.Get("timeout_server"))
	if err != nil {
		return diag.FromErr(err)
	}
	timeoutConnect, err := expandDuration(d.Get("timeout_connect"))
	if err != nil {
		return diag.FromErr(err)
	}
	timeoutTunnel, err := expandDuration(d.Get("timeout_tunnel"))
	if err != nil {
		return diag.FromErr(err)
	}
	createReq := &lbSDK.ZonedAPICreateBackendRequest{
		Zone:                     zone,
		LBID:                     lbID,
		Name:                     expandOrGenerateString(d.Get("name"), "lb-bkd"),
		ForwardProtocol:          expandLbProtocol(d.Get("forward_protocol")),
		ForwardPort:              int32(d.Get("forward_port").(int)),
		ForwardPortAlgorithm:     expandLbForwardPortAlgorithm(d.Get("forward_port_algorithm")),
		StickySessions:           expandLbStickySessionsType(d.Get("sticky_sessions")),
		StickySessionsCookieName: d.Get("sticky_sessions_cookie_name").(string),
		HealthCheck: &lbSDK.HealthCheck{
			Port:            int32(healthCheckPort),
			CheckMaxRetries: int32(d.Get("health_check_max_retries").(int)),
			CheckTimeout:    healthCheckoutTimeout,
			CheckDelay:      healthCheckDelay,
			TCPConfig:       expandLbHCTCP(d.Get("health_check_tcp")),
			HTTPConfig:      expandLbHCHTTP(d.Get("health_check_http")),
			HTTPSConfig:     expandLbHCHTTPS(d.Get("health_check_https")),
			CheckSendProxy:  d.Get("health_check_send_proxy").(bool),
		},
		ServerIP:              expandStrings(d.Get("server_ips")),
		ProxyProtocol:         expandLbProxyProtocol(d.Get("proxy_protocol")),
		TimeoutServer:         timeoutServer,
		TimeoutConnect:        timeoutConnect,
		TimeoutTunnel:         timeoutTunnel,
		OnMarkedDownAction:    expandLbBackendMarkdownAction(d.Get("on_marked_down_action")),
		FailoverHost:          expandStringPtr(d.Get("failover_host")),
		SslBridging:           expandBoolPtr(getBool(d, "ssl_bridging")),
		IgnoreSslServerVerify: expandBoolPtr(getBool(d, "ignore_ssl_server_verify")),
	}

	if maxConn, ok := d.GetOk("max_connections"); ok {
		createReq.MaxConnections = expandInt32Ptr(maxConn)
	}
	if timeoutQueue, ok := d.GetOk("timeout_queue"); ok {
		timeout, err := time.ParseDuration(timeoutQueue.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		createReq.TimeoutQueue = &scw.Duration{Seconds: int64(timeout.Seconds())}
	}
	if redispatchAttemptCount, ok := d.GetOk("redispatch_attempt_count"); ok {
		createReq.RedispatchAttemptCount = expandInt32Ptr(redispatchAttemptCount)
	}
	if maxRetries, ok := d.GetOk("max_retries"); ok {
		createReq.MaxRetries = expandInt32Ptr(maxRetries)
	}
	if healthCheckTransientDelay, ok := d.GetOk("health_check_transient_delay"); ok {
		timeout, err := time.ParseDuration(healthCheckTransientDelay.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		createReq.HealthCheck.TransientCheckDelay = &scw.Duration{Seconds: int64(timeout.Seconds()), Nanos: int32(timeout.Nanoseconds())}
	}

	// deprecated attribute
	createReq.SendProxyV2 = expandBoolPtr(getBool(d, "send_proxy_v2"))

	res, err := lbAPI.CreateBackend(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForLB(ctx, lbAPI, zone, res.LB.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))

	return resourceScalewayLbBackendRead(ctx, d, meta)
}

func resourceScalewayLbBackendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	backend, err := lbAPI.GetBackend(&lbSDK.ZonedAPIGetBackendRequest{
		Zone:      zone,
		BackendID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("lb_id", newZonedIDString(zone, backend.LB.ID))
	_ = d.Set("name", backend.Name)
	_ = d.Set("forward_protocol", flattenLbProtocol(backend.ForwardProtocol))
	_ = d.Set("forward_port", backend.ForwardPort)
	_ = d.Set("forward_port_algorithm", flattenLbForwardPortAlgorithm(backend.ForwardPortAlgorithm))
	_ = d.Set("sticky_sessions", flattenLbStickySessionsType(backend.StickySessions))
	_ = d.Set("sticky_sessions_cookie_name", backend.StickySessionsCookieName)
	_ = d.Set("server_ips", backend.Pool)
	_ = d.Set("proxy_protocol", flattenLbProxyProtocol(backend.ProxyProtocol))
	_ = d.Set("timeout_server", flattenDuration(backend.TimeoutServer))
	_ = d.Set("timeout_connect", flattenDuration(backend.TimeoutConnect))
	_ = d.Set("timeout_tunnel", flattenDuration(backend.TimeoutTunnel))
	_ = d.Set("on_marked_down_action", flattenLbBackendMarkdownAction(backend.OnMarkedDownAction))
	_ = d.Set("send_proxy_v2", flattenBoolPtr(backend.SendProxyV2))
	_ = d.Set("failover_host", backend.FailoverHost)
	_ = d.Set("ssl_bridging", flattenBoolPtr(backend.SslBridging))
	_ = d.Set("ignore_ssl_server_verify", flattenBoolPtr(backend.IgnoreSslServerVerify))
	_ = d.Set("max_connections", flattenInt32Ptr(backend.MaxConnections))
	_ = d.Set("redispatch_attempt_count", flattenInt32Ptr(backend.RedispatchAttemptCount))
	_ = d.Set("max_retries", flattenInt32Ptr(backend.MaxRetries))
	_ = d.Set("timeout_queue", flattenDuration(backend.TimeoutQueue.ToTimeDuration()))

	// HealthCheck
	_ = d.Set("health_check_port", backend.HealthCheck.Port)
	_ = d.Set("health_check_max_retries", backend.HealthCheck.CheckMaxRetries)
	_ = d.Set("health_check_timeout", flattenDuration(backend.HealthCheck.CheckTimeout))
	_ = d.Set("health_check_delay", flattenDuration(backend.HealthCheck.CheckDelay))
	_ = d.Set("health_check_tcp", flattenLbHCTCP(backend.HealthCheck.TCPConfig))
	_ = d.Set("health_check_http", flattenLbHCHTTP(backend.HealthCheck.HTTPConfig))
	_ = d.Set("health_check_https", flattenLbHCHTTPS(backend.HealthCheck.HTTPSConfig))
	_ = d.Set("health_check_transient_delay", flattenDuration(backend.HealthCheck.TransientCheckDelay.ToTimeDuration()))
	_ = d.Set("health_check_send_proxy", backend.HealthCheck.CheckSendProxy)

	_, err = waitForLB(ctx, lbAPI, zone, backend.LB.ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

//gocyclo:ignore
func resourceScalewayLbBackendUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, lbID, err := parseZonedID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	timeoutServer, err := expandDuration(d.Get("timeout_server"))
	if err != nil {
		return diag.FromErr(err)
	}
	timeoutConnect, err := expandDuration(d.Get("timeout_connect"))
	if err != nil {
		return diag.FromErr(err)
	}
	timeoutTunnel, err := expandDuration(d.Get("timeout_tunnel"))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &lbSDK.ZonedAPIUpdateBackendRequest{
		Zone:                     zone,
		BackendID:                ID,
		Name:                     d.Get("name").(string),
		ForwardProtocol:          expandLbProtocol(d.Get("forward_protocol")),
		ForwardPort:              int32(d.Get("forward_port").(int)),
		ForwardPortAlgorithm:     expandLbForwardPortAlgorithm(d.Get("forward_port_algorithm")),
		StickySessions:           expandLbStickySessionsType(d.Get("sticky_sessions")),
		StickySessionsCookieName: d.Get("sticky_sessions_cookie_name").(string),
		ProxyProtocol:            expandLbProxyProtocol(d.Get("proxy_protocol")),
		TimeoutServer:            timeoutServer,
		TimeoutConnect:           timeoutConnect,
		TimeoutTunnel:            timeoutTunnel,
		OnMarkedDownAction:       expandLbBackendMarkdownAction(d.Get("on_marked_down_action")),
		FailoverHost:             expandStringPtr(d.Get("failover_host")),
		SslBridging:              expandBoolPtr(getBool(d, "ssl_bridging")),
		IgnoreSslServerVerify:    expandBoolPtr(getBool(d, "ignore_ssl_server_verify")),
		MaxConnections:           expandInt32Ptr(d.Get("max_connections")),
		RedispatchAttemptCount:   expandInt32Ptr(d.Get("redispatch_attempt_count")),
		MaxRetries:               expandInt32Ptr(d.Get("max_retries")),
	}

	if timeoutQueue, ok := d.GetOk("timeout_queue"); ok {
		timeoutQueueParsed, err := time.ParseDuration(timeoutQueue.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		req.TimeoutQueue = &scw.Duration{Seconds: int64(timeoutQueueParsed.Seconds())}
	}

	// deprecated
	req.SendProxyV2 = expandBoolPtr(getBool(d, "send_proxy_v2"))

	_, err = lbAPI.UpdateBackend(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	healthCheckoutTimeout, err := expandDuration(d.Get("health_check_timeout"))
	if err != nil {
		return diag.FromErr(err)
	}
	healthCheckDelay, err := expandDuration(d.Get("health_check_delay"))
	if err != nil {
		return diag.FromErr(err)
	}
	// Update Health Check
	updateHCRequest := &lbSDK.ZonedAPIUpdateHealthCheckRequest{
		Zone:            zone,
		BackendID:       ID,
		Port:            int32(d.Get("health_check_port").(int)),
		CheckMaxRetries: int32(d.Get("health_check_max_retries").(int)),
		CheckTimeout:    healthCheckoutTimeout,
		CheckDelay:      healthCheckDelay,
		HTTPConfig:      expandLbHCHTTP(d.Get("health_check_http")),
		HTTPSConfig:     expandLbHCHTTPS(d.Get("health_check_https")),
		CheckSendProxy:  d.Get("health_check_send_proxy").(bool),
	}
	if healthCheckTransientDelay, ok := d.GetOk("health_check_transient_delay"); ok {
		timeout, err := time.ParseDuration(healthCheckTransientDelay.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		updateHCRequest.TransientCheckDelay = &scw.Duration{Seconds: int64(timeout.Seconds()), Nanos: int32(timeout.Nanoseconds())}
	}

	// As this is the default behaviour if no other HC type are present we enable TCP
	if updateHCRequest.HTTPConfig == nil && updateHCRequest.HTTPSConfig == nil {
		updateHCRequest.TCPConfig = expandLbHCTCP(d.Get("health_check_tcp"))
	}

	rawConfig := d.GetRawConfig().AsValueMap()
	healthCheckPortValue, healthCheckPortExists := rawConfig["health_check_port"]
	healthCheckPortSetByUser := healthCheckPortExists && !healthCheckPortValue.IsNull()
	if d.HasChange("forward_port") && !healthCheckPortSetByUser {
		updateHCRequest.Port = int32(d.Get("forward_port").(int))
	}

	_, err = lbAPI.UpdateHealthCheck(updateHCRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// Update Backend servers
	_, err = lbAPI.SetBackendServers(&lbSDK.ZonedAPISetBackendServersRequest{
		Zone:      zone,
		BackendID: ID,
		ServerIP:  expandStrings(d.Get("server_ips")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return resourceScalewayLbBackendRead(ctx, d, meta)
}

func resourceScalewayLbBackendDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, lbID, err := parseZonedID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	err = lbAPI.DeleteBackend(&lbSDK.ZonedAPIDeleteBackendRequest{
		Zone:      zone,
		BackendID: ID,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
