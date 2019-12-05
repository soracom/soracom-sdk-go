package soracom

import "strconv"

type EventDateTimeConst string

func (x EventDateTimeConst) String() string {
	return string(x)
}

const (
	IMMEDIATELY             = "IMMEDIATELY"
	AFTER_ONE_DAY           = "AFTER_ONE_DAY"
	BEGINNING_OF_NEXT_DAY   = "BEGINNING_OF_NEXT_DAY"
	BEGINNING_OF_NEXT_MONTH = "BEGINNING_OF_NEXT_MONTH"
	NEVER                   = "NEVER"
)

type EventStatus string

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

func RuleDailyTraffic(mib uint64, datetimeConst EventDateTimeConst) RuleConfig {
	prop := Properties{
		"limitTotalTrafficMegaByte": strconv.FormatUint(mib, 10),
	}
	return buildRuleConfig(EventHandlerRuleTypeDailyTraffic, datetimeConst, prop)
}

func RuleMonthlyTraffic(mib uint64, datetimeConst EventDateTimeConst) RuleConfig {
	prop := Properties{
		"limitTotalTrafficMegaByte": strconv.FormatUint(mib, 10),
	}
	return buildRuleConfig(EventHandlerRuleTypeMonthlyTraffic, datetimeConst, prop)
}

func ActionActivate(datetimeConst EventDateTimeConst) ActionConfig {
	return buildActionConfig(EventHandlerActionTypeActivate, datetimeConst, Properties{})
}

func ActionDectivate(datetimeConst EventDateTimeConst) ActionConfig {
	return buildActionConfig(EventHandlerActionTypeDeactivate, datetimeConst, Properties{})
}

type ActionWebhookProperty struct {
	URL         string
	Method      string
	ContentType string
	Body        string
}

func (p ActionWebhookProperty) toProperty() Properties {
	return Properties{
		"url":         p.URL,
		"httpMethod":  p.Method,
		"contentType": p.ContentType,
		"body":        p.Body,
	}
}

func ActionWebHook(datetimeConst EventDateTimeConst, hookprop ActionWebhookProperty) ActionConfig {
	return buildActionConfig(EventHandlerActionTypeExecuteWebRequest, datetimeConst, hookprop.toProperty())
}

func ActionChangeSpeed(datetimeConst EventDateTimeConst, s SpeedClass) ActionConfig {
	prop := Properties{"speedClass": s.String()}
	return buildActionConfig(EventHandlerActionTypeChangeSpeedClass, datetimeConst, prop)
}

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
func ActionSendEmail(datetimeConst EventDateTimeConst, mailprop ActionSendEmailProperty) ActionConfig {
	return buildActionConfig(EventHandlerActionTypeExecuteWebRequest, datetimeConst, mailprop.toProperty())
}
