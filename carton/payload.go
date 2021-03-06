/*
** Copyright [2013-2016] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */
package carton

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/gulp/db"
)

type Payload struct {
	Id        string `json:"id"`
	CatId     string `json:"cat_id"`   // assemblies_id
	Action    string `json:"action"`   // start, stop ...
	Category  string `json:"category"` // state, control, policy
	CreatedAt string `json:"created_at"`
}

type PayloadConvertor interface {
	Convert(p *Payload) (*Requests, error)
}

func NewPayload(b []byte) (*Payload, error) {
	p := &Payload{}
	err := json.Unmarshal(b, &p)
	if err != nil {
		log.Errorf("Failed to parse the payload message: %s.", err)
		return nil, err
	}
	return p, err
}

func (p *Payload) AsBytes(id string,
	catid string, action string,
	category string,
	createdat string) ([]byte, error) {
	p.Id = id
	p.CatId = catid
	p.Action = action
	p.Category = category
	p.CreatedAt = createdat

	return json.Marshal(p)
}

//fetch the request json from riak and parse the json to struct
func (p *Payload) Convert() (*Requests, error) {
	log.Infof("get requests %s", p.Id)
	if p.CatId != "" {
		r := &Requests{
			Id:        p.Id,
			CatId:     p.CatId,
			Action:    p.Action,
			Category:  p.Category,
			CreatedAt: p.CreatedAt,
		}

		log.Debugf("Requests %v", r)
		return r, nil
	} else {
		r := &Requests{}
		if err := db.Fetch("requests", p.Id, r); err != nil {
			return nil, err
		}
		log.Debugf("Requests %v", r)
		return r, nil
	}
}
