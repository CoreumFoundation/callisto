package events

import (
	"fmt"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// JunoAttributeNotFoundError returns an error message indicating that the attribute was not found
// in the event with the given type.
func JunoAttributeNotFoundError(attr string, event abci.Event) string {
	return fmt.Sprintf("no attribute with key %s found inside event with type %s", attr, event.Type)
}

// FindEventByType returns the event with the given type
func FindEventByType(events sdk.StringEvents, eventType string) (sdk.StringEvent, bool) {
	for _, event := range events {
		if event.Type == eventType {
			return event, true
		}
	}
	return sdk.StringEvent{}, false
}

// FindAttributeByKey returns the attribute with the given key
func FindAttributeByKey(event sdk.StringEvent, key string) (sdk.Attribute, bool) {
	for _, attribute := range event.Attributes {
		if attribute.Key == key {
			return attribute, true
		}
	}
	return sdk.Attribute{}, false
}

// FindEventsByMsgIndex returns all events with the given msg index.
// It first searches for events with msg_index attributes (cosmos-sdk v0.50.x+),
// then falls back to legacy logs-based search for older versions.
func FindEventsByMsgIndex(events sdk.StringEvents, logs sdk.ABCIMessageLogs, msgIndex int) sdk.StringEvents {
	const msgIndexKey = "msg_index"
	msgIndexStr := strconv.Itoa(msgIndex)

	// Search for events with msg_index attributes (cosmos-sdk v0.50.x+)
	var res sdk.StringEvents
	for _, event := range events {
		attribute, exist := FindAttributeByKey(event, msgIndexKey)
		if !exist {
			continue
		}

		if attribute.Value == msgIndexStr {
			res = append(res, event)
		}
	}

	// Early return if we found events
	if len(res) > 0 {
		return res
	}

	// Fallback for events generated before upgrading to cosmos-sdk v0.50.x
	// where msg_index was in logs instead of events
	for _, log := range logs {
		if int(log.MsgIndex) == msgIndex {
			res = log.Events
			continue
		}
	}

	return res
}

// BuildAttributesMap returns a map of attributes from the event
// with the given required and optional attributes.
// It returns an error if any of the required attributes are not found.
func BuildAttributesMap(event sdk.StringEvent, requiredAttributes []string, optionalAttributes []string) (map[string]string, error) {
	result := make(map[string]string)

	// Check for required attributes
	for _, key := range requiredAttributes {
		attr, ok := FindAttributeByKey(event, key)
		if !ok {
			return nil, fmt.Errorf("required attribute %s not found", key)
		}
		result[key] = attr.Value
	}

	// Check for optional attributes
	for _, key := range optionalAttributes {
		attr, ok := FindAttributeByKey(event, key)
		if ok {
			result[key] = attr.Value
		} else {
			result[key] = "" // Set to empty string if not found
		}
	}

	return result, nil
}
