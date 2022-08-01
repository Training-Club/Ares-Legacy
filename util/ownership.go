package util

import (
	"ares/database"
	"ares/model"
)

func CanSeeProfile(viewerId string, viewedId string, params database.QueryParams) bool {
	viewerAccount, err := database.FindDocumentById[model.Account](params, viewerId, viewedId)
	if err != nil {
		return false
	}

	viewedAccount, err := database.FindDocumentById[model.Account](params, viewerId, viewedId)
	if err != nil {
		return false
	}

	if viewerAccount.ID == viewedAccount.ID {
		return true
	}

	// TODO: Implement CanSeeProfile
	return true
}
