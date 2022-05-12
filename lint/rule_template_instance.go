package lint

func NewTemplateInstanceRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-instance-rule",
		description: "Checks that the dashboard has a templated instance.",
		fn: func(d Dashboard) Result {
			template := getTemplateDatasource(d)
			if template == nil || template.Query != Prometheus {
				return ResultSuccess
			}

			if r := checkTemplate(d, "instance"); r != nil {
				return *r
			}

			return ResultSuccess
		},
	}
}
