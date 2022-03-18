// Package when handles custom rules for `when` library
package when

import (
	"regexp"
	"strings"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/olebedev/when/rules"
)

// CasualDate is a rule that matches a date in the format of (tomorrow, yesterday, today, etc.)
func CasualDate(s rules.Strategy, startPeriod time.Time) rules.Rule {
	overwrite := s == rules.Override

	return &rules.F{
		RegExp: regexp.MustCompile(`(?i)(?:\\W|^)(now|today|tonight|last\\s*night|(?:tomorrow|tmr|yesterday)\\s*|tomorrow|tmr|yesterday)(?:\\W|$)`),
		Applier: func(m *rules.Match, c *rules.Context, o *rules.Options, ref time.Time) (bool, error) {
			lower := strings.ToLower(strings.TrimSpace(m.String()))
			hour, minute, second := startPeriod.Clock()

			switch {
			case strings.Contains(lower, "tonight"):
				if c.Hour == nil && c.Minute == nil || overwrite {
					c.Hour = pointer.ToInt(23)
					c.Minute = pointer.ToInt(0)
				}
			case strings.Contains(lower, "today"):
				c.Hour = pointer.ToInt(hour)
				c.Minute = pointer.ToInt(minute)
				c.Second = pointer.ToInt(second)
			case strings.Contains(lower, "tomorrow"), strings.Contains(lower, "tmr"):
				if c.Duration == 0 || overwrite {
					c.Duration += time.Hour * 24
				}
				c.Hour = pointer.ToInt(hour)
				c.Minute = pointer.ToInt(minute)
				c.Second = pointer.ToInt(second)
			case strings.Contains(lower, "yesterday"):
				if c.Duration == 0 || overwrite {
					c.Duration -= time.Hour * 24
				}
				c.Hour = pointer.ToInt(hour)
				c.Minute = pointer.ToInt(minute)
				c.Second = pointer.ToInt(second)
			case strings.Contains(lower, "last night"):
				if (c.Hour == nil && c.Duration == 0) || overwrite {
					c.Hour = pointer.ToInt(23)
					c.Duration -= time.Hour * 24
				}
			}

			return true, nil
		},
	}
}
