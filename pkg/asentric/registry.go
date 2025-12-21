package asentric

type RuleRegistry struct {
	rules map[string]*Rule
}

func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules: make(map[string]*Rule),
	}
}

func (r *RuleRegistry) Register(rule *Rule) error {
	if rule == nil {
		return ErrInvalidRule
	}

	if _, exists := r.rules[rule.ID]; exists {
		return ErrDuplicateRule
	}

	r.rules[rule.ID] = rule
	return nil
}

func (r *RuleRegistry) List() []*Rule {
	out := make([]*Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		out = append(out, rule)
	}
	return out
}
