package reddit

import (
	"context"
	"encoding/json"
	"fmt"
)

// thingEnvelope is the generic { "kind": "...", "data": {...} } wrapper Reddit
// puts around every object.
type thingEnvelope struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

// listingData is the body of a Listing thing.
type listingData struct {
	Before   string          `json:"before"`
	After    string          `json:"after"`
	Dist     int             `json:"dist"`
	Children []thingEnvelope `json:"children"`
}

// decodeListing unwraps a Listing envelope into its children and cursor.
func decodeListing(body []byte) (listingData, error) {
	var env thingEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return listingData{}, fmt.Errorf("decode envelope: %w", err)
	}
	if env.Kind != "Listing" {
		return listingData{}, fmt.Errorf("expected a Listing, got %q", env.Kind)
	}
	var ld listingData
	if err := json.Unmarshal(env.Data, &ld); err != nil {
		return listingData{}, fmt.Errorf("decode listing: %w", err)
	}
	return ld, nil
}

// reasonEnvelope carries Reddit's refusal reason on a 403/404 body.
type reasonEnvelope struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Error   int    `json:"error"`
}

// classifyErrorBody maps a non-200 body to the right sentinel error. It returns
// nil when the body holds no recognizable reason (the caller falls back to a
// status-based error).
func classifyErrorBody(code int, body []byte) error {
	var re reasonEnvelope
	_ = json.Unmarshal(body, &re)
	switch re.Reason {
	case "private":
		return fmt.Errorf("%w", ErrPrivate)
	case "banned":
		return fmt.Errorf("%w", ErrBanned)
	case "quarantined":
		return fmt.Errorf("%w (quarantined community; pass --cookies for a lent session)", ErrBlocked)
	}
	switch code {
	case 404:
		return ErrNotFound
	case 403:
		return ErrBlocked
	}
	return nil
}

// walkListing pages through a listing-returning URL, calling collect for each
// page's children. collect returns the running total and a stop signal. The
// walk stops when after is empty, a page is empty, collect signals stop, or the
// page cap is reached (pages <= 0 means a single page; limit <= 0 with
// pages <= 0 also means a single page).
func (c *Client) walkListing(ctx context.Context, build func(after string, count int) string, pages, limit int, collect func(children []thingEnvelope) (total int, err error)) error {
	after := ""
	count := 0
	page := 0
	for {
		body, code, err := c.cachedFetch(ctx, build(after, count))
		if err != nil {
			return err
		}
		if code != 200 {
			if e := classifyErrorBody(code, body); e != nil {
				return e
			}
			return fmt.Errorf("unexpected HTTP %d", code)
		}
		ld, err := decodeListing(body)
		if err != nil {
			return err
		}
		if len(ld.Children) == 0 {
			return nil
		}
		total, err := collect(ld.Children)
		if err != nil {
			return err
		}
		count += len(ld.Children)
		page++
		after = ld.After
		if after == "" {
			return nil
		}
		if limit > 0 && total >= limit {
			return nil
		}
		if pages > 0 && page >= pages {
			return nil
		}
		if pages <= 0 && limit <= 0 {
			return nil
		}
	}
}
