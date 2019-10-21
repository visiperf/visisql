package visisql

type Creation struct {
	Set        map[string]interface{} `json:"set"`
}

/*
**	Usage example :
**
**	creations := make(map[string]*visisql.Creation)
**	creations[pageRef] = visisql.NewCreation(
**		map[string]interface{}{
**			"userToken":		userToken,
**			"accessToken":		accessToken,
**			"pageRef":			pageRef,
**		},
**	)
*/

func NewCreation(set map[string]interface{}) *Creation {
	return &Creation{Set: set}
}
