package dispatcher

import (
	"context"

	"github.com/Z00mZE/fts/domain/entity"
	"github.com/Z00mZE/fts/domain/service"
	"github.com/Z00mZE/fts/gen"
)

type index interface {
	Search(string) []entity.Document
}
type Dispatcher struct {
	log   service.Logger
	index index
}

func NewDispatcher(index index, log service.Logger) *Dispatcher {
	return &Dispatcher{
		index: index,
		log:   log,
	}
}

func (d *Dispatcher) Search(ctx context.Context, request *gen.Request) (*gen.Response, error) {
	response := new(gen.Response)
	{
		result := d.index.Search(request.Query)
		if resultLength := len(result); resultLength != 0 {
			response.Result = make([]*gen.Response_Document, 0, resultLength)
			for i := 0; i < resultLength; i++ {
				response.Result = append(
					response.Result,
					&gen.Response_Document{
						Id:          result[i].ID,
						Description: result[i].Text,
					},
				)
			}
		}
	}
	return response, nil
}
