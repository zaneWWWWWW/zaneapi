package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/service/authz"

	// Import oauth package to register providers via init()
	_ "github.com/QuantumNous/new-api/oauth"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetApiRouter(router *gin.Engine) {
	apiRouter := router.Group("/api")
	apiRouter.Use(middleware.RouteTag("api"))
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.BodyStorageCleanup()) // 清理请求体存储
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	anonymousRequestBodyLimit := middleware.AnonymousRequestBodyLimit()
	{
		apiRouter.GET("/setup", controller.GetSetup)
		apiRouter.POST("/setup", anonymousRequestBodyLimit, controller.PostSetup)
		apiRouter.GET("/status", controller.GetStatus)
		apiRouter.GET("/uptime/status", middleware.AdminAuth(), middleware.RequirePermission(authz.ChannelRead), controller.GetUptimeKumaStatus)
		apiRouter.GET("/models", middleware.UserAuth(), controller.DashboardListModels)
		apiRouter.GET("/status/test", middleware.AdminAuth(), controller.TestStatus)
		apiRouter.GET("/notice", controller.GetNotice)
		apiRouter.GET("/user-agreement", controller.GetUserAgreement)
		apiRouter.GET("/privacy-policy", controller.GetPrivacyPolicy)
		apiRouter.GET("/about", controller.GetAbout)
		apiRouter.GET("/docs", controller.GetDocs)
		//apiRouter.GET("/midjourney", controller.GetMidjourney)
		apiRouter.GET("/home_page_content", controller.GetHomePageContent)
		apiRouter.GET("/pricing", middleware.HeaderNavModuleAuth("pricing"), controller.GetPricing)
		perfMetricsRoute := apiRouter.Group("/perf-metrics")
		perfMetricsRoute.Use(middleware.HeaderNavModulePublicOrUserAuth("pricing"))
		{
			perfMetricsRoute.GET("/summary", controller.GetPerfMetricsSummary)
			perfMetricsRoute.GET("/groups", middleware.UserAuth(), controller.GetPerfMetricsGroups)
			perfMetricsRoute.GET("", controller.GetPerfMetrics)
		}
		apiRouter.GET("/rankings", middleware.HeaderNavModuleAuth("rankings"), controller.GetRankings)
		apiRouter.GET("/verification", middleware.EmailVerificationRateLimit(), middleware.TurnstileCheck(), controller.SendEmailVerification)
		apiRouter.GET("/reset_password", middleware.AuthRateLimit(), middleware.TurnstileCheck(), controller.SendPasswordResetEmail)
		apiRouter.POST("/user/reset", middleware.AuthRateLimit(), anonymousRequestBodyLimit, controller.ResetPassword)
		// OAuth routes - specific routes must come before :provider wildcard
		apiRouter.POST("/oauth/state", middleware.AuthRateLimit(), middleware.DisableCache(), middleware.TryUserAuth(), anonymousRequestBodyLimit, controller.GenerateOAuthCode)
		apiRouter.POST("/oauth/email/bind", middleware.UserAuth(), middleware.CriticalRateLimit(), controller.EmailBind)
		// Non-standard OAuth (WeChat, Telegram) - keep original routes
		apiRouter.GET("/oauth/wechat", middleware.AuthRateLimit(), middleware.DisableCache(), controller.WeChatAuth)
		apiRouter.POST("/oauth/wechat/bind", middleware.UserAuth(), middleware.CriticalRateLimit(), controller.WeChatBind)
		apiRouter.GET("/oauth/telegram/login", middleware.AuthRateLimit(), middleware.DisableCache(), controller.TelegramLogin)
		apiRouter.POST("/oauth/telegram/bind/start", middleware.UserAuth(), middleware.CriticalRateLimit(), middleware.DisableCache(), controller.TelegramBindStart)
		apiRouter.GET("/oauth/telegram/bind/:flow_token", middleware.CriticalRateLimit(), middleware.DisableCache(), controller.TelegramBind)
		// Standard OAuth providers (GitHub, Discord, OIDC, LinuxDO) - unified route
		apiRouter.GET("/oauth/:provider", middleware.AuthRateLimit(), middleware.DisableCache(), middleware.TryUserAuth(), controller.HandleOAuth)
		apiRouter.GET("/ratio_config", middleware.CriticalRateLimit(), controller.GetRatioConfig)

		apiRouter.POST("/stripe/webhook", anonymousRequestBodyLimit, controller.StripeWebhook)
		apiRouter.POST("/creem/webhook", anonymousRequestBodyLimit, controller.CreemWebhook)
		apiRouter.POST("/waffo/webhook", anonymousRequestBodyLimit, controller.WaffoWebhook)
		// :env separates test vs prod URLs so the operator can register each
		// in Pancake's matching webhook slot; handler enforces env match.
		apiRouter.POST("/waffo-pancake/webhook/:env", anonymousRequestBodyLimit, controller.WaffoPancakeWebhook)

		// Universal secure verification routes
		apiRouter.POST("/verify", middleware.UserAuth(), middleware.CriticalRateLimit(), middleware.DisableCache(), controller.UniversalVerify)

		userRoute := apiRouter.Group("/user")
		{
			userRoute.POST("/auth/refresh", middleware.SessionCookieOriginGuard(), middleware.CriticalRateLimit(), middleware.DisableCache(), controller.RefreshAuth)
			userRoute.POST("/auth/logout", middleware.SessionCookieOriginGuard(), middleware.CriticalRateLimit(), middleware.DisableCache(), controller.AuthLogout)
			userRoute.POST("/register", middleware.AuthRateLimit(), anonymousRequestBodyLimit, middleware.TurnstileCheck(), controller.Register)
			userRoute.POST("/login", middleware.AuthRateLimit(), middleware.DisableCache(), anonymousRequestBodyLimit, middleware.TurnstileCheck(), controller.Login)
			userRoute.POST("/login/2fa", middleware.AuthRateLimit(), middleware.DisableCache(), anonymousRequestBodyLimit, controller.Verify2FALogin)
			userRoute.POST("/passkey/login/begin", middleware.AuthRateLimit(), middleware.DisableCache(), anonymousRequestBodyLimit, controller.PasskeyLoginBegin)
			userRoute.POST("/passkey/login/finish", middleware.AuthRateLimit(), middleware.DisableCache(), anonymousRequestBodyLimit, controller.PasskeyLoginFinish)
			//userRoute.POST("/tokenlog", middleware.CriticalRateLimit(), controller.TokenLog)
			userRoute.POST("/epay/notify", anonymousRequestBodyLimit, controller.EpayNotify)
			userRoute.GET("/epay/notify", controller.EpayNotify)
			userRoute.POST("/hupijiao/notify", anonymousRequestBodyLimit, controller.HupijiaoNotify)
			userRoute.GET("/groups", controller.GetUserGroups)

			selfRoute := userRoute.Group("/")
			selfRoute.Use(middleware.UserAuth())
			{
				selfRoute.GET("/sessions", middleware.DisableCache(), controller.GetLoginSessions)
				selfRoute.DELETE("/sessions/:sid", middleware.DisableCache(), controller.DeleteLoginSession)
				selfRoute.POST("/sessions/revoke-others", middleware.DisableCache(), controller.RevokeOtherLoginSessions)
				selfRoute.GET("/self/groups", controller.GetUserGroups)
				selfRoute.GET("/self", controller.GetSelf)
				selfRoute.GET("/models", controller.GetUserModels)
				selfRoute.PUT("/self", middleware.CriticalRateLimit(), middleware.DisableCache(), controller.UpdateSelf)
				selfRoute.DELETE("/self", controller.DeleteSelf)
				selfRoute.GET("/token", middleware.DisableCache(), controller.GenerateAccessToken)
				selfRoute.GET("/passkey", controller.PasskeyStatus)
				selfRoute.POST("/passkey/register/begin", middleware.DisableCache(), controller.PasskeyRegisterBegin)
				selfRoute.POST("/passkey/register/finish", middleware.DisableCache(), controller.PasskeyRegisterFinish)
				selfRoute.POST("/passkey/verify/begin", middleware.DisableCache(), controller.PasskeyVerifyBegin)
				selfRoute.POST("/passkey/verify/finish", middleware.DisableCache(), controller.PasskeyVerifyFinish)
				selfRoute.DELETE("/passkey", middleware.DisableCache(), controller.PasskeyDelete)
				selfRoute.GET("/aff", controller.GetAffCode)
				selfRoute.GET("/topup/info", controller.GetTopUpInfo)
				selfRoute.GET("/topup/self", controller.GetUserTopUps)
				selfRoute.POST("/topup", middleware.CriticalRateLimit(), controller.TopUp)
				selfRoute.POST("/pay", middleware.CriticalRateLimit(), controller.RequestEpay)
				selfRoute.POST("/amount", controller.RequestAmount)
				selfRoute.POST("/stripe/pay", middleware.CriticalRateLimit(), controller.RequestStripePay)
				selfRoute.POST("/stripe/amount", controller.RequestStripeAmount)
				selfRoute.POST("/creem/pay", middleware.CriticalRateLimit(), controller.RequestCreemPay)
				selfRoute.POST("/waffo/amount", controller.RequestWaffoAmount)
				selfRoute.POST("/waffo/pay", middleware.CriticalRateLimit(), controller.RequestWaffoPay)
				selfRoute.POST("/waffo-pancake/amount", controller.RequestWaffoPancakeAmount)
				selfRoute.POST("/waffo-pancake/pay", middleware.CriticalRateLimit(), controller.RequestWaffoPancakePay)
				selfRoute.POST("/hupijiao/pay", middleware.CriticalRateLimit(), controller.RequestHupijiaoPay)
				selfRoute.POST("/aff_transfer", controller.TransferAffQuota)
				selfRoute.PUT("/setting", controller.UpdateUserSetting)

				// 2FA routes
				selfRoute.GET("/2fa/status", controller.Get2FAStatus)
				selfRoute.POST("/2fa/setup", middleware.DisableCache(), controller.Setup2FA)
				selfRoute.POST("/2fa/enable", middleware.DisableCache(), controller.Enable2FA)
				selfRoute.POST("/2fa/disable", middleware.DisableCache(), controller.Disable2FA)
				selfRoute.POST("/2fa/backup_codes", middleware.DisableCache(), controller.RegenerateBackupCodes)

				// Check-in routes
				selfRoute.GET("/checkin", controller.GetCheckinStatus)
				selfRoute.POST("/checkin", middleware.TurnstileCheck(), controller.DoCheckin)

				// Custom OAuth bindings
				selfRoute.GET("/oauth/bindings", controller.GetUserOAuthBindings)
				selfRoute.DELETE("/oauth/bindings/:provider_id", controller.UnbindCustomOAuth)
			}

			adminRoute := userRoute.Group("/")
			adminRoute.Use(middleware.AdminAuth())
			{
				adminRoute.GET("/", middleware.RequirePermission(authz.UsersRead), controller.GetAllUsers)
				adminRoute.GET("/topup", middleware.RequirePermission(authz.UsersRead), controller.GetAllTopUps)
				adminRoute.POST("/topup/complete", middleware.RequirePermission(authz.UsersWrite), controller.AdminCompleteTopUp)
				adminRoute.GET("/search", middleware.RequirePermission(authz.UsersRead), controller.SearchUsers)
				adminRoute.GET("/:id/oauth/bindings", middleware.RequirePermission(authz.UsersRead), controller.GetUserOAuthBindingsByAdmin)
				adminRoute.DELETE("/:id/oauth/bindings/:provider_id", middleware.RequirePermission(authz.UsersWrite), controller.UnbindCustomOAuthByAdmin)
				adminRoute.DELETE("/:id/bindings/:binding_type", middleware.RequirePermission(authz.UsersWrite), controller.AdminClearUserBinding)
				adminRoute.GET("/:id", middleware.RequirePermission(authz.UsersRead), controller.GetUser)
				adminRoute.POST("/", middleware.RequirePermission(authz.UsersWrite), controller.CreateUser)
				adminRoute.POST("/manage", middleware.RequirePermission(authz.UsersWrite), controller.ManageUser)
				adminRoute.PUT("/", middleware.RequirePermission(authz.UsersWrite), controller.UpdateUser)
				adminRoute.DELETE("/:id", middleware.RequirePermission(authz.UsersWrite), controller.DeleteUser)
				adminRoute.DELETE("/:id/reset_passkey", middleware.RequirePermission(authz.UsersWrite), controller.AdminResetPasskey)

				// Admin 2FA routes
				adminRoute.GET("/2fa/stats", middleware.RequirePermission(authz.UsersRead), controller.Admin2FAStats)
				adminRoute.DELETE("/:id/2fa", middleware.RequirePermission(authz.UsersWrite), controller.AdminDisable2FA)
			}
		}

		// Subscription billing (plans, purchase, admin management)
		subscriptionRoute := apiRouter.Group("/subscription")
		subscriptionRoute.Use(middleware.UserAuth())
		{
			subscriptionRoute.GET("/plans", controller.GetSubscriptionPlans)
			subscriptionRoute.GET("/self", controller.GetSubscriptionSelf)
			subscriptionRoute.PUT("/self/preference", controller.UpdateSubscriptionPreference)
			subscriptionRoute.POST("/balance/pay", middleware.CriticalRateLimit(), controller.SubscriptionRequestBalancePay)
			subscriptionRoute.POST("/epay/pay", middleware.CriticalRateLimit(), controller.SubscriptionRequestEpay)
			subscriptionRoute.POST("/stripe/pay", middleware.CriticalRateLimit(), controller.SubscriptionRequestStripePay)
			subscriptionRoute.POST("/creem/pay", middleware.CriticalRateLimit(), controller.SubscriptionRequestCreemPay)
			subscriptionRoute.POST("/waffo-pancake/pay", middleware.CriticalRateLimit(), controller.SubscriptionRequestWaffoPancakePay)
		}
		subscriptionAdminRoute := apiRouter.Group("/subscription/admin")
		subscriptionAdminRoute.Use(middleware.AdminAuth())
		{
			subscriptionAdminRoute.GET("/plans", middleware.RequirePermission(authz.SubscriptionsRead), controller.AdminListSubscriptionPlans)
			subscriptionAdminRoute.POST("/plans", middleware.RequirePermission(authz.SubscriptionsWrite), controller.AdminCreateSubscriptionPlan)
			subscriptionAdminRoute.PUT("/plans/:id", middleware.RequirePermission(authz.SubscriptionsWrite), controller.AdminUpdateSubscriptionPlan)
			subscriptionAdminRoute.PATCH("/plans/:id", middleware.RequirePermission(authz.SubscriptionsWrite), controller.AdminUpdateSubscriptionPlanStatus)
			subscriptionAdminRoute.POST("/bind", middleware.RequirePermission(authz.SubscriptionsWrite), controller.AdminBindSubscription)
			subscriptionAdminRoute.POST("/plans/:id/subscriptions/reset", middleware.RequirePermission(authz.SubscriptionsWrite), controller.AdminResetPlanSubscriptions)

			// User subscription management (admin)
			subscriptionAdminRoute.GET("/users/:id/subscriptions", middleware.RequirePermission(authz.SubscriptionsRead), controller.AdminListUserSubscriptions)
			subscriptionAdminRoute.POST("/users/:id/subscriptions", middleware.RequirePermission(authz.SubscriptionsWrite), controller.AdminCreateUserSubscription)
			subscriptionAdminRoute.POST("/users/:id/subscriptions/reset", middleware.RequirePermission(authz.SubscriptionsWrite), controller.AdminResetUserSubscriptionsByPlan)
			subscriptionAdminRoute.POST("/user_subscriptions/:id/invalidate", middleware.RequirePermission(authz.SubscriptionsWrite), controller.AdminInvalidateUserSubscription)
			subscriptionAdminRoute.DELETE("/user_subscriptions/:id", middleware.RequirePermission(authz.SubscriptionsWrite), controller.AdminDeleteUserSubscription)
		}

		// Subscription payment callbacks (no auth)
		apiRouter.POST("/subscription/epay/notify", anonymousRequestBodyLimit, controller.SubscriptionEpayNotify)
		apiRouter.GET("/subscription/epay/notify", controller.SubscriptionEpayNotify)
		apiRouter.GET("/subscription/epay/return", controller.SubscriptionEpayReturn)
		apiRouter.POST("/subscription/epay/return", anonymousRequestBodyLimit, controller.SubscriptionEpayReturn)
		optionRoute := apiRouter.Group("/option")
		optionRoute.Use(middleware.RootAuth())
		{
			optionRoute.GET("/", controller.GetOptions)
			optionRoute.PUT("/", controller.UpdateOption)
			optionRoute.POST("/payment_compliance", controller.ConfirmPaymentCompliance)
			optionRoute.GET("/channel_affinity_cache", controller.GetChannelAffinityCacheStats)
			optionRoute.DELETE("/channel_affinity_cache", controller.ClearChannelAffinityCache)
			optionRoute.POST("/rest_model_ratio", controller.ResetModelRatio)
			optionRoute.GET("/waffo-pancake/catalog", controller.ListWaffoPancakeCatalog)
			optionRoute.POST("/waffo-pancake/pair", controller.CreateWaffoPancakePair)
			optionRoute.POST("/waffo-pancake/save", controller.SaveWaffoPancake)
			optionRoute.POST("/waffo-pancake/subscription-product", controller.CreateWaffoPancakeSubscriptionProduct)
			optionRoute.GET("/waffo-pancake/subscription-product-options", controller.ListWaffoPancakeSubscriptionProductOptions)
		}

		// Custom OAuth provider management (root only)
		customOAuthRoute := apiRouter.Group("/custom-oauth-provider")
		customOAuthRoute.Use(middleware.RootAuth())
		{
			customOAuthRoute.POST("/discovery", controller.FetchCustomOAuthDiscovery)
			customOAuthRoute.GET("/", controller.GetCustomOAuthProviders)
			customOAuthRoute.GET("/:id", controller.GetCustomOAuthProvider)
			customOAuthRoute.POST("/", controller.CreateCustomOAuthProvider)
			customOAuthRoute.PUT("/:id", controller.UpdateCustomOAuthProvider)
			customOAuthRoute.DELETE("/:id", controller.DeleteCustomOAuthProvider)
		}
		performanceRoute := apiRouter.Group("/performance")
		performanceRoute.Use(middleware.RootAuth())
		{
			performanceRoute.GET("/stats", controller.GetPerformanceStats)
			performanceRoute.DELETE("/disk_cache", controller.ClearDiskCache)
			performanceRoute.POST("/reset_stats", controller.ResetPerformanceStats)
			performanceRoute.POST("/gc", controller.ForceGC)
			performanceRoute.GET("/logs", controller.GetLogFiles)
			performanceRoute.DELETE("/logs", controller.CleanupLogFiles)
		}
		ratioSyncRoute := apiRouter.Group("/ratio_sync")
		ratioSyncRoute.Use(middleware.RootAuth())
		{
			ratioSyncRoute.GET("/channels", controller.GetSyncableChannels)
			ratioSyncRoute.POST("/fetch", controller.FetchUpstreamRatios)
			ratioSyncRoute.POST("/apply_extras", controller.ApplyPriceAppleExtras)
		}
		registerChannelRoutes(apiRouter)
		registerAuthzRoutes(apiRouter)
		tokenRoute := apiRouter.Group("/token")
		tokenRoute.Use(middleware.UserAuth())
		{
			tokenRoute.GET("/", controller.GetAllTokens)
			tokenRoute.GET("/search", middleware.SearchRateLimit(), controller.SearchTokens)
			tokenRoute.GET("/:id", controller.GetToken)
			tokenRoute.POST("/:id/key", middleware.CriticalRateLimit(), middleware.DisableCache(), controller.GetTokenKey)
			tokenRoute.POST("/", controller.AddToken)
			tokenRoute.PUT("/", controller.UpdateToken)
			tokenRoute.DELETE("/:id", controller.DeleteToken)
			tokenRoute.POST("/batch", controller.DeleteTokenBatch)
			tokenRoute.POST("/batch/keys", middleware.CriticalRateLimit(), middleware.DisableCache(), controller.GetTokenKeysBatch)
		}

		usageRoute := apiRouter.Group("/usage")
		usageRoute.Use(middleware.CORS(), middleware.CriticalRateLimit())
		{
			tokenUsageRoute := usageRoute.Group("/token")
			tokenUsageRoute.Use(middleware.TokenAuthReadOnly())
			{
				tokenUsageRoute.GET("/", controller.GetTokenUsage)
			}
		}

		redemptionRoute := apiRouter.Group("/redemption")
		redemptionRoute.Use(middleware.AdminAuth())
		{
			redemptionRoute.GET("/", middleware.RequirePermission(authz.RedemptionRead), controller.GetAllRedemptions)
			redemptionRoute.GET("/search", middleware.RequirePermission(authz.RedemptionRead), controller.SearchRedemptions)
			redemptionRoute.GET("/:id", middleware.RequirePermission(authz.RedemptionRead), controller.GetRedemption)
			redemptionRoute.POST("/", middleware.RequirePermission(authz.RedemptionWrite), controller.AddRedemption)
			redemptionRoute.PUT("/", middleware.RequirePermission(authz.RedemptionWrite), controller.UpdateRedemption)
			redemptionRoute.DELETE("/invalid", middleware.RequirePermission(authz.RedemptionWrite), controller.DeleteInvalidRedemption)
			redemptionRoute.DELETE("/:id", middleware.RequirePermission(authz.RedemptionWrite), controller.DeleteRedemption)
		}
		logRoute := apiRouter.Group("/log")
		logRoute.GET("/", middleware.AdminAuth(), middleware.RequirePermission(authz.LogsRead), controller.GetAllLogs)
		logRoute.GET("/stat", middleware.AdminAuth(), middleware.RequirePermission(authz.LogsRead), controller.GetLogsStat)
		logRoute.GET("/self/stat", middleware.UserAuth(), controller.GetLogsSelfStat)
		logRoute.GET("/channel_affinity_usage_cache", middleware.AdminAuth(), middleware.RequirePermission(authz.LogsRead), controller.GetChannelAffinityUsageCacheStats)
		logRoute.GET("/search", middleware.AdminAuth(), middleware.RequirePermission(authz.LogsRead), controller.SearchAllLogs)
		logRoute.GET("/self", middleware.UserAuth(), controller.GetUserLogs)
		logRoute.GET("/self/search", middleware.UserAuth(), middleware.SearchRateLimit(), controller.SearchUserLogs)

		systemTaskRoute := apiRouter.Group("/system-task")
		systemTaskRoute.Use(middleware.RootAuth())
		{
			systemTaskRoute.POST("/log-cleanup", controller.CreateLogCleanupSystemTask)
			systemTaskRoute.GET("/list", controller.ListSystemTasks)
			systemTaskRoute.GET("/current", controller.GetCurrentSystemTask)
			systemTaskRoute.GET("/:task_id", controller.GetSystemTask)
		}
		systemInfoRoute := apiRouter.Group("/system-info")
		systemInfoRoute.Use(middleware.RootAuth())
		{
			systemInfoRoute.GET("/instances", controller.ListSystemInstances)
			systemInfoRoute.DELETE("/stale-instances", controller.DeleteStaleSystemInstances)
			systemInfoRoute.DELETE("/instances/:node_name", controller.DeleteStaleSystemInstance)
		}

		dataRoute := apiRouter.Group("/data")
		dataRoute.GET("/", middleware.AdminAuth(), middleware.RequirePermission(authz.DashboardRead), controller.GetAllQuotaDates)
		dataRoute.GET("/users", middleware.AdminAuth(), middleware.RequirePermission(authz.DashboardRead), controller.GetQuotaDatesByUser)
		dataRoute.GET("/self", middleware.UserAuth(), controller.GetUserQuotaDates)
		dataRoute.GET("/flow", middleware.AdminAuth(), middleware.RequirePermission(authz.DashboardRead), controller.GetAllFlowQuotaDates)
		dataRoute.GET("/flow/self", middleware.UserAuth(), controller.GetUserFlowQuotaDates)

		logRoute.Use(middleware.CORS(), middleware.CriticalRateLimit())
		{
			logRoute.GET("/token", middleware.TokenAuthReadOnly(), controller.GetLogByKey)
		}
		groupRoute := apiRouter.Group("/group")
		groupRoute.Use(middleware.AdminAuth())
		{
			groupRoute.GET("/", controller.GetGroups)
		}

		prefillGroupRoute := apiRouter.Group("/prefill_group")
		prefillGroupRoute.Use(middleware.AdminAuth())
		{
			prefillGroupRoute.GET("/", middleware.RequirePermission(authz.ModelsRead), controller.GetPrefillGroups)
			prefillGroupRoute.POST("/", middleware.RequirePermission(authz.ModelsWrite), controller.CreatePrefillGroup)
			prefillGroupRoute.PUT("/", middleware.RequirePermission(authz.ModelsWrite), controller.UpdatePrefillGroup)
			prefillGroupRoute.DELETE("/:id", middleware.RequirePermission(authz.ModelsWrite), controller.DeletePrefillGroup)
		}

		mjRoute := apiRouter.Group("/mj")
		mjRoute.GET("/self", middleware.UserAuth(), controller.GetUserMidjourney)
		mjRoute.GET("/", middleware.AdminAuth(), middleware.RequirePermission(authz.LogsRead), controller.GetAllMidjourney)

		taskRoute := apiRouter.Group("/task")
		{
			taskRoute.GET("/self", middleware.UserAuth(), controller.GetUserTask)
			taskRoute.GET("/", middleware.AdminAuth(), middleware.RequirePermission(authz.LogsRead), controller.GetAllTask)
		}

		vendorRoute := apiRouter.Group("/vendors")
		vendorRoute.Use(middleware.AdminAuth())
		{
			vendorRoute.GET("/", middleware.RequirePermission(authz.ModelsRead), controller.GetAllVendors)
			vendorRoute.GET("/search", middleware.RequirePermission(authz.ModelsRead), controller.SearchVendors)
			vendorRoute.GET("/:id", middleware.RequirePermission(authz.ModelsRead), controller.GetVendorMeta)
			vendorRoute.POST("/", middleware.RequirePermission(authz.ModelsWrite), controller.CreateVendorMeta)
			vendorRoute.PUT("/", middleware.RequirePermission(authz.ModelsWrite), controller.UpdateVendorMeta)
			vendorRoute.DELETE("/:id", middleware.RequirePermission(authz.ModelsWrite), controller.DeleteVendorMeta)
		}

		modelsRoute := apiRouter.Group("/models")
		modelsRoute.Use(middleware.AdminAuth())
		{
			modelsRoute.GET("/sync_upstream/preview", middleware.RequirePermission(authz.ModelsRead), controller.SyncUpstreamPreview)
			modelsRoute.POST("/sync_upstream", middleware.RequirePermission(authz.ModelsWrite), controller.SyncUpstreamModels)
			modelsRoute.GET("/missing", middleware.RequirePermission(authz.ModelsRead), controller.GetMissingModels)
			modelsRoute.GET("/", middleware.RequirePermission(authz.ModelsRead), controller.GetAllModelsMeta)
			modelsRoute.GET("/search", middleware.RequirePermission(authz.ModelsRead), controller.SearchModelsMeta)
			modelsRoute.GET("/:id", middleware.RequirePermission(authz.ModelsRead), controller.GetModelMeta)
			modelsRoute.POST("/", middleware.RequirePermission(authz.ModelsWrite), controller.CreateModelMeta)
			modelsRoute.PUT("/", middleware.RequirePermission(authz.ModelsWrite), controller.UpdateModelMeta)
			modelsRoute.DELETE("/:id", middleware.RequirePermission(authz.ModelsWrite), controller.DeleteModelMeta)
		}

		// Deployments (model deployment management)
		deploymentsRoute := apiRouter.Group("/deployments")
		deploymentsRoute.Use(middleware.AdminAuth())
		{
			deploymentsRoute.GET("/settings", middleware.RequirePermission(authz.ModelsRead), controller.GetModelDeploymentSettings)
			deploymentsRoute.POST("/settings/test-connection", middleware.RequirePermission(authz.ModelsWrite), controller.TestIoNetConnection)
			deploymentsRoute.GET("/", middleware.RequirePermission(authz.ModelsRead), controller.GetAllDeployments)
			deploymentsRoute.GET("/search", middleware.RequirePermission(authz.ModelsRead), controller.SearchDeployments)
			deploymentsRoute.POST("/test-connection", middleware.RequirePermission(authz.ModelsWrite), controller.TestIoNetConnection)
			deploymentsRoute.GET("/hardware-types", middleware.RequirePermission(authz.ModelsRead), controller.GetHardwareTypes)
			deploymentsRoute.GET("/locations", middleware.RequirePermission(authz.ModelsRead), controller.GetLocations)
			deploymentsRoute.GET("/available-replicas", middleware.RequirePermission(authz.ModelsRead), controller.GetAvailableReplicas)
			deploymentsRoute.POST("/price-estimation", middleware.RequirePermission(authz.ModelsRead), controller.GetPriceEstimation)
			deploymentsRoute.GET("/check-name", middleware.RequirePermission(authz.ModelsRead), controller.CheckClusterNameAvailability)
			deploymentsRoute.POST("/", middleware.RequirePermission(authz.ModelsWrite), controller.CreateDeployment)

			deploymentsRoute.GET("/:id", middleware.RequirePermission(authz.ModelsRead), controller.GetDeployment)
			deploymentsRoute.GET("/:id/logs", middleware.RequirePermission(authz.ModelsRead), controller.GetDeploymentLogs)
			deploymentsRoute.GET("/:id/containers", middleware.RequirePermission(authz.ModelsRead), controller.ListDeploymentContainers)
			deploymentsRoute.GET("/:id/containers/:container_id", middleware.RequirePermission(authz.ModelsRead), controller.GetContainerDetails)
			deploymentsRoute.PUT("/:id", middleware.RequirePermission(authz.ModelsWrite), controller.UpdateDeployment)
			deploymentsRoute.PUT("/:id/name", middleware.RequirePermission(authz.ModelsWrite), controller.UpdateDeploymentName)
			deploymentsRoute.POST("/:id/extend", middleware.RequirePermission(authz.ModelsWrite), controller.ExtendDeployment)
			deploymentsRoute.DELETE("/:id", middleware.RequirePermission(authz.ModelsWrite), controller.DeleteDeployment)
		}
	}
}
