package routes

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/devmarvs/bblog/models"
	"github.com/gin-gonic/gin"
)

func createAppVersion(context *gin.Context) {
	var version models.AppVersion
	if err := context.ShouldBindJSON(&version); err != nil {
		respondBadRequest(context, ErrRequestDataParse)
		return
	}

	sanitizeAppVersion(&version)
	if status, message := validateAppVersion(&version); status != 0 {
		respondWithError(context, status, message)
		return
	}

	if err := version.Save(); err != nil {
		context.Error(err)
		respondInternalServerError(context, ErrSaveAppVersion)
		return
	}

	respondWithDataAndMessage(context, http.StatusCreated, version, "Version saved successfully")
}

func getLatestAppVersion(context *gin.Context) {
	version, err := models.GetLatestAppVersion()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithData(context, http.StatusOK, nil)
			return
		}

		context.Error(err)
		respondInternalServerErrorWithData(context, ErrFetchAppVersion)
		return
	}

	respondWithData(context, http.StatusOK, version)
}
