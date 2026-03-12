package components

import (
	"github.com/CheeziCrew/curd"
)

type ResultModel = curd.ResultModel

func NewResultModel(title string, tasks []curd.RepoTask) curd.ResultModel {
	return curd.NewResultModel(title, tasks, curd.SwissgitPalette)
}
