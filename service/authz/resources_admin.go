package authz

const (
	ResourceModels        = "models"
	ResourceUsers         = "users"
	ResourceRedemption    = "redemption"
	ResourceSubscriptions = "subscriptions"
	ResourceLogs          = "logs"
	ResourceDashboard     = "dashboard"
)

var (
	ModelsRead  = Permission{Resource: ResourceModels, Action: ActionRead}
	ModelsWrite = Permission{Resource: ResourceModels, Action: ActionWrite}

	UsersRead  = Permission{Resource: ResourceUsers, Action: ActionRead}
	UsersWrite = Permission{Resource: ResourceUsers, Action: ActionWrite}

	RedemptionRead  = Permission{Resource: ResourceRedemption, Action: ActionRead}
	RedemptionWrite = Permission{Resource: ResourceRedemption, Action: ActionWrite}

	SubscriptionsRead  = Permission{Resource: ResourceSubscriptions, Action: ActionRead}
	SubscriptionsWrite = Permission{Resource: ResourceSubscriptions, Action: ActionWrite}

	LogsRead = Permission{Resource: ResourceLogs, Action: ActionRead}

	DashboardRead = Permission{Resource: ResourceDashboard, Action: ActionRead}
)

func init() {
	RegisterResource(ResourceDefinition{
		Resource: ResourceModels,
		LabelKey: "Model Management",
		Actions: []ActionDefinition{
			{
				Action:         ActionRead,
				LabelKey:       "Read models",
				DescriptionKey: "View model catalog, pricing, vendors, and deployments.",
				DefaultRoles:   []string{BuiltInRoleAdmin},
			},
			{
				Action:         ActionWrite,
				LabelKey:       "Edit models",
				DescriptionKey: "Create, update, and delete models, vendors, and deployments.",
			},
		},
	})
	RegisterResource(ResourceDefinition{
		Resource: ResourceUsers,
		LabelKey: "User Management",
		Actions: []ActionDefinition{
			{
				Action:         ActionRead,
				LabelKey:       "Read users",
				DescriptionKey: "View user lists, quotas, bindings, and 2FA status.",
				DefaultRoles:   []string{BuiltInRoleAdmin},
			},
			{
				Action:         ActionWrite,
				LabelKey:       "Edit users",
				DescriptionKey: "Create, update, disable, or delete users and change quotas.",
			},
		},
	})
	RegisterResource(ResourceDefinition{
		Resource: ResourceRedemption,
		LabelKey: "Redemption Code Management",
		Actions: []ActionDefinition{
			{
				Action:         ActionRead,
				LabelKey:       "Read redemption codes",
				DescriptionKey: "View redemption codes and their status.",
				DefaultRoles:   []string{BuiltInRoleAdmin},
			},
			{
				Action:         ActionWrite,
				LabelKey:       "Edit redemption codes",
				DescriptionKey: "Create, update, or delete redemption codes.",
			},
		},
	})
	RegisterResource(ResourceDefinition{
		Resource: ResourceSubscriptions,
		LabelKey: "Subscription Management",
		Actions: []ActionDefinition{
			{
				Action:         ActionRead,
				LabelKey:       "Read subscriptions",
				DescriptionKey: "View subscription plans and user subscriptions.",
				DefaultRoles:   []string{BuiltInRoleAdmin},
			},
			{
				Action:         ActionWrite,
				LabelKey:       "Edit subscriptions",
				DescriptionKey: "Create, update, bind, or invalidate subscription plans.",
			},
		},
	})
	RegisterResource(ResourceDefinition{
		Resource: ResourceLogs,
		LabelKey: "Usage Log Access",
		Actions: []ActionDefinition{
			{
				Action:         ActionRead,
				LabelKey:       "Read all usage logs",
				DescriptionKey: "View usage, drawing, and task logs for every user.",
				DefaultRoles:   []string{BuiltInRoleAdmin},
			},
		},
	})
	RegisterResource(ResourceDefinition{
		Resource: ResourceDashboard,
		LabelKey: "Site Analytics",
		Actions: []ActionDefinition{
			{
				Action:         ActionRead,
				LabelKey:       "Read site analytics",
				DescriptionKey: "View site-wide dashboard metrics and user analytics.",
				DefaultRoles:   []string{BuiltInRoleAdmin},
			},
		},
	})
}
