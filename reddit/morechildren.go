package reddit

import (
	"context"
	"encoding/json"
	"fmt"
)

// moreData is the body of a "more" stub: the ids of children not yet loaded.
type moreData struct {
	Count    int      `json:"count"`
	Name     string   `json:"name"`
	ParentID string   `json:"parent_id"`
	Depth    int      `json:"depth"`
	Children []string `json:"children"`
}

// flattenTree walks a comment listing depth-first into flat Comment records. A
// "more" stub contributes its count to its parent's MoreCount and, when ids are
// present, to the returned list of pending child ids keyed by depth.
func flattenTree(children []thingEnvelope, out *[]Comment, pending *[]moreData) {
	for _, ch := range children {
		switch ch.Kind {
		case string(kindComment):
			cm, replies, err := parseComment(ch.Data)
			if err != nil {
				continue
			}
			// Look one level down to count direct replies and any more stub.
			rc, mc := summarizeReplies(replies)
			cm.ReplyCount = rc
			cm.MoreCount = mc
			*out = append(*out, cm)
			if len(replies) > 0 {
				if ld, err := decodeListing(replies); err == nil {
					flattenTree(ld.Children, out, pending)
				}
			}
		case "more":
			var md moreData
			if err := json.Unmarshal(ch.Data, &md); err != nil {
				continue
			}
			if len(md.Children) > 0 {
				*pending = append(*pending, md)
			}
		}
	}
}

// summarizeReplies counts the direct t1 replies and the hidden children behind a
// more stub at one level of a replies listing.
func summarizeReplies(replies json.RawMessage) (replyCount, moreCount int) {
	if len(replies) == 0 {
		return 0, 0
	}
	ld, err := decodeListing(replies)
	if err != nil {
		return 0, 0
	}
	for _, ch := range ld.Children {
		switch ch.Kind {
		case string(kindComment):
			replyCount++
		case "more":
			var md moreData
			if json.Unmarshal(ch.Data, &md) == nil {
				moreCount += md.Count
			}
		}
	}
	return replyCount, moreCount
}

// expandMore drains the pending "more" stubs by calling the morechildren
// endpoint in batches, appending the comments it returns and recursing into any
// further stubs, bounded by rounds so it always terminates.
func (c *Client) expandMore(ctx context.Context, linkFullname, sort string, pending []moreData, out *[]Comment, rounds int) error {
	for round := 0; round < rounds && len(pending) > 0; round++ {
		var ids []string
		for _, md := range pending {
			ids = append(ids, md.Children...)
		}
		pending = nil

		for len(ids) > 0 {
			batch := ids
			if len(batch) > MaxPageLimit {
				batch = batch[:MaxPageLimit]
				ids = ids[MaxPageLimit:]
			} else {
				ids = nil
			}
			body, code, err := c.Fetch(ctx, MoreChildrenURL(linkFullname, sort, batch))
			if err != nil {
				return err
			}
			if code != 200 {
				return fmt.Errorf("morechildren HTTP %d", code)
			}
			things, err := decodeMoreChildren(body)
			if err != nil {
				return err
			}
			for _, ch := range things {
				switch ch.Kind {
				case string(kindComment):
					if cm, replies, err := parseComment(ch.Data); err == nil {
						rc, mc := summarizeReplies(replies)
						cm.ReplyCount = rc
						cm.MoreCount = mc
						*out = append(*out, cm)
					}
				case "more":
					var md moreData
					if json.Unmarshal(ch.Data, &md) == nil && len(md.Children) > 0 {
						pending = append(pending, md)
					}
				}
			}
		}
	}
	return nil
}

// moreChildrenResponse is the envelope morechildren returns.
type moreChildrenResponse struct {
	JSON struct {
		Data struct {
			Things []thingEnvelope `json:"things"`
		} `json:"data"`
	} `json:"json"`
}

func decodeMoreChildren(body []byte) ([]thingEnvelope, error) {
	var r moreChildrenResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("decode morechildren: %w", err)
	}
	return r.JSON.Data.Things, nil
}
