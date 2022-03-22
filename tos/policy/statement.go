package policy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

const (
	Allow = "Allow"
	Deny  = "Deny"
)

func AllPrincipals() *Principals {
	return &Principals{
		all: &AllPrincipal{},
	}
}

func SomePrincipals(principal string, principalMore ...string) *Principals {
	principals := []string{principal}
	return &Principals{
		multi: map[string][]string{
			"TOS": append(principals, principalMore...),
		},
	}
}

type Principal interface {
	principal() Principal // return self
}

type AllPrincipal struct{}
type MultiPrincipal map[string][]string

func (p *AllPrincipal) principal() Principal  { return p }
func (p MultiPrincipal) principal() Principal { return p }

type Principals struct {
	all   *AllPrincipal
	multi MultiPrincipal
}

// Principal return one of AllPrincipal, MultiPrincipal or nil
func (p *Principals) Principal() Principal {
	if p.all != nil {
		return p.all
	}

	if p.multi != nil {
		return p.multi
	}

	return nil
}

func (p *Principals) MarshalJSON() ([]byte, error) {
	if p.all != nil {
		return []byte(`"*"`), nil
	}

	if p.multi != nil {
		compacts := make(map[string]interface{}, len(p.multi))
		for k, v := range p.multi {
			if size := len(v); size == 1 {
				compacts[k] = v[0]
			} else if size > 1 {
				compacts[k] = v
			}
		}
		return json.Marshal(compacts)
	}

	return make([]byte, 0), nil
}

func (p *Principals) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	if bytes.Equal(data, []byte(`"*"`)) || bytes.Equal(data, []byte(`["*"]`)) {
		p.all = &AllPrincipal{}
		p.multi = nil
		return nil
	}

	compacts := make(map[string]compact)
	if err := json.Unmarshal(data, &compacts); err != nil {
		return err
	}

	multi := make(map[string][]string)
	for k, v := range compacts {
		multi[k] = v
	}
	p.all = nil
	p.multi = multi
	return nil
}

type Action interface {
	action() Action // return self
}

type SingleAction string
type MultiAction []string

func (a SingleAction) action() Action { return a }
func (a MultiAction) action() Action  { return a }

type Actions struct {
	compact
}

// Action return one of SingleAction, MultiAction or nil
func (as *Actions) Action() Action {
	if len(as.compact) == 1 {
		return SingleAction(as.compact[0])
	}

	if len(as.compact) > 1 {
		return MultiAction(as.compact)
	}

	return nil
}

type compact []string

func (as *compact) MarshalJSON() ([]byte, error) {
	if as == nil {
		return nil, nil
	}

	if len(*as) == 1 {
		return json.Marshal((*as)[0])
	}

	if len(*as) > 1 {
		return json.Marshal(*as)
	}

	return make([]byte, 0), nil
}

func (as *compact) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	token, err := json.NewDecoder(bytes.NewReader(data)).Token()
	if err != nil {
		if err == io.EOF { // empty, no action
			return nil
		}
		return err
	}

	switch tv := token.(type) {
	case json.Delim:
		if tv != json.Delim('[') {
			return errors.New("tos: invalid policy compact value syntax")
		}

		var action []string
		if err = json.Unmarshal(data, &action); err != nil {
			return err
		}
		*as = action
	case string:
		*as = []string{tv}
	default:
		return errors.New("tos: invalid policy compact value syntax")
	}

	return nil
}

func AllActions() *Actions {
	return &Actions{compact: []string{"tos:*"}}
}

func SomeActions(action string, actionMore ...string) *Actions {
	actions := []string{action}
	return &Actions{compact: append(actions, actionMore...)}
}

type Resource interface {
	resource() Resource // return self
}

type SingleResource string
type MultiResource []string

func (a SingleResource) resource() Resource { return a }
func (a MultiResource) resource() Resource  { return a }

type Resources struct {
	compact
}

func (as *Resources) Resource() Resource {
	if len(as.compact) == 1 {
		return SingleResource(as.compact[0])
	}

	if len(as.compact) > 1 {
		return MultiResource(as.compact)
	}

	return nil
}

func SomeResource(resource string, resourceMore ...string) *Resources {
	actions := []string{resource}
	return &Resources{compact: append(actions, resourceMore...)}
}

type Condition map[string][]string

type Conditions map[string]Condition

type Statement struct {
	Sid        string      `json:"Sid,omitempty"`
	Effect     string      `json:"Effect,omitempty"`
	Principals *Principals `json:"Principal,omitempty"`
	Actions    *Actions    `json:"Action,omitempty"`
	Resources  *Resources  `json:"Resource,omitempty"`
	Conditions Conditions  `json:"Condition,omitempty"`
}

type Rules struct {
	Version    string      `json:"Version,omitempty"`
	ID         string      `json:"Id,omitempty"`
	Statements []Statement `json:"Statement,omitempty"`
}
