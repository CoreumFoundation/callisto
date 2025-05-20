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

// FindEventsByMsgIndex returns all events with the given msg index
func FindEventsByMsgIndex(events sdk.StringEvents, msgIndex int) sdk.StringEvents {
	var res sdk.StringEvents
	for _, event := range events {
		attribute, exist := FindAttributeByKey(event, "msg_index")
		if !exist {
			continue
		}

		if strconv.Itoa(msgIndex) == attribute.Value {
			res = append(res, event)
		}
	}
	return res
}
func FindEventMap(event sdk.StringEvent, requiredAttributes []string, optionalAttributes []string) (map[string]string, error) {
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
