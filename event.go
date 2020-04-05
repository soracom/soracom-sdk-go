package soracom

import (
	"fmt"
	"net/http"
	"strconv"
)

// EventDateTimeConst is the time value that can be specified when execute a rule, enable, or disable
type EventDateTimeConst string

func (x EventDateTimeConst) String() string {
	return string(x)
}

const (
	// EventDateTimeImmediately is immediately
	EventDateTimeImmediately EventDateTimeConst = "IMMEDIATELY"

	// EventDateTimeAfterOneDay is one day (24 hours) later
	EventDateTimeAfterOneDay EventDateTimeConst = "AFTER_ONE_DAY"

	// EventDateTimeBeginningOfNextDay is ...
	EventDateTimeBeginningOfNextDay EventDateTimeConst = "BEGINNING_OF_NEXT_DAY"

	// EventDateTimeBeginningOfNextMonth is ...
	EventDateTimeBeginningOfNextMonth EventDateTimeConst = "BEGINNING_OF_NEXT_MONTH"

	// EventDateTimeNever is ...
	EventDateTimeNever EventDateTimeConst = "NEVER"
)

// EventStatus is status of EventHandler
type EventStatus string

// EventStatus is one of active or inactive
const (
	EventStatusActive   EventStatus = "active"
	EventStatusInactive EventStatus = "inactive"
)

func buildRuleConfig(evrtype EventHandlerRuleType, datetimeConst EventDateTimeConst, prop Properties) RuleConfig {
	prop["inactiveTimeoutDateConst"] = datetimeConst.String()
	return RuleConfig{
		Type:       evrtype,
		Properties: prop,
	}
}

func buildActionConfig(acttype EventHandlerActionType, datetimeConst EventDateTimeConst, prop Properties) ActionConfig {
	prop["executionDateTimeConst"] = datetimeConst.String()
	return ActionConfig{
		Type:       acttype,
		Properties: prop,
	}
}

// RuleDailyTraffic build daily traffic RuleConfig
func RuleDailyTraffic(mib uint64, inactiveDatetime EventDateTimeConst) RuleConfig {
	prop := Properties{
		"limitTotalTrafficMegaByte": strconv.FormatUint(mib, 10),
	}
	return buildRuleConfig(EventHandlerRuleTypeDailyTraffic, inactiveDatetime, prop)
}

// RuleMonthlyTraffic build monthly traffic RuleConfig
func RuleMonthlyTraffic(mib uint64, inactiveDatetime EventDateTimeConst) RuleConfig {
	prop := Properties{
		"limitTotalTrafficMegaByte": strconv.FormatUint(mib, 10),
	}
	return buildRuleConfig(EventHandlerRuleTypeMonthlyTraffic, inactiveDatetime, prop)
}

// ActionActivate build Activate Action
func ActionActivate(executionDateTime EventDateTimeConst) ActionConfig {
	return buildActionConfig(EventHandlerActionTypeActivate, executionDateTime, Properties{})
}

// ActionDeactivate build Deactivate Action
func ActionDeactivate(executionDateTime EventDateTimeConst) ActionConfig {
	return buildActionConfig(EventHandlerActionTypeDeactivate, executionDateTime, Properties{})
}

// ActionWebhookProperty keeps value of webhook property
type ActionWebhookProperty struct {
	URL         string
	Method      string
	ContentType string
	Body        string
}

// Verify is check properties
func (p ActionWebhookProperty) Verify() error {
	switch p.Method {
	case http.MethodPost, http.MethodPut:
	default:
		if p.Body != "" {
			return fmt.Errorf("%s method does not use body field [%s]", p.Method, p.Body)
		}
	}
	return nil
}

func (p ActionWebhookProperty) toProperty() Properties {
	prop := Properties{
		"url":         p.URL,
		"httpMethod":  p.Method,
		"contentType": p.ContentType,
	}
	switch p.Method {
	case http.MethodPost, http.MethodPut:
		prop["body"] = p.Body
	}
	return prop
}

// ActionWebHook build webhook action config
func ActionWebHook(executionDateTime EventDateTimeConst, hookprop ActionWebhookProperty) ActionConfig {
	return buildActionConfig(EventHandlerActionTypeExecuteWebRequest, executionDateTime, hookprop.toProperty())
}

// ActionChangeSpeed build change speed action config
func ActionChangeSpeed(executionDateTime EventDateTimeConst, s SpeedClass) ActionConfig {
	prop := Properties{"speedClass": s.String()}
	return buildActionConfig(EventHandlerActionTypeChangeSpeedClass, executionDateTime, prop)
}

// ActionSendEmailProperty keeps value of email property
type ActionSendEmailProperty struct {
	To      string
	Title   string
	Message string
}

func (p ActionSendEmailProperty) toProperty() Properties {
	return Properties{
		"to":      p.To,
		"title":   p.Title,
		"message": p.Message,
	}
}

// ActionSendEmail buils send email config
func ActionSendEmail(datetimeConst EventDateTimeConst, mailprop ActionSendEmailProperty) ActionConfig {
	return buildActionConfig(EventHandlerActionTypeExecuteWebRequest, datetimeConst, mailprop.toProperty())
}
